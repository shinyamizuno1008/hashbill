package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/shinyamizuno1008/hashbill/server/db"
)

func main() {
	r := mux.NewRouter()

	r.Methods("GET").Path("/user/{userID}").Handler(appHandler(getUserHandler))
	r.Methods("GET").Path("/userlist").Handler(appHandler(getAllUserHandler))
	r.Methods("GET").Path("/event/list").Handler(appHandler(getEventsHandler))
	r.Methods("POST").Path("/event/register").Handler(appHandler(registerEventHandler))
	r.Methods("POST").Path("/signup").Handler(appHandler(signupHandler))
	// r.PathPrefix("/").Handler(http.FileServer(http.Dir("../client/dist")))
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8000", r))
}

// signupHandler adds user to the database.
func signupHandler(w http.ResponseWriter, r *http.Request) *appError {
	userID := r.FormValue("userID")
	userName := r.FormValue("userName")

	if err := db.DB.AddUser(&db.User{
		UserID:   userID,
		UserName: userName,
	}); err != nil {
		return appErrorf(err, "could not add user: %v", err)
	}
	return nil
}

// registerEventHandler adds event to the database
func registerEventHandler(w http.ResponseWriter, r *http.Request) *appError {
	event, err := eventFromForm(r)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	if err := db.DB.AddEvent(event); err != nil {
		return appErrorf(err, "could not add event: %v", err)
	}
	return nil
}

// eventFromRequest retrieves a event from the database given a host id and event name
// in the URL's path.
func eventFromRequest(r *http.Request) (*db.Event, error) {
	hostID := r.FormValue("hostID")
	eventName := r.FormValue("eventName")
	event, err := db.DB.GetEvent(hostID, eventName)
	if err != nil {
		return nil, fmt.Errorf("could not find event: %v", err)
	}
	return event, nil
}

// addUserRequest adds a user to the database.
func addUserRequest(w http.ResponseWriter, r *http.Request) *appError {
	user := &db.User{
		UserID:   r.FormValue("userID"),
		UserName: r.FormValue("userName"),
	}

	err := db.DB.AddUser(user)
	if err != nil {
		return appErrorf(err, "could not add user: %v", err)
	}

	return nil
}

// eventFromForm populates the fields of a event from form values.
func eventFromForm(r *http.Request) (*db.Event, error) {
	membersMax, err := strconv.ParseInt(r.FormValue("membersMax"), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse members max: %v", err)
	}
	lottery, err := strconv.ParseBool(r.FormValue("lottery"))
	if err != nil {
		return nil, fmt.Errorf("could not parse lottery max: %v", err)
	}

	event := &db.Event{
		HostID:      r.FormValue("hostID"),
		EventName:   r.FormValue("eventName"),
		Date:        r.FormValue("eventDate") + " " + r.FormValue("eventTime"),
		Deadline:    r.FormValue("deadlineDate") + " " + r.FormValue("deadlineTime"),
		Location:    r.FormValue("location"),
		MembersMax:  membersMax,
		Lottery:     lottery,
		Description: r.FormValue("description"),
	}

	return event, nil
}

// updateEventHandler updates the details of a given event.
func updateEventHanlder(w http.ResponseWriter, r *http.Request) *appError {
	hostID := mux.Vars(r)["hostID"]
	eventName := mux.Vars(r)["eventName"]

	event, err := eventFromForm(r)
	if err != nil {
		return appErrorf(err, "could not parse event from form: %v", err)
	}

	event.HostID = hostID
	event.EventName = eventName

	err = db.DB.UpdateEvent(event)
	if err != nil {
		return appErrorf(err, "could not save event: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/events/%s", event.HostID+eventName), http.StatusFound)
	return nil
}

// deleteHandler deletes a given event.
func deleteEventHandler(w http.ResponseWriter, r *http.Request) *appError {
	hostID := mux.Vars(r)["hostID"]
	eventName := mux.Vars(r)["eventName"]

	err := db.DB.DeleteEvent(hostID, eventName)
	if err != nil {
		return appErrorf(err, "could not save event: %v", err)
	}
	http.Redirect(w, r, "/events", http.StatusFound)
	return nil
}

// getUserHanlder show user.
func getUserHandler(w http.ResponseWriter, r *http.Request) *appError {
	userID := mux.Vars(r)["userID"]
	user, err := db.DB.GetUser(userID)
	if err != nil {
		return appErrorf(err, "could not get user from database: %v", err)
	}

	userJSON, err := json.Marshal(*user)
	w.Write(userJSON)
	return nil
}

//getAllUserHandler show all users.
func getAllUserHandler(w http.ResponseWriter, r *http.Request) *appError {
	users, err := db.DB.ListUsers()
	if err != nil {
		return appErrorf(err, "could not get users from database: %v", err)
	}

	usersJSON, err := json.Marshal(users)
	w.Write(usersJSON)
	return nil
}

// getEventsHandler show all registered evetns.
func getEventsHandler(w http.ResponseWriter, r *http.Request) *appError {
	fmt.Println("hello from event handler")
	events, err := db.DB.ListEvents()
	if err != nil {
		return appErrorf(err, "could not get events from database: %v", err)
	}

	eventsJSON, err := json.Marshal(events)
	w.Write(eventsJSON)
	return nil
}

// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}
