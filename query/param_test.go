package query

import (
	"net/url"
	"strings"
	"testing"
)

func TestParam(t *testing.T) {
	u := `http://0.0.0.0:8000/?state=todo&due:gt=2019-11-02T06:39:42Z&due:lt=2019-11-12T06:39:42Z&page=1&count=20`

	var testURL *url.URL
	var err error
	if testURL, err = url.Parse(u); err != nil {
		t.Fatal(err)
	}

	query := testURL.Query()

	queryParams := map[string]IQueryParam{}

	for key, values := range query {
		if parser, ok := parserMap[key]; ok {
			if parsed, err := parser(values); err != nil {
				t.Logf("Unrecognized key: %s", key)
			} else {
				queryParams[key] = parsed

				name := parsed.Name()
				values := parsed.Values()
				t.Logf("name: %s", name)
				t.Logf("values: %#v", values)
			}
		}
	}

	// LIMIT QUERY:
	var limit *CountQueryParam
	if val, ok := queryParams["count"]; !ok {
		limit = &CountQueryParam{"count", []interface{}{int64(20)}}
	} else {
		limit = val.(*CountQueryParam)
		delete(queryParams, "count")
	}

	// OFFSET QUERY:
	var offset *PageQueryParam
	if val, ok := queryParams["page"]; !ok {
		offset = &PageQueryParam{"page", []interface{}{int64(0)}, limit.Count()}
	} else {
		offset = val.(*PageQueryParam)
		offset.SetCount(limit.Count())
		delete(queryParams, "page")
	}

	// query := ""
	queryFragments := []string{}
	values := []interface{}{}

	for _, value := range queryParams {
		queryFragments = append(queryFragments, value.Name())
		values = append(values, value.Values()...)
	}

	dbQuery := strings.Join(queryFragments, " AND ") + " " + offset.Name() + " " + limit.Name() + ";"
	values = append(values, offset.Values()...)
	values = append(values, limit.Values()...)
	t.Logf("dbQuery: %s", dbQuery)
	t.Logf("values: %#v\n", values)
}
