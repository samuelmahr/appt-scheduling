package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/samuelmahr/appt-scheduling/internal/configuration"
	"github.com/samuelmahr/appt-scheduling/internal/models"
	"github.com/samuelmahr/appt-scheduling/internal/repo"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type errorTyper interface {
	ErrorType() string
}

type V1AppointmentsController struct {
	config *configuration.AppConfig
	repo   repo.AppointmentsRepoType
}

func NewV1AppointmentsController(c *configuration.AppConfig, uRepo repo.AppointmentsRepoType) V1AppointmentsController {
	return V1AppointmentsController{
		config: c,
		repo:   uRepo,
	}
}

func (a *V1AppointmentsController) RegisterRoutes(v1 *mux.Router) {
	v1.Path("/appointments/available").Name("GetAvailableAppointments").Handler(http.HandlerFunc(a.ListAvailableAppointments)).Methods(http.MethodGet)
	v1.Path("/appointments/scheduled").Name("GetScheduledAppointments").Handler(http.HandlerFunc(a.ListScheduledAppointments)).Methods(http.MethodGet)
	v1.Path("/appointments").Name("CreateAppointments").Handler(http.HandlerFunc(a.CreateAppointment)).Methods(http.MethodPost)
}

func (a *V1AppointmentsController) CreateAppointment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	newAppointment := models.AppointmentCreateRequest{}

	err := json.NewDecoder(r.Body).Decode(&newAppointment)
	if err != nil {
		respondError(ctx, w, http.StatusBadRequest, "bad request payload", err)
		return
	}

	err = validateRequest(newAppointment)
	if err != nil {
		respondError(ctx, w, http.StatusBadRequest, "bad request payload, check required fields", err)
		return
	}

	appointment, err := a.repo.CreateAppointment(ctx, newAppointment)
	if err != nil {
		respondError(ctx, w, http.StatusInternalServerError, err.Error(), err)
		return
	}

	respondModel(ctx, w, http.StatusCreated, appointment)
	return
}

func validateRequest(appointment models.AppointmentCreateRequest) error {
	// validate user ID is not 0
	if appointment.UserID == 0 {
		return errors.New("invalid user Id")
	}

	if appointment.TrainerID == 0 {
		return errors.New("invalid user Id")
	}

	err := validateTimeSlot(appointment.StartsAt, appointment.EndsAt)
	if err != nil {
		return err
	}

	return nil
}

func validateTimeSlot(startsAt time.Time, endsAt time.Time) error {
	// Calling LoadLocation
	// method with its parameter
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return errors.New("could not load timezone location")
	}

	if endsAt.Sub(startsAt).Minutes() != 30 {
		return errors.New("invalid time slot, must be 30 minutes")
	}

	if isValidSlot(startsAt) && startsWithinBusinessHours(startsAt, loc) {
		return errors.New("invalid end date/time, must end at 00 or 30 minutes")
	}

	return nil
}

func startsWithinBusinessHours(startsAt time.Time, loc *time.Location) bool {
	// earliest start is 8am, latest start is 4pm and latest start is 4:30pm based on previous conditions
	pacificStart := startsAt.In(loc)
	weekday := pacificStart.Weekday()
	hour := pacificStart.Hour()
	log.Println("checking business hours")

	log.Printf("weekday: %s\n", weekday)
	log.Printf("hour: %d\n", hour)
	return hour >= 8 && hour <= 16 && weekday != time.Saturday && weekday != time.Sunday
}

func isValidSlot(time time.Time) bool {
	return time.Minute() != 0 && time.Minute() != 30 && time.Second() != 0 && time.Nanosecond() != 0
}

func (a *V1AppointmentsController) ListScheduledAppointments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queryParams := r.URL.Query()

	trainerID, err := getTrainerID(queryParams)
	if err != nil {
		respondError(ctx, w, http.StatusBadRequest, "invalid trainer ID", err)
		return
	}

	startsAt, endsAt, err := getTimeRange(queryParams)
	if err != nil {
		respondError(ctx, w, http.StatusBadRequest, "invalid time range values", err)
		return
	}

	appointments, err := a.repo.GetScheduledAppointments(ctx, trainerID, startsAt, endsAt)
	if err != nil {
		respondError(ctx, w, http.StatusInternalServerError, "something bad happened", err)
		return
	}

	respondModel(ctx, w, http.StatusOK, appointments)
	return
}

func (a *V1AppointmentsController) ListAvailableAppointments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	queryParams := r.URL.Query()

	trainerID, err := getTrainerID(queryParams)
	if err != nil {
		respondError(ctx, w, http.StatusBadRequest, "invalid trainer ID", err)
		return
	}

	startsAt, endsAt, err := getTimeRange(queryParams)
	if err != nil {
		respondError(ctx, w, http.StatusBadRequest, "invalid time range values", err)
		return
	}

	if startsAt.IsZero() || endsAt.IsZero() {
		respondError(ctx, w, http.StatusBadRequest, "invalid time range is required", err)
		return
	}

	timeSlots, err := a.repo.GetScheduledAppointmentsAsTimeSlots(ctx, trainerID, startsAt, endsAt)
	if err != nil {
		respondError(ctx, w, http.StatusInternalServerError, "something bad happened", err)
		return
	}

	availableAppointments, err := buildAvailableAppointments(startsAt, endsAt, trainerID, timeSlots)
	if err != nil {
		respondError(ctx, w, http.StatusInternalServerError, "something went wrong", err)
		return
	}

	respondModel(ctx, w, http.StatusOK, availableAppointments)
	return
}

func buildAvailableAppointments(startsAt time.Time, endsAt time.Time, trainerID int64, timeSlots map[int64]int64) ([]models.Appointment, error) {
	// Calling LoadLocation
	// method with its parameter
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		return []models.Appointment{}, errors.New("could not load timezone location")
	}

	appointments := make([]models.Appointment, 0)
	currentTimeSlot := startsAt

	log.Printf("scheduled timeslots unix time: %#v\n", timeSlots)
	for {
		if currentTimeSlot.Equal(endsAt) {
			break
		}

		log.Println("checking timeslot")

		log.Printf("current unix slot: %d\n", currentTimeSlot.Unix())
		// check if this time is a scheduled time
		_, ok := timeSlots[currentTimeSlot.Unix()]

		// if within business hours and unscheduled
		if startsWithinBusinessHours(currentTimeSlot, loc) && !ok {
			appointments = append(appointments,
				models.Appointment{
					TrainerID: trainerID,
					StartsAt:  currentTimeSlot,
					EndsAt:    currentTimeSlot.Add(30 * time.Minute),
				})
		}

		currentTimeSlot = currentTimeSlot.Add(30 * time.Minute)
	}

	return appointments, nil

}

func getTrainerID(queryParams url.Values) (int64, error) {
	trainerIDStr := queryParams.Get("trainer_id")
	if trainerIDStr == "" {
		return 0, nil
	}

	var trainerID int64
	trainerID, err := strconv.ParseInt(trainerIDStr, 10, 64)
	return trainerID, err
}

func getTimeRange(queryParams url.Values) (time.Time, time.Time, error) {
	startsAtStr := queryParams.Get("starts_at")
	var startsAt time.Time
	if startsAtStr != "" {
		// validates expected time format, will return error if not expected format
		t, err := time.Parse(time.RFC3339, startsAtStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		startsAt = t
	}

	endsAtStr := queryParams.Get("ends_at")
	var endsAt time.Time
	if startsAtStr != "" {
		// validates expected time format, will return error if not expected format
		t, err := time.Parse(time.RFC3339, endsAtStr)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}

		endsAt = t
	}

	return startsAt, endsAt, nil
}
