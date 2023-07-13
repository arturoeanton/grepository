package repositories

import (
	"database/sql"
	"os"
	"testing"

	"github.com/arturoeanton/go-struct2serve/config"
	_ "github.com/mattn/go-sqlite3"
)

func MockSqlite() (*sql.DB, error) {
	//os.Remove()
	config.FlagLog = false

	filePath := "./test.db"
	_, err := os.Stat(filePath)
	if !os.IsNotExist(err) {
		os.Remove(filePath)
	}

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return db, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS user (id INTEGER PRIMARY KEY, first_name TEXT, email TEXT, group_id INTEGER)")
	if err != nil {
		return db, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS roles (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		return db, err
	}
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS user_roles (id INTEGER PRIMARY KEY, user_id INTEGER, role_id INTEGER)")
	if err != nil {
		return db, err
	}

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS groups (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		return db, err
	}
	//validate if exist users
	var count int
	err = db.QueryRow("SELECT count(*) FROM roles").Scan(&count)
	if err != nil {
		return db, err
	}
	if count == 0 {
		_, err = db.Exec("INSERT INTO roles (name) VALUES ('admin')")
		if err != nil {
			return db, err
		}
		_, err = db.Exec("INSERT INTO roles (name) VALUES ('user')")
		if err != nil {
			return db, err
		}

		_, err = db.Exec("INSERT INTO groups (name) VALUES ('group1')")
		if err != nil {
			return db, err
		}
		_, err = db.Exec("INSERT INTO groups (name) VALUES ('group2')")
		if err != nil {
			return db, err
		}

		_, err = db.Exec("INSERT INTO user (first_name, email, group_id) VALUES ('admin', 'admin@admin.com',1)")
		if err != nil {
			return db, err
		}
		_, err = db.Exec("INSERT INTO user (first_name, email, group_id) VALUES ('user', 'user@user.com',1)")
		if err != nil {
			return db, err
		}

		_, err = db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (1, 1)")
		if err != nil {
			return db, err
		}
		_, err = db.Exec("INSERT INTO user_roles (user_id, role_id) VALUES (2, 2)")
		if err != nil {
			return db, err
		}

	}
	return db, nil
}

type User struct {
	IDUser    int    `json:"id" db:"id" sql_id:"true"` // mark this field as id with tag sql_id:"true"
	FirstName string `json:"first_name" db:"first_name"`
	Email     string `json:"email" db:"email"`
	Roles     []Role `json:"roles" sql:"select * from roles r where r.id in (select role_id from user_roles where user_id = ?)"`
	GroupId   *int   `json:"-" db:"group_id" sql_update_value:"Group.ID"` // mark this field as id with tag sql_update_value:"Group.ID" because json not send nil values json:"-"
	Group     Group  `json:"group" sql:"select * from groups where id = ?" sql_param:"GroupId"`
}

type Role struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

type Group struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}

func TestGetAll(t *testing.T) {
	config.DB, _ = MockSqlite()
	defer config.DB.Close()

	repoUser := NewRepository[User]()
	repoRole := NewRepositoryWithTable[Role]("roles")
	users, _ := repoUser.GetAll()
	if len(users) == 0 {
		t.Error("users is empty")
	}
	roles, _ := repoRole.GetAll()
	if len(roles) == 0 {
		t.Error("roles is empty")
	}
}

func TestGetByID(t *testing.T) {
	config.DB, _ = MockSqlite()
	defer config.DB.Close()

	repoUser := NewRepository[User]()
	repoRole := NewRepositoryWithTable[Role]("roles")
	user, _ := repoUser.GetByID(1)
	if user.IDUser != 1 {
		t.Error("user is empty")
	}

	if user.FirstName != "admin" {
		t.Error("user name is not admin")
	}

	role, _ := repoRole.GetByID(1)
	if role.ID != 1 {
		t.Error("role is empty")
	}
	if role.Name != "admin" {
		t.Error("role name is not admin")
	}
}

func TestGetByCriteria(t *testing.T) {
	config.DB, _ = MockSqlite()
	defer config.DB.Close()

	repoUser := NewRepository[User]()
	repoRole := NewRepositoryWithTable[Role]("roles")
	users, _ := repoUser.GetByCriteria("first_name = ?", "admin")
	if len(users) == 0 {
		t.Error("users is empty")
	}
	roles, _ := repoRole.GetByCriteria("name = ?", "admin")
	if len(roles) == 0 {
		t.Error("roles is empty")
	}
}

func TestDelete(t *testing.T) {
	config.DB, _ = MockSqlite()
	defer config.DB.Close()

	repoUser := NewRepository[User]()
	err := repoUser.Delete(1)
	if err != nil {
		t.Error(err)
	}
	u, err := repoUser.GetByID(1)
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error(err)
		}
	}
	if u != nil {
		t.Error("user is not nil")
	}
}

func TestUpdate(t *testing.T) {
	config.DB, _ = MockSqlite()
	defer config.DB.Close()

	repoUser := NewRepository[User]()
	user, _ := repoUser.GetByID(1)
	user.FirstName = "admin2"
	user.GroupId = nil // set nil for try sql_update_value:"Group.ID" (this way I simulate json reqeust, because json not send nil values)
	err := repoUser.Update(user)

	if err != nil {
		t.Error(err)
	}
	u, err := repoUser.GetByID(1)
	if err != nil {
		if err != sql.ErrNoRows {
			t.Error(err)
		}
	}
	t.Log("**** ", u.Group.ID)
	if u.FirstName != "admin2" {
		t.Error("user is not updated")
	}
}
