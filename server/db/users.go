package db

import (
	"database/sql"
	"errors"
	"fmt"
)

// newMySQLDB creates a new BookDatabase backed by a given MySQL server.
func newMySQLUsersDB(config MySQLConfig) (*userDB, error) {
	// Check database and table exists. If not, create it.
	if err := config.ensureTableExisits(usersTable); err != nil {
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

	userDB := &userDB{
		conn: conn,
	}

	// Prepared statements. The actual SQL queries are in the code near the
	// relevant method (e.g. addUser)

	if userDB.list, err = conn.Prepare(listUserStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare list in user db: %v", err)
	}
	if userDB.get, err = conn.Prepare(getUserStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare get in user db: %v", err)
	}
	if userDB.insert, err = conn.Prepare(insertUserStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare insert in user db: %v", err)
	}
	if userDB.update, err = conn.Prepare(updateUserStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare update in user db: %v", err)
	}
	if userDB.delete, err = conn.Prepare(deleteUserStatement); err != nil {
		return nil, fmt.Errorf("mysql: prepare delete in user db: %v", err)
	}

	return userDB, nil

}

// scanUser reads a user from a sql.Row or sql.Rows
func scanUser(s rowScanner) (*User, error) {
	var (
		userID   string
		userName string
	)
	if err := s.Scan(&userID, &userName); err != nil {
		return nil, err
	}

	user := &User{
		UserID:   userID,
		UserName: userName,
	}

	return user, nil
}

const listUserStatement = "SELECT * FROM users ORDER BY user_name"

// ListUsers returns a list of users.
func (userDB *userDB) ListUsers() ([]*User, error) {
	fmt.Println("hello from list users")
	rows, err := userDB.list.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("mysql: could not read row: %v", err)
		}

		users = append(users, user)
	}
	return users, nil
}

const getUserStatement = "SELECT * FROM users WHERE user_id = ?"

// GetUser retrieves a user by its ID.
func (userDB *userDB) GetUser(userID string) (*User, error) {
	user, err := scanUser(userDB.get.QueryRow(userID))
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("mysql: could not find user with id %s", userID)
	}
	if err != nil {
		return nil, fmt.Errorf("mysql: could not get user: %v", err)
	}
	fmt.Println(user)
	return user, nil
}

const insertUserStatement = `
	INSERT INTO users (
	user_id, user_name
	) VALUES (?, ?)
	`

// AddUser saves a given user.
func (userDB *userDB) AddUser(u *User) error {
	_, err := execAffectingOneRow(userDB.insert, u.UserID, u.UserName)
	if err != nil {
		return err
	}
	return nil
}

const updateUserStatement = `
	UPDATE users 
	SET user_id=?, user_name=? 
	WHERE user_id = ?`

func (userDB *userDB) UpdateUser(u *User) error {
	if u.UserID == "" {
		return errors.New("mysql: user with unassigned ID passed into updateBook")
	}

	_, err := execAffectingOneRow(userDB.update, u.UserID, u.UserName)
	return err
}

const deleteUserStatement = "DELETE FROM users WHERE user_id = ?"

func (userDB *userDB) DeleteUser(userID string) error {
	if userID == "" {
		return errors.New("mysql: book with unassigned ID passed into deleteUser")
	}

	_, err := execAffectingOneRow(userDB.delete, userID)
	return err
}
