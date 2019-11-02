# TODO app

## Environment Variables
| **NAME**           | **DEFAULT** |
| :----------------- | :---------- |
| **`TODO_DB`**      | `todo.db`   |
| **`TODO_PORT`**    | `:8000`     |
| **`TODO_LIMIT`**   | `20`        |

## Query Parameters
| **NAME**           | **TYPE**                   |
| :----------------- | :------------------------- |
| **`due:gt`**       | **RFC-3339 DATETIME**      |
| **`due:gte`**      | **RFC-3339 DATETIME**      |
| **`due:lt`**       | **RFC-3339 DATETIME**      |
| **`due:lte`**      | **RFC-3339 DATETIME**      |
| **`state`**        | **[todo,in_process,done]** |
| **`page`**         | **int**                    |
| **`count`**        | **int**                    |

## Tests
```bash
go test 					# Run entire test suite 
go test -run TestRegex -v 	# Run single test in verbose mode
go test -run TestTodo -v 	# Run single test in verbose mode

go test -run HTTP/DELETE -v
go test -cover
```

## NOTES:
Comparing Due dates with time.Unix(). Fix!
