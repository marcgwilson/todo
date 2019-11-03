package query

import (
	"fmt"
	"net/url"
	"strings"
	"sort"
	"strconv"
)

func All() *ParseResult {
	return &ParseResult{"", []interface{}{}, []error{}, &CountQueryParam{}, &PageQueryParam{}}
}

type ParseResult struct {
	query string
	values []interface{}
	errors []error
	limit *CountQueryParam
	offset *PageQueryParam
}

func (r *ParseResult) Query() string {
	return r.query
}

func (r *ParseResult) Values() []interface{} {
	return r.values
}

func (r *ParseResult) Errors() []error {
	if len(r.errors) == 0 {
		return nil
	}
	return r.errors
}

func (r *ParseResult) NextPage(u *url.URL, count int64) string {
	if r.offset.Offset() + r.limit.Count() < count {
		u, _ := url.Parse(u.String())
		values := u.Query()
		values.Set("page", strconv.FormatInt(r.offset.Page() + 1, 10))
		queryString, _ := url.QueryUnescape(values.Encode())
		u.RawQuery = queryString
		return u.String()
	}
	return ""
}

func (r *ParseResult) PrevPage(u *url.URL) string {
	if r.offset.Page() > 1 {
		u, _ := url.Parse(u.String())
		values := u.Query()
		values.Set("page", strconv.FormatInt(r.offset.Page() - 1, 10))
		queryString, _ := url.QueryUnescape(values.Encode())
		u.RawQuery = queryString
		return u.String()
	}
	return ""
}

func ParseQuery(query url.Values) *ParseResult {
	queryParams := map[string]IQueryParam{}
	errors := []error{}

	for key, values := range query {
		if pm, ok := parserMap[key]; ok {

			a, b := pm(values)
			if len(b) == 0 {
				queryParams[key] = a
			} else {
				errors = append(errors, b...)
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
		offset = &PageQueryParam{"page", []interface{}{int64(1)}, limit.Count()}
	} else {
		offset = val.(*PageQueryParam)
		offset.SetCount(limit.Count())
		delete(queryParams, "page")
	}

	queryFragments := []string{}
	values := []interface{}{}

	keys := make([]string, len(queryParams))
	i := 0
	for k := range queryParams {
	    keys[i] = k
	    i++
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := queryParams[key]
		queryFragments = append(queryFragments, value.Name())
		values = append(values, value.Values()...)
	}

	queryString := strings.Join(queryFragments, " AND ")
	if len(queryString) > 0 {
		queryString = fmt.Sprintf("WHERE %s ", queryString)
	}

	queryString = queryString + limit.Name() + " " + offset.Name() + ";"
	values = append(values, limit.Values()...)
	values = append(values, offset.Values()...)
	
	return &ParseResult{queryString, values, errors, limit, offset}
}
