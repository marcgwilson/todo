# TODO app

## Environment Variables
| **NAME**           | **DEFAULT** |
| :----------------- | :---------- |
| **`TODO_DB`**      | `todo.db`   |
| **`TODO_PORT`**    | `:8000`     |
| **`TODO_LIMIT`**   | `20`        |

## API
| **NAME**           | **METHOD**  | **URL**     |
| :----------------- | :---------- | :---------- |
| List               | **GET**     | `/`         |
| Create             | **POST**    | `/`         |
| Update             | **PATCH**   | `/:id/`     |
| Retrieve           | **GET**     | `/:id/`     |
| Delete             | **DELETE**  | `/:id/`     |

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
go test ./... -cover                # Run entire test suite and print coverage %
go test -run TestTodo -v            # Run single test in verbose mode
go test -run Handler/DELETE -v      # Run DELETE subtest
go test -run Handler/LIST=all -v    # Run LIST=all subtest
```

## NOTES:
Comparing Due dates with time.Unix(). Fix!
