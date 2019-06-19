package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

type tables struct {
	users, evetns, participants string
}

const usersTable = "users"
const eventsTable = "events"
const participantsTable = "participants"

var a = &tables{users: ""}

var createTableStatements = []string{
	`CREATE DATABASE IF NOT EXISTS event_list DEFAULT CHARACTER SET = 'utf8' DEFAULT COLLATE 'utf8_general_ci';`,
	`USE event_list;`,
	`CREATE TABLE IF NOT EXISTS users (
		user_id VARCHAR(255) NOT NULL,
		user_name VARCHAR(255) NOT NULL,
		PRIMARY KEY (user_id)
	);`,
	`CREATE TABLE IF NOT EXISTS events (
		host_id VARCHAR(255) NOT NULL,
		event_name VARCHAR(255) NOT NULL,
		date DATETIME NOT NULL,
		deadline DATETIME NOT NULL,
		location VARCHAR(512) NOT NULL,
		members_max INT NULL,
		lottery BOOL DEFAULT FALSE,
		description VARCHAR(1024) NULL,
		PRIMARY KEY (host_id, event_name)
	);`,
	`CREATE TABLE IF NOT EXISTS participants (
		host_id VARCHAR(255) NOT NULL, 
		event_name VARCHAR(255) NOT NULL,
		participant_id VARCHAR(255) NOT NULL,
		PRIMARY KEY (host_id,event_name,participant_id),
		FOREIGN KEY (host_id) REFERENCES users(user_id),
		FOREIGN KEY (participant_id) REFERENCES users(user_id)
	);`,
}

// mysqlDB persists books to a MySQL instance.
type mysqlDB struct {
	conn   *sql.DB
	list   *sql.Stmt
	insert *sql.Stmt
	get    *sql.Stmt
	update *sql.Stmt
	delete *sql.Stmt
}

type eventListDB struct {
	*userDB
	*eventDB
	*participantDB
}

type userDB mysqlDB
type eventDB struct {
	*mysqlDB
	listedBy *sql.Stmt
}
type participantDB struct {
	*mysqlDB
	listedBy *sql.Stmt
}

// Ensure mysqlDB conforms to the EventDatabase interface.
var _ EventListDatabase = &eventListDB{}

type MySQLConfig struct {
	Username, Password string

	// Host of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Host string

	// Port of the MySQL instance.
	//
	// If set, UnixSocket should be unset.
	Port int

	// UnixSocket is the filepath to a unix socket.
	//
	// If set, Host and Port should be unset.
	UnixSocket string
}

// Close closes the database, freeing up any resources.
func (db *mysqlDB) Close() {
	db.conn.Close()
}

// rowScanner is implemented by sql.Row and sql.Rows
type rowScanner interface {
	Scan(dest ...interface{}) error
}

func newMySQLDB(config MySQLConfig) (*eventListDB, error) {
	userDB, err := newMySQLUsersDB(config)
	if err != nil {
		return nil, err
	}
	eventDB, err := newMySQLEventsDB(config)
	if err != nil {
		return nil, err
	}
	participantDB, err := newMySQLParticipantsDB(config)
	if err != nil {
		return nil, err
	}

	db := &eventListDB{
		userDB:        userDB,
		eventDB:       eventDB,
		participantDB: participantDB,
	}

	return db, nil
}

// dataStoreName returns a connection string suitable for sql.Open
func (c MySQLConfig) dataStoreName(databaseName string) string {
	var cred string
	// [username[:passowrd]@]
	if c.Username != "" {
		cred = c.Username
		if c.Password != "" {
			cred = cred + ":" + c.Password
		}
		cred = cred + "@"
	}
	return fmt.Sprintf("%stcp([%s]:%d)/%s", cred, c.Host, c.Port, databaseName)
}

func (config MySQLConfig) ensureTableExisits(tableName string) error {
	conn, err := sql.Open("mysql", config.dataStoreName(""))
	if err != nil {
		return fmt.Errorf("mysql: could not get a connection: %v", err)
	}
	defer conn.Close()

	// Check the connection.
	if conn.Ping() == driver.ErrBadConn {
		return fmt.Errorf("mysql: could not connect to the database." +
			"could be bad address, or this address is not whitelisted for access.")
	}

	if _, err := conn.Exec("USE event_list"); err != nil {
		// MySQL error 1049 is "database does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1049 {
			return createTable(conn)
		}

		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the databae: %v", err)
	}

	if _, err := conn.Exec("DESCRIBE " + tableName); err != nil {
		// MySQL error 1146 is "table does not exist"
		if mErr, ok := err.(*mysql.MySQLError); ok && mErr.Number == 1146 {
			return createTable(conn)
		}
		// Unknown error.
		return fmt.Errorf("mysql: could not connect to the database: %v", err)
	}
	return nil
}

// createTable creates the table, and if necessary, the databse.
func createTable(conn *sql.DB) error {
	for _, stmt := range createTableStatements {
		_, err := conn.Exec(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

// execAffectingOneRow executes a given statement, expecting one row to be affected.
func execAffectingOneRow(stmt *sql.Stmt, args ...interface{}) (sql.Result, error) {
	r, err := stmt.Exec(args...)
	if err != nil {
		return r, fmt.Errorf("mysql: could not execute statement: %v", err)
	}

	rowsAffected, err := r.RowsAffected()
	if err != nil {
		return r, fmt.Errorf("mysql: could not get rows affected: %v", err)
	} else if rowsAffected != 1 {
		return r, fmt.Errorf("mysql: expected 1 row affected, got %d", rowsAffected)
	}
	return r, nil
}
