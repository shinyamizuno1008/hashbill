package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// newMySQLDB creates a new BookDatabase backed by a given MySQL server.
func newMySQLEventsDB(config MySQLConfig) (*eventDB, error) {
	// Check database and table exists. If not, create it.

	if err := config.ensureTableExisits(eventsTable); err != nil {
		return nil, err
	}

	conn, err := sql.Open("mysql", config.dataStoreName("event_list"))
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("mysql: could not establish a good connection: %v", err)
	}

	eventDB := &eventDB{
		mysqlDB: &mysqlDB{conn: conn},
	}

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. addEvent)

	if eventDB.list, err = conn.Prepare(listEventStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare list in event db: %v", err)
	}
	if eventDB.listedBy, err = conn.Prepare(listByHostAndEventStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare list by in event db: %v", err)
	}
	if eventDB.get, err = conn.Prepare(getEventStatementWithHostId); err != nil {
		return nil, fmt.Errorf("mysql: prepare get in event db: %v", err)
	}
	if eventDB.insert, err = conn.Prepare(insertEventStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert event db: %v", err)
	}
	if eventDB.update, err = conn.Prepare(updateEventStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare update event db: %v", err)
	}
	if eventDB.delete, err = conn.Prepare(deleteEventStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare delete event db: %v", err)
	}

	return eventDB, nil

}

// scanEvent reads a event from a sql.Row or sql.Rows
func scanEvent(s rowScanner) (*Event, error) {
	var (
		hostID      string
		eventName   string
		date        string
		deadline    string
		location    string
		membersMax  int64
		lottery     bool
		description string
	)
	if err := s.Scan(&hostID, &eventName, &date, &deadline, &location, &membersMax, &lottery, &description); err != nil {
		return nil, err
	}

	event := &Event{
		HostID:      hostID,
		EventName:   eventName,
		Date:        date,
		Deadline:    deadline,
		Location:    location,
		MembersMax:  membersMax,
		Lottery:     lottery,
		Description: description,
	}

	return event, nil
}

const listEventStatement = "SELECT * FROM events ORDER BY host_id"

// ListEvents returns a list of events.
func (eventDB *eventDB) ListEvents() ([]*Event, error) {
	rows, err := eventDB.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		events = append(events, event)
	}

	return events, nil
}

const listByHostAndEventStatement = `
	SELECT * FROM events 
	WHERE host_id = ? ORDER BY event_name
`

// ListEventsHostedBy returns a list of events, ordered by event name, filtered by
// the host id who created and host the event.
func (eventDB *eventDB) ListEventsHostedBy(hostID string) ([]*Event, error) {
	if hostID == "" {
		return eventDB.ListEvents()
	}

	rows, err := eventDB.listedBy.Query(hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		events = append(events, event)
	}

	return events, nil
}

const getEventStatementWithHostId = "SELECT * FROM events WHERE host_id = ? AND event_name = ?"

// GetEvent retrieves a event by its ID.
func (eventDB *eventDB) GetEvent(hostID, eventName string) (*Event, error) {
	event, err := scanEvent(eventDB.get.QueryRow(hostID, eventName))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mysql: could not find event with id %d", hostID)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get event: %v", err)
	}
	return event, nil
}

const insertEventStatement = `
	INSERT INTO events (
	host_id, event_name, date, deadline, location, members_max, lottery, description 
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

// AddEvent saves a given event.
func (eventDB *eventDB) AddEvent(e *Event) error {
	_, err := execAffectingOneRow(eventDB.insert, e.HostID, e.EventName,
		e.Date, e.Deadline, e.Location, e.MembersMax, e.Lottery, e.Description)
	if err != nil {
		return err
	}
	return nil
}

const updateEventStatement = `
	UPDATE events 
	SET host_id=?, event_name=?, date=?, deadline=?, location=?, members_max=?, lottery=?, description=?
	WHERE host_id = ? AND event_name = ?`

// UpdateEvent updates the entry for a given event.
func (eventDB *eventDB) UpdateEvent(e *Event) error {
	if e.HostID == "" && e.EventName == "" {
		return errors.New("mysql: event with unassigned host ID and event name passed into updateEvent")
	}

	_, err := execAffectingOneRow(eventDB.update, e.HostID, e.EventName)
	return err
}

const deleteEventStatement = "DELETE FROM events WHERE host_id = ? AND event_name = ?"

// DeleteEvent removes a given event by its host ID and event Name
func (eventDB *eventDB) DeleteEvent(hostID, eventName string) error {
	if hostID == "" && eventName == "" {
		return errors.New("mysql: book with unassigned ID passed into deleteEvent")
	}

	_, err := execAffectingOneRow(eventDB.delete, hostID, eventName)
	return err
}
