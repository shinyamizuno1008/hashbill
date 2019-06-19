package db

// EventListDatabase proviedes thread-safe access to a database of event list.
type EventListDatabase interface {
	UserDatabase
	EventDatabase
	ParticipantDatabase
}

// User holds metadata about a user.
type User struct {
	UserID   string `json:"userID"`
	UserName string `json:"userName"`
}

// UserDatabase provides thread-safe access to a database of users.
type UserDatabase interface {
	// ListUsers() returns a list of event.
	ListUsers() ([]*User, error)

	// GetUser retrieves a user by its ID.
	GetUser(userID string) (*User, error)

	// AddUser saves a given user.
	AddUser(u *User) error

	// DeleteEvent removes a given user by its ID.
	DeleteUser(userID string) error

	// UpdateEvent updates the entry for a given Event.
	UpdateUser(u *User) error
}

// Event holds metadata about a event.
type Event struct {
	HostID      string
	EventName   string
	Date        string
	Deadline    string
	Location    string
	MembersMax  int64
	Lottery     bool
	Description string
}

// EventDatabase provides thread-safe access to a database of events.
type EventDatabase interface {
	// ListUsers() returns a list of event.
	ListEvents() ([]*Event, error)

	// ListEventCreatedBy returns a list of event, filterred by
	// the user who created the book entry.
	ListEventsHostedBy(hostID string) ([]*Event, error)

	// GetEvent retrieves a event by its ID.
	GetEvent(hostID, eventName string) (*Event, error)

	// AddEvent saves a given event.
	AddEvent(e *Event) error

	// DeleteEvent removes a given event by its ID.
	DeleteEvent(hostID, eventName string) error

	// UpdateEvent updates the entry for a given Event.
	UpdateEvent(e *Event) error
}

// Participant holds metadata about a participant.
type Participant struct {
	HostID        string
	EventName     string
	ParticipantID string
}

// ParticipantDatabase provides thread-safe access to a database of participants.
type ParticipantDatabase interface {
	// ListUsers() returns a list of participants.
	ListParticipants() ([]*Participant, error)

	// ListParticipantsHostedBy returns a list of participant, filterred by
	// the event they are going to participante in and the host who host the event.
	ListParticipantsHostedBy(hostID, evetntName string) ([]*Participant, error)

	// Get retrieves a participant of a specific participant by its ID.
	GetParticipant(p *Participant) (*Participant, error)

	// AddUser saves a given partifipant of a specific event .
	AddParticipant(p *Participant) error

	// DeleteEvent removes a given user of a specific event by its ID.
	DeleteParticipant(p *Participant) error

	// UpdateEvent updates the entry for a given participant of a specific event .
	UpdateParticipant(p *Participant) error
}
