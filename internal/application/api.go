package application

import (
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/samuelmahr/appt-scheduling/internal/configuration"
	"github.com/samuelmahr/appt-scheduling/internal/repo"
	"github.com/samuelmahr/appt-scheduling/internal/routers"
	"log"
	"net/http"
	"time"
)

type APIApplication struct {
	config *configuration.AppConfig
	srv    *http.Server
}

func NewAPIApplication(c *configuration.AppConfig) *APIApplication {
	db, err := sqlx.Connect("postgres", c.DatabaseURL)
	if err != nil {
		log.Fatal("can't connect to db")
	}

	db.SetConnMaxLifetime(time.Duration(c.PostgresMaxConnLifetimeSeconds))
	db.SetMaxIdleConns(c.PostgresMaxIdleConns)
	db.SetMaxOpenConns(c.PostgresMaxOpenConns)

	appointmentsRepo := repo.NewAppointmentsRepository(db)
	rootRouter := mux.NewRouter()
	r := routers.NewV1Router(c, appointmentsRepo)
	r.Register(rootRouter)

	srv := &http.Server{
		Handler: rootRouter,
		Addr:    ":8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	return &APIApplication{
		config: c,
		srv:    srv,
	}
}

func (a *APIApplication) Run() {
	// hard coded... didn't know if I had it running or not
	log.Println("listening on port 8000")
	log.Fatal(a.srv.ListenAndServe())
}
