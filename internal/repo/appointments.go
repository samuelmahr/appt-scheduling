package repo

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/samuelmahr/appt-scheduling/internal/models"
	"time"
)

type AppointmentsRepository interface {
	CreateAppointment(ctx context.Context, newUser models.AppointmentCreateRequest) (models.Appointment, error)
	GetScheduledAppointments(ctx context.Context, tID int64, startsAt time.Time, endsAt time.Time) ([]models.Appointment, error)
	GetScheduledAppointmentsAsTimeSlots(ctx context.Context, tID int64, startsAt time.Time, endsAt time.Time) (map[int64]int64, error)
}

type AppointmentsRepoType struct {
	db *sqlx.DB
}

func NewAppointmentsRepository(db *sqlx.DB) AppointmentsRepoType {
	return AppointmentsRepoType{
		db: db,
	}
}

const createAppointmentQuery = `
insert into scheduling.appointments(trainer_id, user_id, starts_at, ends_at)
VALUES ($1, $2, $3, $4)
returning id, trainer_id, user_id, starts_at, ends_at, created_at, updated_at, canceled_at
`

func (ar *AppointmentsRepoType) CreateAppointment(ctx context.Context, newAppt models.AppointmentCreateRequest) (models.Appointment, error) {
	var a models.Appointment
	err := ar.db.QueryRowx(createAppointmentQuery, newAppt.TrainerID, newAppt.UserID, newAppt.StartsAt, newAppt.EndsAt).StructScan(&a)

	if err != nil {
		return models.Appointment{}, errors.Wrap(err, "error creating appointment")
	}

	return a, nil
}

func (ar *AppointmentsRepoType) GetScheduledAppointments(ctx context.Context, trainerID int64, startsAt time.Time, endsAt time.Time) ([]models.Appointment, error) {
	sql, args, err := buildGetScheduledApptsQuery(trainerID, startsAt, endsAt)
	if err != nil {
		return []models.Appointment{}, errors.Wrap(err, "error getting appointments")
	}

	rows, err := ar.db.Queryx(sql, args...)
	if err != nil {
		return []models.Appointment{}, errors.Wrap(err, "error getting appointments")
	}

	appts := make([]models.Appointment, 0)
	for rows.Next() {
		var a models.Appointment
		if err := rows.StructScan(&a); err != nil {
			return []models.Appointment{}, errors.Wrap(err, "error getting appointments")
		}

		appts = append(appts, a)
	}

	return appts, nil
}

func (ar *AppointmentsRepoType) GetScheduledAppointmentsAsTimeSlots(ctx context.Context, trainerID int64, startsAt time.Time, endsAt time.Time) (map[int64]int64, error) {
	sql, args, err := buildGetScheduledApptsQuery(trainerID, startsAt, endsAt)
	if err != nil {
		return map[int64]int64{}, errors.Wrap(err, "error getting appointments")
	}

	rows, err := ar.db.Queryx(sql, args...)
	if err != nil {
		return map[int64]int64{}, errors.Wrap(err, "error getting appointments")
	}

	startToEndUnix := make(map[int64]int64)
	for rows.Next() {
		var a models.Appointment
		if err := rows.StructScan(&a); err != nil {
			return map[int64]int64{}, errors.Wrap(err, "error getting appointments")
		}

		startToEndUnix[a.StartsAt.Unix()] = a.EndsAt.Unix()
	}

	return startToEndUnix, nil
}

func buildGetScheduledApptsQuery(trainerID int64, startsAt time.Time, endsAt time.Time) (string, []interface{}, error) {
	query := sq.Select("id", "trainer_id", "user_id", "starts_at", "ends_at", "created_at", "updated_at", "canceled_at").From("scheduling.appointments")
	if trainerID != 0 {
		// find for trainer ID
		query = query.Where(sq.Eq{"trainer_id": trainerID})
	}

	query = query.PlaceholderFormat(sq.Dollar)
	if !startsAt.IsZero() && !endsAt.IsZero() {
		// check between times
		query = query.Where(sq.And{sq.GtOrEq{"starts_at": startsAt}, sq.LtOrEq{"ends_at": endsAt}})
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return "", []interface{}{}, errors.Wrap(err, "error getting appointments")
	}
	return sql, args, err
}
