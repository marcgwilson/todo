package query

import (
	"github.com/marcgwilson/todo/state"

	"fmt"
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
	"due:gt": DueDateParser("due >"),
	"due:lt": DueDateParser("due <"),
	"due:gte": DueDateParser("due >="),
	"due:lte": DueDateParser("due <="),
	"due": DueDateParser("due ="),
	"state": StateParser("state"),
	"page": PageParser("page"),
	"count": CountParser("count"),
}

type IQueryParam interface {
    Name() string
    Values() []interface{}
}

type ParamListParser func([]string) (IQueryParam, []error)

type DueDateQueryParam struct {
	name string
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
	return func(values []string) (IQueryParam, []error) {
		results := []interface{}{}
		errors := []error{}
		for _, value := range values {
			if t, err := time.Parse(time.RFC3339, value); err != nil {
				errors = append(errors, &ParseError{Key: param, Value: value, Message: TimeErrorMessage})
			} else {
				results = append(results, t.Unix())
			}
		}
		return &DueDateQueryParam{param, results}, errors
	}
}

type StateQueryParam struct {
	name string
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
	return func(values []string) (IQueryParam, []error) {
		results := []interface{}{}
		errors := []error{}
		for _, value := range values {
			if val, ok := state.States[value]; !ok {
				errors = append(errors, &ParseError{Key: "state", Value: value, Message: "invalid state"})
			} else {
				results = append(results, val)
			}
		}

		return &StateQueryParam{"state", results}, errors
	}
}

type PageQueryParam struct {
	name string
	values []interface{}
	count int64
}

func (r *PageQueryParam) SetCount(c int64)  {
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
	return func (values []string) (IQueryParam, []error) {
		results := []interface{}{}
		errors := []error{}

		if result, err := strconv.ParseInt(values[0], 10, 64); err != nil {
			errors = append(errors, &ParseError{Key: "page", Value: values[0], Message: PageErrorMessage})
		} else {
			if result < 1 {
				errors = append(errors, &ParseError{Key: "page", Value: values[0], Message: PageErrorMessage})
			} else {
				results = append(results, result)
			}
		}

		return &PageQueryParam{"page", results, 0}, errors
	}
}

type CountQueryParam struct {
	name string
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
	return func (values []string) (IQueryParam, []error) {
		results := []interface{}{}
		errors := []error{}

		if result, err := strconv.ParseInt(values[0], 10, 64); err != nil {
			errors = append(errors, &ParseError{Key: "count", Value: values[0], Message: CountErrorMessage})
		} else {
			if result < 1 {
				errors = append(errors, &ParseError{Key: "count", Value: values[0], Message: CountErrorMessage})
			} else {
				results = append(results, result)
			}
		}

		return &CountQueryParam{"count", results}, errors
	}
}
