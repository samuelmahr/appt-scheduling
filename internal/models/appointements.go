package models

import "time"

// Appointment models database table
type Appointment struct {
	ID         int64      `json:"id" db:"id"`
	TrainerID  int64      `json:"trainer_id" db:"trainer_id"`
	UserID     int64      `json:"user_id" db:"user_id"`
	StartsAt   time.Time  `json:"starts_at" db:"starts_at"`
	EndsAt     time.Time  `json:"ends_at" db:"ends_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
	CanceledAt *time.Time `json:"canceled_at" db:"canceled_at"`
}

// AppointmentCreateRequest models API Request Payload to create an appointment
type AppointmentCreateRequest struct {
	ID        int64     `json:"id" db:"id"`
	TrainerID int64     `json:"trainer_id" db:"trainer_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	StartsAt  time.Time `json:"starts_at" db:"starts_at"`
	EndsAt    time.Time `json:"ends_at" db:"ends_at"`
}
