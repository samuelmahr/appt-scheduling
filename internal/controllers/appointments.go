package controllers

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/samuelmahr/appt-scheduling/internal/configuration"
	"github.com/samuelmahr/appt-scheduling/internal/models"
	"github.com/samuelmahr/appt-scheduling/internal/repo"
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
	v1.Path("/appointments").Name("GetAppointments").Handler(http.HandlerFunc(a.ListAppointments)).Methods(http.MethodGet)
	v1.Path("/appointments").Name("GetAppointments").Handler(http.HandlerFunc(a.CreateAppointment)).Methods(http.MethodPost)
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
	if endsAt.Sub(startsAt).Minutes() != 30 {
		return errors.New("invalid time slot, must be 30 minutes")
	}

	if endsAt.Minute() != 0 && endsAt.Second() != 0 && endsAt.Nanosecond() != 0 {
		return errors.New("invalid end date/time, must end at 00 or 30 minutes")
	}

	// this is really just a safety net. the first condition and the second should make this unreachable
	if endsAt.Minute() != 0 && endsAt.Second() != 0 && endsAt.Nanosecond() != 0 {
		return errors.New("invalid end date/time, must end at 00 or 30 minutes")
	}

	return nil
}

func (a *V1AppointmentsController) ListAppointments(w http.ResponseWriter, r *http.Request) {
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

	appointments, err := a.repo.GetAppointments(ctx, trainerID, startsAt, endsAt)
	if err != nil {
		respondError(ctx, w, http.StatusInternalServerError, "lmfao something happened", err)
		return
	}

	respondModel(ctx, w, http.StatusOK, appointments)
	return
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
