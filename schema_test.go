package main

import (
	"github.com/xeipuuv/gojsonschema"

	"testing"
)

func TestCreateValidator(t *testing.T) {
	var loader = gojsonschema.NewStringLoader(CreateSchema)
	if schema, err := gojsonschema.NewSchema(loader); err != nil {
		t.Error(err)
	} else {
		t.Logf("schema: %#v\n", schema)

		validDocument := gojsonschema.NewStringLoader(`{"desc": "My Todo", "due": "2002-10-02T10:00:00-05:00", "state": "todo"}`)

		if result, err := schema.Validate(validDocument); err != nil {
			t.Error(err)
		} else {
			t.Logf("result: %#v", result)
		}

		invalidDocument := gojsonschema.NewStringLoader(`{"desc": "My Todo", "due": "2002-10-02T10:00:00-05:00"}`)
		if result, err := schema.Validate(invalidDocument); err != nil {
			t.Error(err)
		} else {
			for i, e := range result.Errors() {
				t.Logf("%d: %#v", i, e)
			}
		}
	}
}

func TestUpdateValidator(t *testing.T) {
	var loader = gojsonschema.NewStringLoader(UpdateSchema)
	if schema, err := gojsonschema.NewSchema(loader); err != nil {
		t.Error(err)
	} else {
		t.Logf("schema: %#v\n", schema)

		validDocument := gojsonschema.NewStringLoader(`{"desc": "My Todo", "due": "2002-10-02T10:00:00-05:00", "state": "todo"}`)

		if result, err := schema.Validate(validDocument); err != nil {
			t.Error(err)
		} else {
			t.Logf("result: %#v", result)
		}

		invalid1 := gojsonschema.NewStringLoader(`{"desc": "My Todo", "due": "asdf"}`)

		if result, err := schema.Validate(invalid1); err != nil {
			t.Error(err)
		} else {
			for i, e := range result.Errors() {
				t.Logf("%d: %#v", i, e)
			}
		}

		invalid2 := gojsonschema.NewStringLoader(`{"state": "mystate"}`)

		if result, err := schema.Validate(invalid2); err != nil {
			t.Error(err)
		} else {
			for i, e := range result.Errors() {
				t.Logf("%d: %#v", i, e)
			}
		}

		invalid3 := gojsonschema.NewStringLoader(`{"mykey": "mystate"}`)

		if result, err := schema.Validate(invalid3); err != nil {
			t.Error(err)
		} else {
			for i, e := range result.Errors() {
				t.Logf("%d: %#v", i, e)
			}
		}
	}
}
