package repo

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/samuelmahr/appt-scheduling/internal/models"
	log "github.com/sirupsen/logrus"
	"time"
)

type UsersRepository interface {
	CreateAppointment(ctx context.Context, newUser models.AppointmentCreateRequest) (models.Appointment, error)
	GetAppointments(ctx context.Context, tID int64, startsAt time.Time, endsAt time.Time) ([]models.Appointment, error)
}

type AppointmentsRepoType struct {
	db *sqlx.DB
}

func NewUsersRepository(db *sqlx.DB) AppointmentsRepoType {
	return AppointmentsRepoType{
		db: db,
	}
}

const createAppointmentQuery = `
insert into scheduling.appointments(trainer_id, user_id, starts_at, ends_at)
VALUES ($1, $2, $3, $4)
returning id trainer_id, user_id, starts_at, ends_at, created_at, updated_at, canceled_at
`

func (ar *AppointmentsRepoType) CreateAppointment(ctx context.Context, newAppt models.AppointmentCreateRequest) (models.Appointment, error) {
	var a models.Appointment
	err := ar.db.QueryRowx(createAppointmentQuery, newAppt.TrainerID, newAppt.UserID, newAppt.StartsAt, newAppt.EndsAt).StructScan(&a)

	if err != nil {
		return models.Appointment{}, errors.Wrap(err, "error creating user")
	}

	return a, nil
}

func (ar *AppointmentsRepoType) GetAppointments(ctx context.Context, trainerID int64, startsAt time.Time, endsAt time.Time) ([]models.Appointment, error) {
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
	log.Info(sql)
	log.Info(args)
	if err != nil {
		return []models.Appointment{}, errors.Wrap(err, "error getting appointments")
	}

	appts := make([]models.Appointment, 0)

	rows, err := ar.db.Queryx(sql, args...)
	if err != nil {
		return []models.Appointment{}, errors.Wrap(err, "error getting appointments")
	}

	for rows.Next() {
		var a models.Appointment
		if err := rows.StructScan(&a); err != nil {
			return []models.Appointment{}, errors.Wrap(err, "error getting appointments")
		}

		appts = append(appts, a)
	}

	return appts, nil
}
