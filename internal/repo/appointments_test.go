package repo

import (
	"context"
	"fmt"
	"github.com/samuelmahr/appt-scheduling/internal/models"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	SetupTestDB()
	code := m.Run()
	os.Exit(code)
}

func TestAppointmentRepository_CreateAppointment(t *testing.T) {
	type appt struct {
		createRequest models.AppointmentCreateRequest
		wantAssert    bool
	}
	type args struct {
		appointments []appt
	}

	tests := []struct {
		name    string
		args    args
		want    models.Appointment
		wantErr bool
	}{
		{
			name: "happy path",
			want: models.Appointment{
				ID:        1,
				TrainerID: 1,
				UserID:    1,
				StartsAt:  time.Date(2022, 03, 17, 12, 0, 0, 0, time.UTC),
				EndsAt:    time.Date(2022, 03, 17, 12, 30, 0, 0, time.UTC),
			},
			args: args{
				appointments: []appt{
					{
						createRequest: models.AppointmentCreateRequest{
							TrainerID: 1,
							UserID:    1,
							StartsAt:  time.Date(2022, 03, 17, 12, 0, 0, 0, time.UTC),
							EndsAt:    time.Date(2022, 03, 17, 12, 30, 0, 0, time.UTC),
						},
						wantAssert: true,
					},
				},
			},
		},
		{
			name: "error time slot already exists",
			want: models.Appointment{
				ID:        1,
				TrainerID: 1,
				UserID:    1,
				StartsAt:  time.Date(2022, 03, 17, 12, 0, 0, 0, time.UTC),
				EndsAt:    time.Date(2022, 03, 17, 12, 30, 0, 0, time.UTC),
			},
			wantErr: true,
			args: args{
				appointments: []appt{
					{
						createRequest: models.AppointmentCreateRequest{
							TrainerID: 1,
							UserID:    1,
							StartsAt:  time.Date(2022, 03, 17, 12, 0, 0, 0, time.UTC),
							EndsAt:    time.Date(2022, 03, 17, 12, 30, 0, 0, time.UTC),
						},
						wantAssert: true,
					},
					{
						createRequest: models.AppointmentCreateRequest{
							TrainerID: 1,
							UserID:    2,
							StartsAt:  time.Date(2022, 03, 17, 12, 0, 0, 0, time.UTC),
							EndsAt:    time.Date(2022, 03, 17, 12, 30, 0, 0, time.UTC),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PurgeTables()

			r := &AppointmentsRepoType{
				db: DB,
			}

			for _, appt := range tt.args.appointments {
				got, err := r.CreateAppointment(context.Background(), appt.createRequest)
				fmt.Printf("%#v", got)
				if err != nil && tt.wantErr {
					// I'd really prefer to assert the error otherwise we could have false positive tests
					return
				} else if err != nil {
					t.Fatal(err)
				} else if appt.wantAssert {
					// only asserting fields we want asserted based on test
					assert.Equal(t, tt.want.ID, got.ID)
					assert.Equal(t, tt.want.TrainerID, got.TrainerID)
					assert.Equal(t, tt.want.UserID, got.UserID)
					assert.Equal(t, tt.want.StartsAt, got.StartsAt)
					assert.Equal(t, tt.want.EndsAt, got.EndsAt)
				}
			}
		})
	}
}
