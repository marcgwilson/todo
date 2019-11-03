package query

import (
	"github.com/davecgh/go-spew/spew"

	"github.com/marcgwilson/todo/state"

	"reflect"
	"net/url"
	"testing"
)

func TestParseValues(t *testing.T) {
	u := `http://0.0.0.0:8000/?state=todo&due:gt=2019-11-02T06:39:42Z&due:lt=2019-11-12T06:39:42Z&page=1&count=20`

	var testURL *url.URL
	var err error
	if testURL, err = url.Parse(u); err != nil {
		t.Fatal(err)
	}

	result := ParseValues(testURL.Query())
	if len(result.Errors()) != 0 {
		t.Errorf("len(Errors) != 0")
	}

	q := result.Query()
	expectedQuery := " WHERE due > ? AND due < ? AND state IN (?) LIMIT ? OFFSET ?;"
	actualQuery := q.Query()
	if expectedQuery != actualQuery {
		t.Errorf("%s != %s", expectedQuery, actualQuery)
	}

	expectedValues := []interface{}{int64(1572676782), int64(1573540782), state.Todo, int64(20), int64(0)}
	actualValues := q.Values()

	if !reflect.DeepEqual(expectedValues, actualValues) {
		t.Errorf("%s != %s", spew.Sdump(expectedValues), spew.Sdump(actualValues))
	}

	rc := result.ShallowCopy().Depaginate()
	q = rc.Query()
	expectedQuery = " WHERE due > ? AND due < ? AND state IN (?);"
	actualQuery = q.Query()
	if expectedQuery != actualQuery {
		t.Errorf("%s != %s", expectedQuery, actualQuery)
	}

	expectedValues = []interface{}{int64(1572676782), int64(1573540782), state.Todo}
	actualValues = q.Values()

	if !reflect.DeepEqual(expectedValues, actualValues) {
		t.Errorf("%s != %s", spew.Sdump(expectedValues), spew.Sdump(actualValues))
	}
}