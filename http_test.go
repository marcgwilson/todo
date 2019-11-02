package main

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/marcgwilson/todo/state"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func TestHTTP(t *testing.T) {
	db, err := OpenDB(":memory:")

	if err != nil {
		t.Fatal(err)
	}

	tm := &TodoManager{db, 20}

	todos := GenerateTodos(100, tm)

	t.Logf("Created %d todos", len(todos))

	ts := httptest.NewServer(NewMux(tm))
	defer ts.Close()

	t.Run("CREATE", testCreate(ts, tm, todos))
	t.Run("CREATE-ERRORS", testCreateErrors(ts, tm, todos))
	t.Run("UPDATE", testUpdate(ts, tm, todos))
	t.Run("UPDATE-ERRORS", testUpdateErrors(ts, tm, todos))
	t.Run("RETRIEVE", testRetrieve(ts, tm, todos))
	t.Run("LIST=all", testList(ts, tm, todos))
	t.Run("LIST=state", listFilterState(ts, tm, todos))
	t.Run("LIST=due", listFilterDue(ts, tm, todos))
	t.Run("DELETE", testDelete(ts, tm, todos))
	t.Run("DELETE-ERRORS", testDeleteErrors(ts, tm, todos))
}

func testCreate(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var payload []byte
		var err error
		var actual *Todo
		var expected *Todo

		client := ts.Client()

		p := map[string]interface{}{
			"desc":  "Created TODO",
			"due":   time.Now(),
			"state": state.Todo,
		}

		if payload, err = json.Marshal(p); err != nil {
			t.Fatal(err)
		}

		if res, err = client.Post(ts.URL, "application/json", bytes.NewBuffer(payload)); err != nil {
			t.Fatal(err)
		}

		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		if err != nil {
			t.Fatal(err)
		} else {
			if actual, err = UnmarshalTodo(body); err != nil {
				t.Fatal(err)
			}

			if expected, err = tm.Get(actual.ID); err != nil {
				t.Error(err)
			} else if !expected.Equals(actual) {
				t.Errorf("expected != actual: %#v != %#v", expected, actual)
			}
		}
	}
}

func testCreateErrors(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var payload []byte
		var err error

		client := ts.Client()

		cases := []map[string]interface{}{
			map[string]interface{}{},
			map[string]interface{}{
				"desc": "Created TODO",
			},
			map[string]interface{}{
				"desc":  "Created TODO",
				"state": "invalid state",
			},
			map[string]interface{}{
				"due":   "invalid date",
				"state": "invalid state",
			},
		}

		errors := [][]*ValidationError{
			[]*ValidationError{
				&ValidationError{Key: "desc", Value: "", Message: "required attribute"},
				&ValidationError{Key: "due", Value: "", Message: "required attribute"},
				&ValidationError{Key: "state", Value: "", Message: "required attribute"},
			},
			[]*ValidationError{
				&ValidationError{Key: "due", Value: "", Message: "required attribute"},
				&ValidationError{Key: "state", Value: "", Message: "required attribute"},
			},

			[]*ValidationError{
				&ValidationError{Key: "due", Value: "", Message: "required attribute"},
				&ValidationError{Key: "state", Value: "invalid state", Message: "state must be one of the following: \"todo\", \"in_progress\", \"done\""},
			},
			[]*ValidationError{
				&ValidationError{Key: "desc", Value: "", Message: "required attribute"},
				&ValidationError{Key: "due", Value: "invalid date", Message: "Does not match format 'date-time'"},
				&ValidationError{Key: "state", Value: "invalid state", Message: "state must be one of the following: \"todo\", \"in_progress\", \"done\""},
			},
		}

		for i, c := range cases {
			if payload, err = json.Marshal(c); err != nil {
				t.Fatal(err)
			}

			if res, err = client.Post(ts.URL, "application/json", bytes.NewBuffer(payload)); err != nil {
				t.Fatal(err)
			}

			body, err = ioutil.ReadAll(res.Body)
			res.Body.Close()

			if res.StatusCode != http.StatusBadRequest {
				t.Fatalf("%d: statusCode = %d != %d", i, res.StatusCode, http.StatusBadRequest)
			} else {
				actual := &APIValidationError{}
				if err = json.Unmarshal(body, &actual); err != nil {
					t.Fatalf("%d: Error unmarshalling body: %s", i, err)
				}

				if !reflect.DeepEqual(errors[i], actual.Errors) {
					t.Errorf("actual.Errors: %s", spew.Sdump(actual.Errors))
					t.Logf("expected: %s", spew.Sdump(errors[i]))
				}
			}
		}
	}
}

func testUpdate(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var req *http.Request
		var res *http.Response

		var actual *Todo
		var body []byte
		var payload []byte
		var err error

		expected := &Todo{
			ID:          td[0].ID,
			Description: "Updated Todo!",
			Due:         time.Now(),
			State:       state.Done,
		}

		p := map[string]interface{}{
			"desc":  expected.Description,
			"due":   expected.Due,
			"state": expected.State,
		}

		if payload, err = json.Marshal(p); err != nil {
			t.Fatal(err)
		}

		url := fmt.Sprintf("%s/%d/", ts.URL, expected.ID)

		client := ts.Client()

		req, err = http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		if res, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("statusCode = %d != %d", res.StatusCode, http.StatusOK)
		} else {
			if actual, err = UnmarshalTodo(body); err != nil {
				t.Fatal(err)
			} else {
				if !expected.Equals(actual) {
					t.Errorf("expected != actual")
				}
			}
		}
	}
}

func testUpdateErrors(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var req *http.Request
		var res *http.Response

		var todo *Todo
		var body []byte
		var payload []byte
		var err error

		data := map[string]interface{}{
			"desc":  "DELETE Todo",
			"due":   time.Now().UTC().Format(time.RFC3339),
			"state": "in_progress",
		}

		if todo, err = tm.Create(data); err != nil {
			t.Fatalf("Error creating todo: %s", err)
		}

		updateID := todo.ID

		p := map[string]interface{}{
			"due":   "invalid date",
			"state": "invalid state",
		}

		expectedErrors := []*ValidationError{
			&ValidationError{Key: "due", Value: "invalid date", Message: "Does not match format 'date-time'"},
			&ValidationError{Key: "state", Value: "invalid state", Message: "state must be one of the following: \"todo\", \"in_progress\", \"done\""},
		}

		if payload, err = json.Marshal(p); err != nil {
			t.Fatal(err)
		}

		url := fmt.Sprintf("%s/%d/", ts.URL, updateID)

		client := ts.Client()

		req, err = http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
		if err != nil {
			t.Fatal(err)
		}

		if res, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode != http.StatusBadRequest {
			t.Fatalf("statusCode = %d != %d", res.StatusCode, http.StatusBadRequest)
		} else {
			actual := &APIValidationError{}
			if err = json.Unmarshal(body, &actual); err != nil {
				t.Fatalf("Error unmarshalling body: %s", err)
			} else {
				if !reflect.DeepEqual(expectedErrors, actual.Errors) {
					t.Errorf("actual.Errors: %s", spew.Sdump(actual.Errors))
					t.Logf("expected: %s", spew.Sdump(expectedErrors))
				}
			}
		}
	}
}

func testRetrieve(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var err error

		expected, _ := tm.Get(td[0].ID)

		url := fmt.Sprintf("%s/%d/", ts.URL, expected.ID)

		client := ts.Client()
		if res, err = client.Get(url); err != nil {
			t.Fatal(err)
		}

		var body []byte
		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
		}
		if err != nil {
			t.Fatal(err)
		} else {
			if actual, err := UnmarshalTodo(body); err != nil {
				t.Errorf("UnmarshalTodo: %s", err)
			} else {
				if !expected.Equals(actual) {
					t.Errorf("expected != actual: %#v != %#v", expected, actual)
				}
			}
		}

		url = ts.URL + "/9000/"

		if res, err = client.Get(url); err != nil {
			t.Fatal(err)
		}

		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		expectedError := []*APIError{
			&APIError{
				Code:    http.StatusNotFound,
				Message: "Not found",
			},
		}

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
		}

		if err != nil {
			t.Fatal(err)
		} else {
			actual := []*APIError{}
			if err = json.Unmarshal(body, &actual); err != nil {
				t.Fatalf("Error unmarshalling body: %s", err)
			}

			if !reflect.DeepEqual(expectedError, actual) {
				for _, elem := range actual {
					t.Logf("%#v", elem)
				}
				t.Error("expected != actual")
			}
		}
	}
}

func testList(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var err error

		all, _ := tm.List(nil)

		expected := all[60:80]

		client := ts.Client()

		if res, err = client.Get(ts.URL + "/?page=4"); err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Error(err)
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
			t.Logf("body: %s", string(body))
		} else {
			result := &PaginatedResult{}
			if err = json.Unmarshal(body, result); err != nil {
				t.Fatal(err)
			}

			t.Logf("next: %s\nprevious: %s\n", result.Next, result.Previous)
			if !expected.Equals(result.Results) {
				t.Errorf("slices not equal len(actual) = %d, len(expected) = %d", len(result.Results), len(expected))
			}
			if result.Next != "page=5" {
				t.Errorf("result.Next=%s", result.Next)
			}

			if result.Previous != "page=3" {
				t.Errorf("result.Previous=%s", result.Previous)
			}
		}
	}
}

func listFilterState(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var err error

		client := ts.Client()

		if res, err = client.Get(ts.URL + "/?state=todo&state=done"); err != nil {
			t.Fatal(err)
		}

		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
			t.Logf("body: %s", string(body))
		} else {

			result := &PaginatedResult{}
			if err = json.Unmarshal(body, result); err != nil {
				t.Fatal(err)
			}

			t.Logf("next: %s\nprevious: %s\n", result.Next, result.Previous)
			if len(result.Results) != 20 {
				t.Errorf("len(result.Results = %d", len(result.Results))
			}

			if result.Next != "page=2&state=todo&state=done" {
				t.Errorf("result.Next=%s", result.Next)
			}

			if result.Previous != "" {
				t.Errorf("result.Previous=%s", result.Previous)
			}
		}
	}
}

func listFilterDue(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var err error

		now := time.Now()
		gte := now.Add(time.Second * -10)
		lte := now.Add(time.Second * 10)

		filter := "/?due:gt=" + gte.Format(time.RFC3339) + "&due:lt=" + lte.Format(time.RFC3339)

		t.Logf("filter: %s", filter)

		client := ts.Client()
		if res, err = client.Get(ts.URL + filter); err != nil {
			t.Fatal(err)
		}

		defer res.Body.Close()
		body, err = ioutil.ReadAll(res.Body)

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
		} else {

			result := &PaginatedResult{}
			if err = json.Unmarshal(body, result); err != nil {
				t.Fatal(err)
			}

			t.Logf("next: %s\nprevious: %s\n", result.Next, result.Previous)
			if len(result.Results) != 20 {
				t.Errorf("len(result.Results = %d", len(result.Results))
			}

			if result.Next == "" {
				t.Errorf("result.Next=%q", result.Next)
			}

			if result.Previous != "" {
				t.Errorf("result.Previous=%s", result.Previous)
			}
		}
	}
}

func testDelete(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var todo *Todo
		var req *http.Request
		var res *http.Response
		var body []byte
		var err error

		data := map[string]interface{}{
			"desc":  "DELETE Todo",
			"due":   time.Now().UTC().Format(time.RFC3339),
			"state": "in_progress",
		}

		if todo, err = tm.Create(data); err != nil {
			t.Fatalf("Error creating todo: %s", err)
		}

		todoID := todo.ID

		url := ts.URL + fmt.Sprintf("/%d/", todo.ID)

		client := ts.Client()

		if req, err = http.NewRequest("DELETE", url, nil); err != nil {
			t.Fatal(err)
		}

		if res, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != http.StatusNoContent {
			t.Errorf("%d != %d", res.StatusCode, http.StatusNoContent)
		}

		defer res.Body.Close()
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Fatal(err)
		} else {
			if len(body) != 0 {
				t.Error("len(body) > 0")
			}
		}

		_, err = tm.Get(todoID)

		if err.Error() != "sql: no rows in result set" {
			t.Errorf("err != sql: no rows in result set")
		}
	}
}

func testDeleteErrors(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var req *http.Request
		var res *http.Response
		var body []byte
		var err error

		url := ts.URL + "/9000/"

		client := ts.Client()

		if req, err = http.NewRequest("DELETE", url, nil); err != nil {
			t.Fatal(err)
		}

		if res, err = client.Do(req); err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("%d != %d", res.StatusCode, http.StatusNotFound)
		}

		expected := []*APIError{
			&APIError{
				Code:    http.StatusNotFound,
				Message: "Not found",
			},
		}

		defer res.Body.Close()
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Fatal(err)
		} else {
			actual := []*APIError{}
			if err = json.Unmarshal(body, &actual); err != nil {
				t.Fatalf("Error unmarshalling body: %s", err)
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Error("expected != actual")
			}
		}
	}
}
