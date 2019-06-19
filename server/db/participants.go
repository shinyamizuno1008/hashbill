package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// newMySQLDB creates a new BookDatabase backed by a given MySQL server.
func newMySQLParticipantsDB(config MySQLConfig) (*participantDB, error) {
	// Check database and table exists. If not, create it.
	if err := config.ensureTableExisits(participantsTable); err != nil {
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

	participantDB := &participantDB{
		mysqlDB: &mysqlDB{conn: conn},
	}

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. addParticipant)

	if participantDB.list, err = conn.Prepare(listParticipantStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare list in participant db: %v", err)
	}
	if participantDB.listedBy, err = conn.Prepare(listParticipantHostedByStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare listedby in participant db: %v", err)
	}
	if participantDB.get, err = conn.Prepare(getParticipantStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare get in participant db: %v", err)
	}
	if participantDB.insert, err = conn.Prepare(insertParticipantStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert in participant db: %v", err)
	}
	if participantDB.update, err = conn.Prepare(updateParticipantStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare update in partcipant db: %v", err)
	}
	if participantDB.delete, err = conn.Prepare(deleteParticipantStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare delete in participant db: %v", err)
	}

	return participantDB, nil

}

// scanParticipant reads a user from a sql.Row or sql.Rows
func scanParticipant(s rowScanner) (*Participant, error) {
	var (
		hostID        string
		eventName     string
		participantID string
	)
	if err := s.Scan(&hostID, &eventName, &participantID); err != nil {
		return nil, err
	}

	participant := &Participant{
		HostID:        hostID,
		EventName:     eventName,
		ParticipantID: participantID,
	}

	return participant, nil
}

const listParticipantStatement = "SELECT * FROM participants ORDER BY participant_id"

// ListParticipants returns a list of users.
func (participantDB *participantDB) ListParticipants() ([]*Participant, error) {
	rows, err := participantDB.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*Participant
	for rows.Next() {
		participant, err := scanParticipant(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

const listParticipantHostedByStatement = `
	SELECT * FROM participants 
	WHERE host_id = ? AND event_name = ?
	ORDER BY event_name
`

// ListEventsHostedBy returns a list of participants, ordered by name, filtered by
// the event they are going to participate in and the host id who created and host the event.
func (participantDB *participantDB) ListParticipantsHostedBy(hostID, eventName string) ([]*Participant, error) {
	if hostID == "" || eventName == "" {
		return participantDB.ListParticipants()
	}

	rows, err := participantDB.listedBy.Query(hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []*Participant
	for rows.Next() {
		participant, err := scanParticipant(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

const getParticipantStatement = "SELECT * FROM participants WHERE host_id = ? AND event_name"

// GetParticipant retrieves a participant by its ID.
func (participantDB *participantDB) GetParticipant(p *Participant) (*Participant, error) {
	participant, err := scanParticipant(participantDB.get.QueryRow(p.HostID, p.EventName, p.ParticipantID))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mysql: could not find participant with ID %s in event %s hosted by host %s", p.HostID, p.EventName, p.ParticipantID)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get user: %v", err)
	}
	return participant, nil
}

const insertParticipantStatement = `
	INSERT INTO participants (
	host_id, event_name, participant_id
	) VALUES (?, ?, ?)
	`

// AddParticipant saves a given participant.
func (participantDB *participantDB) AddParticipant(p *Participant) error {
	_, err := execAffectingOneRow(participantDB.insert, p.HostID, p.EventName, p.ParticipantID)
	if err != nil {
		return err
	}
	return nil
}

const updateParticipantStatement = `
	UPDATE participants 
	SET host_id=?, event_name=?, participant_id=? 
	WHERE host_id=? AND event_name=? AND participant_id=?`

func (participantDB *userDB) UpdateParticipant(p *Participant) error {
	if p.ParticipantID == "" {
		return errors.New("mysql: user with unassigned ID passed into updateBook")
	}

	_, err := execAffectingOneRow(participantDB.update, p.HostID, p.EventName, p.ParticipantID)
	return err
}

const deleteParticipantStatement = "DELETE FROM participants WHERE host_id = ? AND event_name = ? AND participant_id = ?"

func (participantDB *userDB) DeleteParticipant(p *Participant) error {
	if p.HostID == "" || p.EventName == "" || p.ParticipantID == "" {
		return errors.New("mysql: book with unassigned ID passed into deleteParticipant")
	}

	_, err := execAffectingOneRow(participantDB.delete, p.HostID, p.EventName, p.ParticipantID)
	return err
}
