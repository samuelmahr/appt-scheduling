package repo

import (
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/jmoiron/sqlx"
	"github.com/samuelmahr/appt-scheduling/internal/configuration"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	// blank import is ok here
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/lib/pq"
)

var config *configuration.AppConfig
var DB *sqlx.DB
var migrateDir string
var migrator *migrate.Migrate

func SetupTestDB() {
	c, err := configuration.Configure()
	if err != nil {
		fmt.Println("error loading config. "+err.Error(), err)
		os.Exit(1)
	}

	if c.TestDatabaseURL == "" {
		fmt.Println("no test database url configured. exiting")
		os.Exit(1)
	}

	if c.TestDatabaseURL == c.DatabaseURL {
		fmt.Println("test DB url is the same as actual database url, this is probably unintentional. exiting")
		os.Exit(1)
	}

	config = c
	DB = dbCon(config.TestDatabaseURL)

	md := findMigrationRoot()
	if md == "" {
		fmt.Println("error: could not find migrate directory")
		os.Exit(1)
	}

	migrateDir = md

	migrator, err = migrate.New("file://"+migrateDir, config.TestDatabaseURL)
	if err != nil {
		fmt.Println("error setting up migrator. exiting", err)
		os.Exit(1)
	}

	withTimeout(time.Second*2, func() error {
		if err := migrator.Down(); err != nil && err != migrate.ErrNoChange {
			fmt.Println("error purging database. "+err.Error(), err)
			return err
		}

		if err := migrator.Up(); err != nil {
			fmt.Println("error migrating database after purge. "+err.Error(), err)
			return err
		}

		return nil
	})
}

func dbCon(url string) *sqlx.DB {
	var d *sqlx.DB
	var err error

	if d, err = sqlx.Connect("postgres", url); err != nil {
		fmt.Println("error connection to DB. "+err.Error(), err)
		os.Exit(1)
	}

	return d
}

func findMigrationRoot() string {
	p, _ := os.Getwd()
	for p != "" && p != "/" {
		infos, err := ioutil.ReadDir(p)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		for _, i := range infos {
			if i.IsDir() && i.Name() == "migrations" {
				return filepath.Join(p, "migrations")
			}
		}

		p = filepath.Dir(p)
	}

	return ""
}

func PurgeTables() {
	withTimeout(time.Second*2, func() error {
		if _, err := DB.Exec("delete from scheduling.appointments;"); err != nil {
			return err
		}
		return nil
	})
	withTimeout(time.Second*2, func() error {
		if _, err := DB.Exec("ALTER SEQUENCE scheduling.appointments_id_seq RESTART WITH 1;"); err != nil {
			return err
		}
		return nil
	})
}

func withTimeout(timeout time.Duration, work func() error) {
	timeoutch := time.After(timeout)
	workch := make(chan bool, 1)
	var workErr error

	go func(c chan<- bool) {
		workErr = work()
		c <- true
	}(workch)

	select {
	case _ = <-timeoutch:
		debug.PrintStack()
		log.Fatal("timeout")
	case _ = <-workch:
		if workErr != nil {
			log.Fatal(workErr)
		}
	}
}
