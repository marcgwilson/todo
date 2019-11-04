package main

import (
	"github.com/marcgwilson/todo/state"

	"fmt"
	"time"
)

func GenerateTodos(count int, tm *TodoManager) TodoList {
	results := []*Todo{}

	stateList := []state.State{state.Todo, state.InProgress, state.Done}

	for i := 0; i < count; i++ {
		t := time.Now().UTC()

		data := map[string]interface{}{
			"desc":  fmt.Sprintf("Todo %d", i),
			"due":   t.Format(time.RFC3339Nano),
			"state": stateList[i%3],
		}

		if todo, err := tm.Create(data); err != nil {
			panic(err)
		} else {
			results = append(results, todo)
		}
	}

	return results
}
