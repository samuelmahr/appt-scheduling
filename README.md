# Appointment Scheduling

## Setup
- `docker compose up` will stand up database
- Load `appointments.json`:
  - From root directory run `go run ./scripts/initial_db_load.go` (uses libraries sqlx and squirrel)
  - Load however you know how to load a db 
- Run `./cmd/api/main.go` to start up the API
- Test with `http://localhost:8000`

## Design Considerations & More
### Project Structure
This is basically how I'm used writing Go applications,
except for models package. I just wanted to separate out structs and see if I like it better this way.

### Testing
Unit tests mock most interfaces except for code that actually accesses data (`repo` package)

Code hitting the database does have unit tests that will persist data into the table.
I prefer to ensure interactions with the database work as expected.

Having the database running with docker compose is required.

### Data Model
The db name, schema name, and table name were hard to do because naming is hard! Really, just named so it wasn't database appointments, schema appointments, and table appointments
The main table has 3 extra columns that aren't in the json file example:
1. `created_at`
2. `updated_at`
3. `canceled_at`

The third additional column is just for future flexibility. Not necessarily needed in this take home.

There is a unique index on `trainer_id`, `starts_at`, `ends_at` and `canceled_at = null`. 
This unique index will be helpful when searching for a trainer's specific availability and also to prevent creating a double booked appointment.

Additional indexes I would consider for the future is an index on `user_id` and an index on `trainer_id`, but it's not necessary for this exercise

### Data Access
I went with a postgres table in order to represent an API data store rather than updating the original file.
In the code, you may see I have a written out query for the insert, and I am using squirrel to build the List query.
1. I prefer queries to be written out, so you know what is actually being run
2. If I can't dynamically build a query (in a pretty manner), it's nice to use squirrel help dynamically build queries

It should be straightforward overall. One query inserts unless there is a constraint, while the query gets a list of data.

I did not add pagination to start, but if a business case required it (tables to display), then I would add it in

### API Endpoints
#### Get Appointments
The prompt mentioned two ways to get appointments
1. by trainer
2. by start/end for a trainer

This will only need one endpoint. It will work well as a list endpoint with query params, and based on what query params is how it's filtered when querying the database.

If there are additional query params added that are unexpected, they will be ignored.

If there are no params submitted, it will return all appointments

The accepted time format for start/end params is`time.RFC3339`

Again, I did not add pagination to start, but if a business case required it (tables to display), then I would add it in

#### Create Appointment
My assumption is that you can list appointments that a trainer is available and then pick a time slot to create an appointment.

This should be relatively safe to just create any appointment, but as a safety net, there will be no double-booked appointments with the unique index set in the table

Validation that it is a 30-minute time slot that starts or ends on 00 and 30 will be validated in the controller and the repo layer will insert what ever it is given

The accepted time format via API is `time.RFC3339`