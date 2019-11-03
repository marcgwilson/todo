package query

import "fmt"

type ParseError struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

func (r *ParseError) Error() string {
	return fmt.Sprintf("key: %s, value: %s, message: %s", r.Key, r.Value, r.Message)
}
