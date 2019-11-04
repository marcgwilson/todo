package main

import (
	"github.com/marcgwilson/todo/state"

	"fmt"
	"strings"
	"time"
)

type AttrTransform func(interface{}) (interface{}, error)

func TransformDue(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case string:
		if t, err := time.Parse(time.RFC3339Nano, v); err != nil {
			return nil, fmt.Errorf("Invalid datetime format")
		} else {
			return t, nil
		}
	case time.Time:
		return v, nil
	default:
		return v, fmt.Errorf("Invalid type")
	}
}

func TransformState(i interface{}) (interface{}, error) {
	switch v := i.(type) {
	case string:
		return v, nil
	case state.State:
		return string(v), nil
	default:
		return v, fmt.Errorf("Invalid type")
	}
}

var attrMap = map[string]AttrTransform{
	"due":   TransformDue,
	"state": TransformState,
}

type TodoMap map[string]interface{}

func (r TodoMap) State() state.State {
	return state.State(r["state"].(string))
}

func (r TodoMap) Due() time.Time {
	switch v := r["due"].(type) {
	case time.Time:
		return v
	default:
		return time.Unix(0, v.(int64))
	}
}

func (r TodoMap) Description() string {
	return r["desc"].(string)
}

type SQLData struct {
	Names    string
	Bindvars string
	Values   []interface{}
}

func Transform(data TodoMap) (TodoMap, error) {
	result := map[string]interface{}{}
	for k, v := range data {
		if transform, ok := attrMap[k]; ok {
			if transformed, err := transform(v); err != nil {
				return nil, err
			} else {
				result[k] = transformed
			}
		} else {
			result[k] = v
		}
	}
	return result, nil
}

func (r TodoMap) InsertVars() *SQLData {
	length := len(r)

	keys := make([]string, length, length)
	bindvars := make([]string, length, length)
	values := make([]interface{}, length, length)
	i := 0
	for k, v := range r {
		keys[i] = k
		values[i] = v
		bindvars[i] = "?"
		i++
	}

	return &SQLData{strings.Join(keys, ", "), strings.Join(bindvars, ", "), values}
}

func (r TodoMap) UpdateVars() *SQLData {
	length := len(r)
	bindvars := make([]string, length, length)
	values := make([]interface{}, length, length)
	i := 0
	for k, v := range r {
		values[i] = v
		bindvars[i] = fmt.Sprintf("%s = ?", k)
		i++
	}

	return &SQLData{"", strings.Join(bindvars, ", "), values}
}
