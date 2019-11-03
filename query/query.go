package query

import (
	"fmt"
	"net/url"
	"strings"
	"sort"
)

func All() *Query {
	return &Query{";", []interface{}{}}
}

type QueryParams struct {
	params map[string]IQueryParam
	errors []error
}

func (r *QueryParams) Offset() *PageQueryParam {
	if val, ok := r.params["page"]; ok {
		return val.(*PageQueryParam)
	}
	return nil
}

func (r *QueryParams) Limit() *CountQueryParam {
	if val, ok := r.params["count"]; ok {
		return val.(*CountQueryParam)
	}
	return nil
}

func (r *QueryParams) Params() map[string]IQueryParam {
	return r.params
}

func (r *QueryParams) Errors() []error {
	if len(r.errors) == 0 {
		return nil
	}
	return r.errors
}

func (r *QueryParams) Paginate() *QueryParams {
	var limit *CountQueryParam

	if val, ok := r.params["count"]; !ok {
		limit = &CountQueryParam{"count", []interface{}{int64(20)}}
		r.params["count"] = limit
	} else {
		limit = val.(*CountQueryParam)
	}

	var offset *PageQueryParam
	if val, ok := r.params["page"]; !ok {
		offset = &PageQueryParam{"page", []interface{}{int64(1)}, limit.Count()}
		r.params["page"] = offset
	} else {
		offset = val.(*PageQueryParam)
		offset.SetCount(limit.Count())
	}

	return r
}

func (r *QueryParams) Depaginate() *QueryParams {
	delete(r.params, "count")
	delete(r.params, "page")
	return r
}

func (r *QueryParams) ShallowCopy() *QueryParams {
	params := map[string]IQueryParam{}
	for k,v := range r.params {
	  params[k] = v
	}

	errors := make([]error, len(r.errors), len(r.errors))
	for i, err := range r.errors {
		errors[i] = err
	}

	return &QueryParams{params, errors}
}

func (r *QueryParams) Query() *Query {
	// OFFSET QUERY:
	var offset *PageQueryParam
	if val, ok := r.params["page"]; ok {
		offset = val.(*PageQueryParam)
	}

	// LIMIT QUERY:
	var limit *CountQueryParam

	if val, ok := r.params["count"]; ok {
		limit = val.(*CountQueryParam)
	} else {
		if offset != nil {
			limit = &CountQueryParam{"count", []interface{}{int64(20)}}
			offset.SetCount(limit.Count())
		}
	}

	queryFragments := []string{}
	values := []interface{}{}

	keys := []string{}
	for k := range r.params {
		if k != "page" && k != "count" {
			keys = append(keys, k)
		}
	}

	sort.Strings(keys)

	for _, key := range keys {
		value := r.params[key]
		queryFragments = append(queryFragments, value.Name())
		values = append(values, value.Values()...)
	}

	queryString := strings.Join(queryFragments, " AND ")
	if len(queryString) > 0 {
		queryString = fmt.Sprintf(" WHERE %s", queryString)
	}

	if limit != nil {
		queryString = queryString + " " + limit.Name()
		values = append(values, limit.Values()...)
	}

	if offset != nil {
		queryString = queryString + " " + offset.Name()
		values = append(values, offset.Values()...)
	}

	queryString = queryString + ";"

	return &Query{queryString, values}
}

type Query struct {
	query string
	values []interface{}
}

func (r *Query) Query() string {
	return r.query
}

func (r *Query) Values() []interface{} {
	return r.values
}

func ParseValues(query url.Values) *QueryParams {
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
	return &QueryParams{queryParams, errors}
}
