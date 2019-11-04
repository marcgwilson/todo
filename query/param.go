package query

import (
	"github.com/marcgwilson/todo/apierror"
	"github.com/marcgwilson/todo/state"

	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	TimeErrorMessage  = "value must be in RFC-3339 format"
	PageErrorMessage  = "value must be an integer greater than 0"
	CountErrorMessage = "value must be an integer greater than 0"
)

var parserMap = map[string]ParamListParser{
	"due:gt":  DueDateParser("due >"),
	"due:lt":  DueDateParser("due <"),
	"due:gte": DueDateParser("due >="),
	"due:lte": DueDateParser("due <="),
	"due":     DueDateParser("due ="),
	"state":   StateParser("state"),
	"page":    PageParser("page"),
	"count":   CountParser("count"),
}

type IQueryParam interface {
	Name() string
	Values() []interface{}
}

type ParamListParser func([]string) (IQueryParam, *apierror.Error)

type DueDateQueryParam struct {
	name   string
	values []interface{}
}

func (r *DueDateQueryParam) Name() string {
	results := make([]string, len(r.values), len(r.values))
	for i := 0; i < len(r.values); i++ {
		results[i] = fmt.Sprintf("%s ?", r.name)
	}
	return strings.Join(results, " AND ")
}

func (r *DueDateQueryParam) Values() []interface{} {
	return r.values
}

func DueDateParser(param string) ParamListParser {
	return func(values []string) (IQueryParam, *apierror.Error) {
		var ae *apierror.Error
		results := []interface{}{}
		errors := []*apierror.ErrorDetail{}

		for _, value := range values {
			if t, err := time.Parse(time.RFC3339Nano, value); err != nil {
				errors = append(errors, &apierror.ErrorDetail{Key: param, Value: value, Message: TimeErrorMessage})
			} else {
				results = append(results, t.UTC()) //t.UnixNano()) //t.Unix())
			}
		}

		if len(errors) > 0 {
			ae = &apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid query parameters",
				Errors:  errors,
			}
		}
		return &DueDateQueryParam{param, results}, ae
	}
}

type StateQueryParam struct {
	name   string
	values []interface{}
}

func (r *StateQueryParam) Name() string {
	results := make([]string, len(r.values), len(r.values))
	for i := 0; i < len(r.values); i++ {
		results[i] = "?"
	}

	return fmt.Sprintf("%s IN (%s)", r.name, strings.Join(results, ", "))
}

func (r *StateQueryParam) Values() []interface{} {
	return r.values
}

func StateParser(param string) ParamListParser {
	return func(values []string) (IQueryParam, *apierror.Error) {
		var ae *apierror.Error
		results := []interface{}{}
		errors := []*apierror.ErrorDetail{}

		for _, value := range values {
			if val, ok := state.States[value]; !ok {
				errors = append(errors, &apierror.ErrorDetail{Key: "state", Value: value, Message: "invalid state"})
			} else {
				results = append(results, val)
			}
		}

		if len(errors) > 0 {
			ae = &apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid query parameters",
				Errors:  errors,
			}
		}

		return &StateQueryParam{"state", results}, ae
	}
}

type PageQueryParam struct {
	name   string
	values []interface{}
	count  int64
}

func (r *PageQueryParam) SetCount(c int64) {
	r.count = c
}

func (r *PageQueryParam) Page() int64 {
	v := int64(1)
	if len(r.values) > 0 {
		v = r.values[0].(int64)
	}
	return v
}

func (r *PageQueryParam) Offset() int64 {
	return (r.Page() - 1) * r.count
}

func (r *PageQueryParam) Name() string {
	return "OFFSET ?"
}

func (r *PageQueryParam) Values() []interface{} {
	return []interface{}{r.Offset()}
}

func PageParser(param string) ParamListParser {
	return func(values []string) (IQueryParam, *apierror.Error) {
		var ae *apierror.Error
		results := []interface{}{}
		errors := []*apierror.ErrorDetail{}

		if result, err := strconv.ParseInt(values[0], 10, 64); err != nil {
			errors = append(errors, &apierror.ErrorDetail{Key: "page", Value: values[0], Message: PageErrorMessage})
		} else {
			if result < 1 {
				errors = append(errors, &apierror.ErrorDetail{Key: "page", Value: values[0], Message: PageErrorMessage})
			} else {
				results = append(results, result)
			}
		}

		if len(errors) > 0 {
			ae = &apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid query parameters",
				Errors:  errors,
			}
		}

		return &PageQueryParam{"page", results, 0}, ae
	}
}

type CountQueryParam struct {
	name   string
	values []interface{}
}

func (r *CountQueryParam) Count() int64 {
	if len(r.values) > 0 {
		return r.values[0].(int64)
	} else {
		return int64(0)
	}
}

func (r *CountQueryParam) Name() string {
	return "LIMIT ?"
}

func (r *CountQueryParam) Values() []interface{} {
	return r.values
}

func CountParser(param string) ParamListParser {
	return func(values []string) (IQueryParam, *apierror.Error) {
		var ae *apierror.Error
		results := []interface{}{}
		errors := []*apierror.ErrorDetail{}

		if result, err := strconv.ParseInt(values[0], 10, 64); err != nil {
			errors = append(errors, &apierror.ErrorDetail{Key: "count", Value: values[0], Message: CountErrorMessage})
		} else {
			if result < 1 {
				errors = append(errors, &apierror.ErrorDetail{Key: "count", Value: values[0], Message: CountErrorMessage})
			} else {
				results = append(results, result)
			}
		}

		if len(errors) > 0 {
			ae = &apierror.Error{
				Code:    http.StatusBadRequest,
				Message: "Invalid query parameters",
				Errors:  errors,
			}
		}

		return &CountQueryParam{"count", results}, ae
	}
}
