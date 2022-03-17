package main

import (
	"encoding/json"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"io/ioutil"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type appointments struct {
	ID        int64     `json:"id"`
	TrainerID int64     `json:"trainer_id"`
	UserID    int64     `json:"user_id"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
}

func main() {
	jsonFile, err := os.Open("./appointments.json")
	if err != nil {
		panic(err)
	}

	fmt.Println("file read")
	defer jsonFile.Close()

	bytes, _ := ioutil.ReadAll(jsonFile)

	var appointments []appointments
	err = json.Unmarshal(bytes, &appointments)
	if err != nil {
		panic(err)
	}

	fmt.Println("unmarshaled")

	query := sq.Insert("scheduling.appointments").Columns("id", "trainer_id", "user_id", "starts_at", "ends_at").PlaceholderFormat(sq.Dollar)
	fmt.Println("building insert query")
	for _, appt := range appointments {
		// original print debug logs
		// fmt.Printf("ID: %d\n", appt.ID)
		// fmt.Printf("Trainer ID: %d\n", appt.TrainerID)
		// fmt.Printf("User ID: %d\n", appt.UserID)
		// fmt.Printf("Starts At: %v\n", appt.StartsAt)
		// fmt.Printf("Ends At: %v\n", appt.EndsAt)

		query = query.Values(appt.ID, appt.TrainerID, appt.UserID, appt.StartsAt, appt.EndsAt)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		panic(err)
	}

	// fmt.Printf("Query: %s\n", sql)
	// fmt.Printf("args: %#v\n", args)

	db, err := sqlx.Connect("postgres", "postgres://master:Passw0rd@localhost:5432/fitness?sslmode=disable&TimeZone=utc")
	if err != nil {
		panic(err)
	}

	fmt.Println("inserting...")
	_, err = db.Exec(sql, args...)
	if err != nil {
		panic(err)
	}

}
