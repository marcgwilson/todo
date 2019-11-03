package state

import (
	"encoding/json"
)

type State string

const (
	Todo       State = "todo"
	InProgress State = "in_progress"
	Done       State = "done"
)

var States = map[string]State{
	"todo":        Todo,
	"in_progress": InProgress,
	"done":        Done,
}

func (r State) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(r))
}
