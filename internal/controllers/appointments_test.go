package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/samuelmahr/appt-scheduling/internal/configuration"
	"github.com/samuelmahr/appt-scheduling/internal/models"
	"github.com/samuelmahr/appt-scheduling/internal/repo"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var aRepo repo.AppointmentsRepository
var config *configuration.AppConfig
var appointmentsController V1AppointmentsController

func setup() {
	var err error
	config, err = configuration.Configure()
	if err != nil {
		panic("configuration error")
	}
}

func teardown() {

}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	teardown()
	os.Exit(code)
}

func TestV1Appointments_CreateAppointment(t *testing.T) {
	type args struct {
		ctx     context.Context
		request []byte
		aRepo   repo.MockAppointments
	}

	tests := []struct {
		name     string
		args     args
		response int
		errMsg   string
	}{
		{
			name: "success within business hours",
			args: args{
				ctx: context.TODO(),
				request: []byte(`{
					"user_id": 1,
					"trainer_id": 1,
					"starts_at": "2022-03-17T19:00:00Z",
					"ends_at": "2022-03-17T19:30:00Z"
				}`),
				aRepo: repo.MockAppointments{
					CreateAppointmentsResponse: models.Appointment{
						ID:        1,
						TrainerID: 1,
						UserID:    1,
						StartsAt:  time.Date(2022, 03, 17, 19, 0, 0, 0, time.UTC),
						EndsAt:    time.Date(2022, 03, 17, 19, 30, 0, 0, time.UTC),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}},
			},
			response: http.StatusCreated,
		},
		{
			name: "fail outside business hours",
			args: args{
				ctx: context.TODO(),
				request: []byte(`{
					"user_id": 1,
					"trainer_id": 1,
					"starts_at": "2022-03-17T08:00:00Z",
					"ends_at": "2022-03-17T08:30:00Z"
				}`),
				aRepo: repo.MockAppointments{},
			},
			response: http.StatusBadRequest,
			errMsg:   "bad request payload, check times",
		},
	}

	url := "/appointments"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aRepo = &tt.args.aRepo

			appointmentsController = NewV1AppointmentsController(config, aRepo)

			getHandler := http.HandlerFunc(appointmentsController.CreateAppointment)

			req, err := http.NewRequest("POST", url, bytes.NewReader(tt.args.request))
			if err != nil {
				t.Fatal(err)
			}

			response := httptest.NewRecorder()
			getHandler.ServeHTTP(response, req)
			assert.Equal(t, tt.response, response.Code)

			if tt.response != http.StatusCreated {
				resp := make(map[string]string)
				err = json.Unmarshal(response.Body.Bytes(), &resp)
				assert.Equal(t, tt.errMsg, resp["error"])
			}
		})
	}
}
