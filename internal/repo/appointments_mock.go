package repo

import (
	"context"
	"github.com/samuelmahr/appt-scheduling/internal/models"
	"time"
)

// MockAppointments is an implementation of AppointmentsRepository to set values to use as a mock when testing
type MockAppointments struct {
	CreateAppointmentsResponse models.Appointment
	CreateAppointmentsErr      error

	GetScheduledAppointmentsResponse []models.Appointment
	GetScheduledAppointmentsErr      error

	GetScheduledAppointmentsAsTimeSlotsResponse map[int64]int64
	GetScheduledAppointmentsAsTimeSlotsErr      error
}

func (m *MockAppointments) CreateAppointment(ctx context.Context, newUser models.AppointmentCreateRequest) (models.Appointment, error) {
	return m.CreateAppointmentsResponse, m.CreateAppointmentsErr
}

func (m *MockAppointments) GetScheduledAppointments(ctx context.Context, tID int64, startsAt time.Time, endsAt time.Time) ([]models.Appointment, error) {
	return m.GetScheduledAppointmentsResponse, m.GetScheduledAppointmentsErr
}

func (m *MockAppointments) GetScheduledAppointmentsAsTimeSlots(ctx context.Context, tID int64, startsAt time.Time, endsAt time.Time) (map[int64]int64, error) {
	return m.GetScheduledAppointmentsAsTimeSlotsResponse, m.GetScheduledAppointmentsAsTimeSlotsErr
}
