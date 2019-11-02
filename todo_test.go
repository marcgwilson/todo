package main

import (
	"github.com/marcgwilson/todo/state"

	"encoding/json"
	"testing"
	"time"
)

func TestTodoManager(t *testing.T) {
	db, err := OpenDB(":memory:")
	if err != nil {
		t.Fatal(err)
	}

	defer db.Close()

	tm := &TodoManager{db, 20}

	expected := &Todo{
		ID:          0,
		Description: "todo 1",
		Due:         time.Now().UTC(),
		State:       state.Todo,
	}

	data := map[string]interface{}{
		"desc":  expected.Description,
		"due":   expected.Due.Format(time.RFC3339),
		"state": expected.State,
	}

	var actual *Todo

	if actual, err = tm.Create(data); err != nil {
		t.Error(err)
	} else {
		expected.ID = actual.ID

		if !expected.Equals(actual) {
			t.Errorf("expected != actual")
		}
	}

	expected2 := &Todo{
		ID:          0,
		Description: "todo 2",
		Due:         time.Now().UTC(),
		State:       state.Todo,
	}

	data2 := map[string]interface{}{
		"desc":  expected2.Description,
		"due":   expected2.Due.Format(time.RFC3339),
		"state": expected2.State,
	}

	var actual2 *Todo

	if actual2, err = tm.Create(data2); err != nil {
		t.Error(err)
	} else {
		expected2.ID = actual2.ID

		if !expected2.Equals(actual2) {
			t.Errorf("expected != actual")
		}
	}

	var tds []*Todo

	if tds, err = tm.List(nil); err != nil {
		t.Error(err)
	} else {
		for i, a := range tds {
			t.Logf("tds[%d]=%#v", i, a)
		}
	}

	expected3 := &Todo{
		ID:          0,
		Description: "todo 3",
		Due:         time.Now().UTC(),
		State:       state.Todo,
	}

	data3 := map[string]interface{}{
		"desc":  expected3.Description,
		"due":   expected3.Due.Format(time.RFC3339),
		"state": expected3.State,
	}

	var actual3 *Todo

	if actual3, err = tm.Create(data3); err != nil {
		t.Error(err)
	} else {
		expected3.ID = actual3.ID

		if !expected3.Equals(actual3) {
			t.Errorf("expected != actual")
		}
	}

	if tds, err = tm.List(nil); err != nil {
		t.Error(err)
	} else {
		for i, a := range tds {
			t.Logf("tds[%d]=%#v", i, a)
		}
	}

	if err = tm.Delete(actual2.ID); err != nil {
		t.Error(err)
	}

	if tds, err = tm.List(nil); err != nil {
		t.Error(err)
	} else {
		for i, a := range tds {
			t.Logf("tds[%d]=%#v", i, a)
		}
	}
}

func TestTodo(t *testing.T) {
	expected := &Todo{
		ID:          1,
		Description: "My Todo",
		Due:         time.Now(),
		State:       state.Todo,
	}

	if rawBytes, err := expected.Marshal(); err != nil {
		t.Errorf("actual.Marshal: %s", err)
	} else {
		actual := &Todo{}
		if err := json.Unmarshal(rawBytes, actual); err != nil {
			t.Errorf("json.Unmarshal: %s", err)
		} else if !expected.Equals(actual) {
			t.Logf("expected != actual: %#v", actual)
		}
	}
}
