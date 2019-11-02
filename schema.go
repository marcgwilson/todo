package main

import (
	"github.com/xeipuuv/gojsonschema"
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
      "format": "date-time"
    },
    "state": {
      "type": "string",
      "enum": ["todo", "in_progress", "done"]
    }
  },
  "required": ["desc", "due", "state"]
}`

// strict-rfc3339
const UpdateSchema = `{
  "title": "Todo Update Schema",
  "type": "object",
  "properties": {
    "desc": {
      "type": "string"
    },
    "due": {
      "type": "string",
      "format": "date-time"
    },
    "state": {
      "type": "string",
      "enum": ["todo", "in_progress", "done"]
    }
  }
}`

var CreateValidator = gojsonschema.NewStringLoader(CreateSchema)
var UpdateValidator = gojsonschema.NewStringLoader(UpdateSchema)
