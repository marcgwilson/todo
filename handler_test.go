package main

import (
	"github.com/marcgwilson/todo/apierror"
	"github.com/marcgwilson/todo/query"
	"github.com/marcgwilson/todo/state"

	"github.com/davecgh/go-spew/spew"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

func filterTodoWithURLString(t *testing.T, tm *TodoManager, urlString string) TodoList {
	// TodoList from query with query parameters from url string
	var err error
	var u *url.URL
	var params *query.QueryParams
	var ae *apierror.Error
	var todoList TodoList

	if u, err = url.Parse(urlString); err != nil {
		t.Fatal(err)
	}

	if params, ae = query.ParseValues(u.Query()); ae != nil {
		t.Fatal(ae)
	}

	params.Depaginate()

	if todoList, err = tm.Query(params.Query()); err != nil {
		t.Fatal(err)
	}

	return todoList
}

func TestHandler(t *testing.T) {
	db, err := OpenDB(":memory:")

	if err != nil {
		t.Fatal(err)
	}

	tm := &TodoManager{db, 20}

	todos := GenerateTodos(100, tm)

	t.Logf("Created %d todos", len(todos))

	ts := httptest.NewServer(NewRouter(tm))
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

		cases := []map[string]interface{}{
			map[string]interface{}{
				"desc":  "Created TODO",
				"due":   time.Now(),
				"state": state.Todo,
			},
			map[string]interface{}{
				"desc":  "TODO 2019-11-02",
				"due":   "2019-11-02T12:25:01Z",
				"state": state.Todo,
			},
		}

		for _, tc := range cases {
			if payload, err = json.Marshal(tc); err != nil {
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
				if res.StatusCode != http.StatusCreated {
					t.Logf("body: %s", string(body))
					t.Fatalf("res.StatusCode = %d != %d", res.StatusCode, http.StatusCreated)
				}

				if actual, err = UnmarshalTodo(body); err != nil {
					t.Fatal(err)
				}

				if expected, err = tm.Get(actual.ID); err != nil {
					t.Error(err)
				} else if !expected.Equal(actual) {
					t.Errorf("expected != actual: %#v != %#v", expected, actual)
				}
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
			map[string]interface{}{
				"id":    1000,
				"due":   time.Now(),
				"state": state.Todo,
			},
		}

		errors := []*apierror.Error{
			&apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON",
				Errors: []*apierror.ErrorDetail{
					&apierror.ErrorDetail{Key: "desc", Value: "", Message: "required attribute"},
					&apierror.ErrorDetail{Key: "due", Value: "", Message: "required attribute"},
					&apierror.ErrorDetail{Key: "state", Value: "", Message: "required attribute"},
				},
			},
			&apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON",
				Errors: []*apierror.ErrorDetail{
					&apierror.ErrorDetail{Key: "due", Value: "", Message: "required attribute"},
					&apierror.ErrorDetail{Key: "state", Value: "", Message: "required attribute"},
				},
			},

			&apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON",
				Errors: []*apierror.ErrorDetail{
					&apierror.ErrorDetail{Key: "due", Value: "", Message: "required attribute"},
					&apierror.ErrorDetail{Key: "state", Value: "invalid state", Message: "state must be one of the following: \"todo\", \"in_progress\", \"done\""},
				},
			},

			&apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON",
				Errors: []*apierror.ErrorDetail{
					&apierror.ErrorDetail{Key: "desc", Value: "", Message: "required attribute"},
					&apierror.ErrorDetail{Key: "due", Value: "invalid date", Message: "Does not match format 'rfc3339'"},
					&apierror.ErrorDetail{Key: "state", Value: "invalid state", Message: "state must be one of the following: \"todo\", \"in_progress\", \"done\""},
				},
			},
			&apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON",
				Errors: []*apierror.ErrorDetail{
					&apierror.ErrorDetail{Key: "(root)", Value: float64(1000), Message: "Additional property id is not allowed"},
					&apierror.ErrorDetail{Key: "desc", Value: "", Message: "required attribute"},
				},
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
				actual := &apierror.Error{}
				if err = json.Unmarshal(body, actual); err != nil {
					t.Fatalf("%d: Error unmarshalling body: %s", i, err)
				}

				if !reflect.DeepEqual(errors[i], actual) {
					t.Errorf("actual.Errors: %s", spew.Sdump(actual))
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
				if !expected.Equal(actual) {
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
			"due":   time.Now().UTC().Format(time.RFC3339Nano),
			"state": "in_progress",
		}

		if todo, err = tm.Create(data); err != nil {
			t.Fatalf("Error creating todo: %s", err)
		}

		updateID := todo.ID

		cases := []map[string]interface{}{
			map[string]interface{}{
				"due":   "invalid date",
				"state": "invalid state",
			},
			map[string]interface{}{
				"id":    1000,
				"due":   time.Now(),
				"state": state.Todo,
			},
		}

		errors := []*apierror.Error{
			&apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON",
				Errors: []*apierror.ErrorDetail{
					&apierror.ErrorDetail{Key: "due", Value: "invalid date", Message: "Does not match format 'rfc3339'"},
					&apierror.ErrorDetail{Key: "state", Value: "invalid state", Message: "state must be one of the following: \"todo\", \"in_progress\", \"done\""},
				},
			},
			&apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid JSON",
				Errors: []*apierror.ErrorDetail{
					&apierror.ErrorDetail{Key: "(root)", Value: float64(1000), Message: "Additional property id is not allowed"},
				},
			},
		}

		for i, c := range cases {
			if payload, err = json.Marshal(c); err != nil {
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
				actual := &apierror.Error{}
				if err = json.Unmarshal(body, actual); err != nil {
					t.Fatalf("Error unmarshalling body: %s", err)
				} else {
					if !reflect.DeepEqual(errors[i], actual) {
						t.Errorf("actual.Errors: %s", spew.Sdump(actual))
						t.Logf("expected: %s", spew.Sdump(errors[i]))
					}
				}
			}
		}
	}
}

func testRetrieve(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var err error

		expected, _ := tm.Get(td[0].ID)

		url := fmt.Sprintf("%s/%d/", ts.URL, expected.ID)

		client := ts.Client()
		if res, err = client.Get(url); err != nil {
			t.Fatal(err)
		}

		defer res.Body.Close()
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Fatal(err)
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
		}
		if err != nil {
			t.Fatal(err)
		} else {
			if actual, err := UnmarshalTodo(body); err != nil {
				t.Errorf("UnmarshalTodo: %s", err)
			} else {
				if !expected.Equal(actual) {
					t.Errorf("expected != actual: %#v != %#v", expected, actual)
				}
			}
		}

		url = ts.URL + "/9000/"

		if res, err = client.Get(url); err != nil {
			t.Fatal(err)
		}

		defer res.Body.Close()
		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Fatal(err)
		}

		expectedError := &apierror.Error{
			Code:    http.StatusNotFound,
			Message: "Not found",
		}

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
		}

		if err != nil {
			t.Fatal(err)
		} else {
			actual := &apierror.Error{}
			if err = json.Unmarshal(body, actual); err != nil {
				t.Fatalf("Error unmarshalling body: %s", err)
			}

			if !reflect.DeepEqual(expectedError, actual) {
				t.Error("expected != actual")
				t.Logf("expected: %s", spew.Sdump(expected))
				t.Logf("actual: %s", spew.Sdump(actual))
			}
		}
	}
}

func testList(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var err error
		var todoList TodoList

		if todoList, err = tm.Query(query.All()); err != nil {
			t.Fatal(err)
		}

		client := ts.Client()

		if res, err = client.Get(ts.URL + "/?page=2"); err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Error(err)
		}

		expected := &PaginatedResponse{
			Results: todoList[20:40],
			Previous: "/?page=1",
			Next: "/?page=3",
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
			t.Logf("body: %s", string(body))
		} else {
			actual := &PaginatedResponse{}
			if err = json.Unmarshal(body, actual); err != nil {
				t.Fatal(err)
			}

			if !expected.Equal(actual) {
				t.Errorf("actual: %s", spew.Sdump(actual))
				t.Logf("expected: %s", spew.Sdump(expected))
			}
		}
	}
}

func listFilterState(ts *httptest.Server, tm *TodoManager, td TodoList) func(*testing.T) {
	return func(t *testing.T) {
		var res *http.Response
		var body []byte
		var err error

		testURL := ts.URL + "/?state=todo&state=done&page=2"
		todoList := filterTodoWithURLString(t, tm, testURL)

		client := ts.Client()

		if res, err = client.Get(testURL); err != nil {
			t.Fatal(err)
		}

		body, err = ioutil.ReadAll(res.Body)
		res.Body.Close()
		if err != nil {
			t.Fatal(err)
		}

		expected := &PaginatedResponse{
			Results: todoList[20:40],
			Previous: "/?page=1&state=todo&state=done",
			Next: "/?page=3&state=todo&state=done",
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
			t.Logf("body: %s", string(body))
		} else {
			actual := &PaginatedResponse{}
			if err = json.Unmarshal(body, actual); err != nil {
				t.Fatal(err)
			}

			if !expected.Equal(actual) {
				t.Errorf("actual: %s", spew.Sdump(actual))
				t.Logf("expected: %s", spew.Sdump(expected))
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

		testURL := fmt.Sprintf("%s/?due:gte=%s&due:lte=%s", ts.URL, gte.Format(time.RFC3339Nano), lte.Format(time.RFC3339Nano))
		todoList := filterTodoWithURLString(t, tm, testURL)

		client := ts.Client()
		if res, err = client.Get(testURL); err != nil {
			t.Fatal(err)
		}

		defer res.Body.Close()
		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}

		expected := &PaginatedResponse{
			Results: todoList[:20],
			Previous: "",
			Next: fmt.Sprintf("/?due:gte=%s&due:lte=%s&page=%d", gte.Format(time.RFC3339Nano), lte.Format(time.RFC3339Nano), 2),
		}

		if res.StatusCode != http.StatusOK {
			t.Errorf("res.StatusCode == %d", res.StatusCode)
			t.Logf("body: %s", string(body))
		} else {
			actual := &PaginatedResponse{}
			if err = json.Unmarshal(body, actual); err != nil {
				t.Fatal(err)
			}

			if !expected.Equal(actual) {
				t.Errorf("actual: %s", spew.Sdump(actual))
				t.Logf("expected: %s", spew.Sdump(expected))
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
			"due":   time.Now().UTC().Format(time.RFC3339Nano),
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
		defer res.Body.Close()

		if res.StatusCode != http.StatusNoContent {
			t.Errorf("%d != %d", res.StatusCode, http.StatusNoContent)
		}

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

		defer res.Body.Close()

		if res.StatusCode != http.StatusNotFound {
			t.Errorf("%d != %d", res.StatusCode, http.StatusNotFound)
		}

		expected := &apierror.Error{
			Code:    http.StatusNotFound,
			Message: "Not found",
		}

		if body, err = ioutil.ReadAll(res.Body); err != nil {
			t.Fatal(err)
		} else {
			actual := &apierror.Error{}
			if err = json.Unmarshal(body, actual); err != nil {
				t.Fatalf("Error unmarshalling body: %s", err)
			}

			if !reflect.DeepEqual(expected, actual) {
				t.Error("expected != actual")
				t.Logf("expected: %s", spew.Sdump(expected))
				t.Logf("actual: %s", spew.Sdump(actual))
			}
		}
	}
}
