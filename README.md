# Appointment Scheduling

# Table of Contents
- [Setup](#setup)
- [Design Considerations & More](#design-considerations--more)
- [Data Model](#data-model)
- [API](#api)
- [Get Scheduled Appointments](#get-scheduled-appointments)
- [Get Available Appointments](#get-available-appointments)
- [Create Appointment](#create-appointment)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Data Access](#data-access)
- [Smoke Test](#smoke-test)

## Setup
- `docker compose up` will stand up database
- Two options to load `appointments.json`:
  - From root directory run `go run ./scripts/initial_db_load.go` (uses libraries sqlx and squirrel)
  - Load however you know how to load a db 
- Run `./cmd/api/main.go` to start up the API
- Test with `http://localhost:8000`
- API documentation is in `./docs/api.yml`

## Design Considerations & More
### Data Model
I went with a postgres table in order to represent an API data store rather than updating the original file.

The db name, schema name, and table name were hard to do because naming is hard! Really, just named, so it wasn't database appointments, schema appointments, and table appointments
The main table has 3 extra columns that aren't in the json file example:
1. `created_at`
2. `updated_at`
3. `canceled_at`

The third additional column is just for future flexibility. Not necessarily needed in this take home.

There is a unique index on `trainer_id`, `starts_at`, `ends_at` and `canceled_at = null`. 
This unique index will be helpful when searching for a trainer's specific availability and also to prevent creating a double booked appointment.

Due to preloading the appointment data from `appointments.json`, the pkey sequence may be incorrect (starting at ID 1 when it already exists) so I restarted it at 1000 for the primary key in the initial db migration

Additional indexes I would consider for the future is an index on `user_id` and an index on `trainer_id`, but it's not necessary for this exercise

### API
API documentation is in `./docs/api.yml`. The main thing documented are happy paths, error responses are not in the docs

#### Get Scheduled Appointments
Path: `GET /appointments/scheduled`

I separated available and scheduled endpoints just for code isolation purposes.
They're very similar, but end up returning a completely different context of data set even though the properties are the same

The prompt mentioned the method to get appointments by only #1 below, but I added #2 because I misread!
1. by trainer
2. by start/end for a trainer

This endpoint accepts query params, and based on what query params is how the response is filtered when querying the database.

If there are additional query params added that are unexpected, they will be ignored.

If there are no params submitted, it will return all appointments

The accepted time format for start/end params is`time.RFC3339`

Times returned are in UTC... It felt normal to do that than to make all times Pacific.
Internally the API will validate time zone, but in communications it is in UTC

I did not add pagination to start, but if a business case required it (tables to display), then I would add it in

#### Get Available Appointments
Path: `GET /appointments/available`

I separated available and scheduled endpoints just for code isolation purposes.
They're very similar, but end up returning a completely different context of data set even though the properties are the same

The prompt mentioned the way to get available appointments:
1. by trainer
2. by start/end for a trainer

Get available appointments was a little tricky because we know what's scheduled, but I didn't want to loop through too many times to build time slots.
I used the unix time of start:end for schedule appointments as a way to track what timeslots are unavailable as I built the list of available timeslots from start/end datetime.
The available time slots should be during business hours pacific time, though the API returns UTC times.

The response will use the same object as List Scheduled Appointments, except it will omit the user ID

If there are additional query params added that are unexpected, they will be ignored.

If there are no params submitted, it will return all appointments

The accepted time format for start/end params is`time.RFC3339`

Again, I did not add pagination to start, but if a business case required it (tables to display), then I would add it in

#### Create Appointment
Path: `POST /appointments`
My assumption is that you can list appointments that a trainer is available and then pick a time slot to create an appointment.

This should be relatively safe to just create any appointment, but as a safety net, there will be no double-booked appointments with the unique index set in the table

Validations in controller:
- 30-minute time slot
- starts or ends on 00 and 30 
- time is within business hours M-F

The repo layer will insert what ever it is given, which should be fair based on validations


Times returned are in UTC... It felt normal to do that than to make all times Pacific.
Internally the API will validate time zone, but in communications it is in UTC

The accepted time format via API is `time.RFC3339`

API Response will echo the appointment that was just created

For the case the time slot is already booked, it returns a db specific error in the API. 
If I spent more time on it, I'd fix the error handling to return a more user-friendly error

### Project Structure
This is basically how I'm used writing Go applications, except for models package. I just wanted to separate out structs and see if I like it better this way.

In a larger service, there may be multiple objects that would make this project structure make more sense.
It's a little overkill for this project, but it can be expanded on just for fun (add users table, trainers table, etc)

### Testing
Normally I would have unit tests mock most interfaces except for code that actually accesses data (`repo` package)

Code hitting the database does have unit tests that will persist data into the table.
I prefer to ensure interactions with the database work as expected.

I am only testing the repo package just for read/write sanity. Ignoring the controller package, the other packages are mostly app/config/router setup.

There are tests in controller package, hitting 81.8% code coverage. 
Tests are mainly for obvious issues like invalid requests and happy path

Having the database running with docker compose is required.

### Data Access
In the code, you may see I have a written out query for the insert, and I am using squirrel to build the List query.
1. I prefer queries to be written out, so you know what is actually being run
2. If I can't dynamically build a query (in a pretty manner), it's nice to use squirrel help dynamically build queries

It should be straightforward overall. One query inserts unless there is a constraint, while the query gets a list of data.

## Smoke Test
Ran a smoke test for simple happy path scenario, and it appeared to work as expected.
1. Call scheduled endpoint to see what is scheduled, make sure data is returned with different filters
2. Call available endpoint to see what time slots are available (around the time slots in `appointments.json`)
3. Call POST to create an appointment with a time slot available from #2, added user_id, and it created
4. Called scheduled to see new appointment scheduled, it was there
5. Called available to see the time slot is no longer available
6. Called to create another appointment with the same time slot for same trainer and an error was returned