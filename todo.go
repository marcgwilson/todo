package main

import (
	"github.com/marcgwilson/todo/query"
	"github.com/marcgwilson/todo/state"

	_ "github.com/mattn/go-sqlite3"

	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var CreateStmt = `CREATE TABLE IF NOT EXISTS todo (
    desc TEXT,
    due INTEGER,
    state TEXT
);`

type Todo struct {
	ID          int64       `db:"id" json:"id"`
	Description string      `db:"desc" json:"desc"`
	Due         time.Time   `db:"due" json:"due"`
	State       state.State `db:"state" json:"state"`
}

func (r *Todo) Equals(t *Todo) bool {
	if r.ID != t.ID {
		return false
	}

	if r.Description != t.Description {
		return false
	}

	if r.Due.Unix() != t.Due.Unix() {
		return false
	}

	if r.State != t.State {
		return false
	}

	return true
}

type TodoList []*Todo

func (r TodoList) Equals(t TodoList) bool {
	if len(r) != len(t) {
		return false
	}

	for i, elem := range r {
		if !elem.Equals(t[i]) {
			return false
		}
	}

	return true
}

func (r *Todo) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalTodo(b []byte) (*Todo, error) {
	t := &Todo{}
	if err := json.Unmarshal(b, t); err != nil {
		return nil, err
	}
	return t, nil
}

func NewTodo(desc string, due time.Time, s state.State) *Todo {
	return &Todo{-1, desc, due, s}
}

func UnmarshalTodoList(b []byte) (TodoList, error) {
	t := TodoList{}
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return t, nil
}

type TodoManager struct {
	Database *sql.DB
	Limit    int
}

func NewManager(db *sql.DB, limit int) *TodoManager {
	return &TodoManager{db, limit}
}

func (r *TodoManager) Get(id int64) (*Todo, error) {
	var stmt *sql.Stmt
	var row *sql.Row
	var err error

	if stmt, err = r.Database.Prepare("SELECT rowid, desc, due, state FROM todo WHERE rowid = ?"); err != nil {
		return nil, err
	}
	defer stmt.Close()

	t := &Todo{}
	row = stmt.QueryRow(id)

	var due int64
	err = row.Scan(&t.ID, &t.Description, &due, &t.State)
	if err != nil {
		return nil, err
	} else {
		t.Due = time.Unix(due, 0)
		return t, nil
	}
}

func (r *TodoManager) Query(filter *query.Query) (TodoList, error) {
	var stmt *sql.Stmt
	var rows *sql.Rows
	var err error

	q := "SELECT rowid, desc, due, state FROM todo" + filter.Query()

	log.Println(q)
	log.Printf("%#v\n", filter.Values())

	if stmt, err = r.Database.Prepare(q); err != nil {
		return nil, err
	}
	defer stmt.Close()

	if rows, err = stmt.Query(filter.Values()...); err != nil {
		return nil, err
	}

	defer rows.Close()

	results := []*Todo{}

	for rows.Next() {
		t := &Todo{}
		var d int64
		if err = rows.Scan(&t.ID, &t.Description, &d, &t.State); err != nil {
			return nil, err
		}

		t.Due = time.Unix(d, 0)
		results = append(results, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func (r *TodoManager) Count(filter *query.Query) (int64, error) {
	var stmt *sql.Stmt
	var err error
	var count int64

	q := "SELECT COUNT(*) FROM todo" + filter.Query()

	log.Println(q)
	log.Printf("%#v\n", filter.Values())

	if stmt, err = r.Database.Prepare(q); err != nil {
		return 0, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(filter.Values()...)

	err = row.Scan(&count)
	return count, err
}

func (r *TodoManager) Create(data map[string]interface{}) (*Todo, error) {
	var err error
	var id int64
	var result sql.Result
	var tx *sql.Tx
	var stmt *sql.Stmt

	var d TodoMap

	if d, err = Transform(data); err != nil {
		return nil, err
	}

	insert := d.InsertVars()
	sql := fmt.Sprintf("INSERT INTO todo(%s) VALUES(%s)", insert.Names, insert.Bindvars)

	if tx, err = r.Database.Begin(); err != nil {
		return nil, err
	}

	if stmt, err = tx.Prepare(sql); err != nil {
		return nil, err
	}

	defer stmt.Close()

	if result, err = stmt.Exec(insert.Values...); err != nil {
		return nil, err
	}

	if id, err = result.LastInsertId(); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &Todo{id, d.Description(), d.Due(), d.State()}, nil
}

func (r *TodoManager) Update(id int64, data map[string]interface{}) (*Todo, error) {
	var err error
	var todo *Todo
	var tx *sql.Tx
	var stmt *sql.Stmt

	var d TodoMap

	if d, err = Transform(data); err != nil {
		return nil, err
	}

	update := d.UpdateVars()
	query := fmt.Sprintf("UPDATE todo SET %s WHERE rowid = ?;", update.Bindvars)

	if tx, err = r.Database.Begin(); err != nil {
		return nil, err
	}

	if stmt, err = tx.Prepare(query); err != nil {
		return nil, err
	}

	defer stmt.Close()

	values := update.Values
	values = append(values, id)
	if _, err = stmt.Exec(values...); err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	if todo, err = r.Get(id); err != nil {
		return nil, err
	}

	return todo, nil
}

func (r *TodoManager) Delete(id int64) error {
	var err error

	var tx *sql.Tx
	var stmt *sql.Stmt

	if tx, err = r.Database.Begin(); err != nil {
		return err
	}

	if stmt, err = tx.Prepare("DELETE FROM todo WHERE rowid=?;"); err != nil {
		return err
	}

	if _, err = stmt.Exec(id); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
