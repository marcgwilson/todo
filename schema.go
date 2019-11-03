package main

import (
	"github.com/xeipuuv/gojsonschema"

	"time"
)

const CreateSchema = `{
  "title": "Todo Create Schema",
  "type": "object",
  "properties": {
    "desc": {
      "type": "string"
    },
    "due": {
	  "type": "string",
      "format": "rfc3339"
    },
    "state": {
      "type": "string",
      "enum": ["todo", "in_progress", "done"]
    }
  },
  "required": ["desc", "due", "state"],
  "additionalProperties": false
}`

const UpdateSchema = `{
  "title": "Todo Update Schema",
  "type": "object",
  "properties": {
    "desc": {
      "type": "string"
    },
    "due": {
      "type": "string",
      "format": "rfc3339"
    },
    "state": {
      "type": "string",
      "enum": ["todo", "in_progress", "done"]
    }
  },
  "additionalProperties": false
}`

func init() {
	gojsonschema.FormatCheckers.Add("rfc3339", RFC3339FormatChecker{})
}

var CreateValidator = gojsonschema.NewStringLoader(CreateSchema)
var UpdateValidator = gojsonschema.NewStringLoader(UpdateSchema)

type RFC3339FormatChecker struct{}

func (f RFC3339FormatChecker) IsFormat(input interface{}) bool {
	asString, ok := input.(string)
	if !ok {
		return false
	}

	formats := []string{
		time.RFC3339,
	}

	for _, format := range formats {
		if _, err := time.Parse(format, asString); err == nil {
			return true
		}
	}

	return false
}
