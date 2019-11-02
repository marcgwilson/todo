package main

import (
	"github.com/marcgwilson/todo/state"

	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func CopyValues(q url.Values) url.Values {
	c := url.Values{}
	for k, v := range q {
		w := make([]string, len(v))
		copy(w, v)
		c[k] = w
	}
	return c
}

func GetQueryPage(q url.Values) int {
	if vals, ok := q["page"]; ok {
		if page, err := strconv.Atoi(vals[0]); err == nil {
			return page
		}
	}

	return 1
}

var UnrecognizedOp = errors.New("Unrecognized operator")

func timeClause(comp string, rfc3339 string) (string, error) {
	if t, err := time.Parse(time.RFC3339, rfc3339); err != nil {
		return "", err
	} else {
		op := ""
		switch comp {
		case "gt":
			op = ">"
		case "lt":
			op = "<"
		case "gte":
			op = ">="
		case "lte":
			op = "<="
		default:
			return "", UnrecognizedOp
		}

		return fmt.Sprintf("due %s %d", op, t.Unix()), nil
	}
}

type ParseError struct {
	Key     string
	Value   string
	Message string
}

type PaginatedFilter struct {
	Where []string
	Page  int
	Limit int
}

func (r *PaginatedFilter) Offset() int {
	return (r.Page - 1) * r.Limit
}

func (r *PaginatedFilter) Copy() *PaginatedFilter {
	where := make([]string, len(r.Where))
	copy(where, r.Where)
	return &PaginatedFilter{where, r.Page, r.Limit}
}

func (r *PaginatedFilter) String() string {
	result := ""

	if len(r.Where) > 0 {
		result = " WHERE " + strings.Join(r.Where, " AND ")
	}

	if r.Limit > 0 {
		result = fmt.Sprintf("%s LIMIT %d", result, r.Limit)
	}

	if r.Offset() > 0 {
		result = fmt.Sprintf("%s OFFSET %d", result, r.Offset())
	}

	return result + ";"
}

func (r *PaginatedFilter) CountString() string {
	result := ""

	if len(r.Where) > 0 {
		result = " WHERE " + strings.Join(r.Where, " AND ")
	}

	return result + ";"
}

var (
	TimeErrorMessage  = "value must be in RFC-3339 format"
	PageErrorMessage  = "value must be an integer greater than 0"
	CountErrorMessage = "value must be an integer greater than 0"
)

func ParseFilter(q map[string][]string) (*PaginatedFilter, []*ParseError) {
	where := []string{}
	page := 1
	limit := 20

	var errors []*ParseError

	for k, v := range q {
		comp := strings.Split(k, ":")

		switch comp[0] {
		case "due":
			if len(comp) == 2 {
				if c, err := timeClause(comp[1], v[0]); err == nil {
					where = append(where, c)
				} else if err != UnrecognizedOp {
					errors = append(errors, &ParseError{Key: k, Value: v[0], Message: TimeErrorMessage})
				}
			}
		case "state":
			ids := []string{}
			for _, elem := range v {
				if s, ok := state.States[elem]; !ok {
					errors = append(errors, &ParseError{Key: k, Value: elem, Message: "Invalid state"})
				} else {
					ids = append(ids, strconv.Quote(string(s)))
				}
			}
			where = append(where, fmt.Sprintf("state IN (%s)", strings.Join(ids, ",")))
		case "page":
			if p, err := strconv.Atoi(v[0]); err != nil {
				errors = append(errors, &ParseError{Key: k, Value: v[0], Message: PageErrorMessage})
			} else {
				if p < 1 {
					errors = append(errors, &ParseError{Key: k, Value: v[0], Message: PageErrorMessage})
				} else {
					page = p
				}
			}
		case "count":
			if p, err := strconv.Atoi(v[0]); err != nil {
				errors = append(errors, &ParseError{Key: k, Value: v[0], Message: CountErrorMessage})
			} else {
				if p < 1 {
					errors = append(errors, &ParseError{Key: k, Value: v[0], Message: CountErrorMessage})
				} else {
					limit = p
				}
			}
		}
	}

	return &PaginatedFilter{Where: where, Page: page, Limit: limit}, errors
}
