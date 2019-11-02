package main

import (
	"reflect"
	"sort"
	"testing"
)

func TestQuery(t *testing.T) {
	q := map[string][]string{
		"due:gt": []string{"2019-10-30T03:09:33Z"},
		"due:lt": []string{"2019-10-30T03:09:53Z"},
		"state":  []string{"todo", "done"},
		"page":   []string{"5"},
		"count":  []string{"20"},
	}

	// expected := "due > 1572404973 AND due < 1572404993 AND state IN (1,3) OFFSET 4 LIMIT 20"
	expected := &PaginatedFilter{
		Where: []string{"state IN (\"todo\",\"done\")", "due > 1572404973", "due < 1572404993"},
		Page:  5,
		Limit: 20,
	}

	if actual, err := ParseFilter(q); err != nil {
		t.Error(err)
	} else {
		sort.Strings(expected.Where)
		sort.Strings(actual.Where)
		if !reflect.DeepEqual(expected, actual) {
			t.Errorf("%#v != %#v", expected, actual)
		}

		if expected.String() != actual.String() {
			t.Errorf("%s != %s", expected.String(), actual.String())
		}
	}
}

func TestQueryErrors(t *testing.T) {
	q := map[string][]string{
		"due:gt": []string{"2019-10-30T03:09:33Z"},
		"due:lt": []string{"asdf"},
		"state":  []string{"todo", "unknown"},
		"page":   []string{"string instead of number"},
		"count":  []string{"0"},
	}

	errorSet := map[string]*ParseError{
		"due:lt": &ParseError{Key: "due:lt", Value: "asdf", Message: "value must be in RFC-3339 format"},
		"state":  &ParseError{Key: "state", Value: "unknown", Message: "Invalid state"},
		"page":   &ParseError{Key: "page", Value: "string instead of number", Message: "value must be an integer greater than 0"},
		"count":  &ParseError{Key: "count", Value: "0", Message: "value must be an integer greater than 0"},
	}

	if _, err := ParseFilter(q); err != nil {
		for i, e := range err {
			expected := errorSet[e.Key]
			if !reflect.DeepEqual(expected, e) {
				t.Errorf("%d: %#v != %#v", i, expected, e)
			}
		}
	} else {
		t.Error("Expected non-nil error")
	}
}
