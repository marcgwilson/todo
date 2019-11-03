package query

import (
	"reflect"
	"net/url"
	"testing"
)

func TestParseQuery(t *testing.T) {
	u := `http://0.0.0.0:8000/?state=todo&due:gt=2019-11-02T06:39:42Z&due:lt=2019-11-12T06:39:42Z&page=1&count=20`

	var testURL *url.URL
	var err error
	if testURL, err = url.Parse(u); err != nil {
		t.Fatal(err)
	}

	result := ParseQuery(testURL.Query())
	if len(result.Errors()) != 0 {
		t.Errorf("len(Errors) != 0")
	}

	expectedQuery := "WHERE due > ? AND due < ? AND state IN (?) LIMIT ? OFFSET ?;"
	actualQuery := result.Query()
	if expectedQuery != actualQuery {
		t.Errorf("%s != %s", expectedQuery, actualQuery)
	}

	expectedValues := []interface{}{int64(1572676782), int64(1573540782), "\"todo\"", int64(20), int64(20)}
	actualValues := result.Values()

	if !reflect.DeepEqual(expectedValues, actualValues) {
		t.Errorf("%#v != %#v", expectedValues, actualValues)
	}
}
