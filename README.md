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

## Response Body
### List
```json
{
  "next": "/?page=2&state=todo",
  "previous": "",
  "results": [
        {
            "id": 1,
            "desc": "My Todo",
            "due": "2019-11-12T06:14:11Z",
            "state": "todo"
        }
    ]
}
```

### Create, Update, Retrive
```json
{
  "id": 88,
  "desc": "In progress TODO",
  "due": "2019-11-13T23:50:33Z",
  "state": "in_progress"
}
```

### Delete
Empty response body


### Error
```json
{
  "code": 400,
  "message": "Invalid JSON",
  "errors": [
    {
      "key": "due",
      "value": "gabagoo",
      "message": "Does not match format 'rfc3339'"
    }
  ]
}
```

## Tests
```bash
go test ./... -cover                # Run entire test suite and print coverage %
go test -run TestTodo -v            # Run single test in verbose mode
go test -run Handler/DELETE -v      # Run DELETE subtest
go test -run Handler/LIST=all -v    # Run LIST=all subtest
```

## NOTES:
~~Comparing Due dates with time.Unix(). Fix!~~  
If the declared column type is `TIMESTAMP`, `go-sqlite3` attemps to handle `time.Time` instances for you. `INSERT` and `UPDATE` operations convert `time.Time` to `UTC`; however, `SELECT` queries do not perform this conversion automatically. See [sqlite3 doesn't have datetime/timestamp types #748](https://github.com/mattn/go-sqlite3/issues/748)  

