package query

import (
	"github.com/davecgh/go-spew/spew"

	"github.com/marcgwilson/todo/state"

	"fmt"
	"net/url"
	"reflect"
	"testing"
	"time"
)

func TestParseValues(t *testing.T) {
	gt := time.Now().AddDate(0, 0, -2) // Greater than 2 days ago
	lt := time.Now().AddDate(0, 0, 2) // Les than 2 days from now

	gts := gt.Format(time.RFC3339Nano)
	lts := lt.Format(time.RFC3339Nano)
	u := fmt.Sprintf(`http://0.0.0.0:8000/?state=todo&due:gt=%s&due:lt=%s&page=1&count=20`, gts, lts)

	var testURL *url.URL
	var err error
	if testURL, err = url.Parse(u); err != nil {
		t.Fatal(err)
	}

	result, ae := ParseValues(testURL.Query())
	if ae != nil {
		t.Errorf("Unexpected error: %s", spew.Sdump(ae))
	}

	q := result.Query()
	expectedQuery := " WHERE due > ? AND due < ? AND state IN (?) LIMIT ? OFFSET ?;"
	actualQuery := q.Query()
	if expectedQuery != actualQuery {
		t.Errorf("%s != %s", expectedQuery, actualQuery)
	}

	expectedValues := []interface{}{gt.UTC(), lt.UTC(), state.Todo, int64(20), int64(0)}

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

	expectedValues = []interface{}{gt.UTC(), lt.UTC(), state.Todo}
	actualValues = q.Values()

	if !reflect.DeepEqual(expectedValues, actualValues) {
		t.Errorf("%s != %s", spew.Sdump(expectedValues), spew.Sdump(actualValues))
	}
}
