package apierror

import (
	"github.com/xeipuuv/gojsonschema"

	"net/http"
	"sort"
)

type Error struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Errors  []*ErrorDetail `json:"errors"`
}

type ErrorDetail struct {
	Key     string      `json:"key"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

type sortByKey []*ErrorDetail

func (r sortByKey) Len() int           { return len(r) }
func (r sortByKey) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r sortByKey) Less(i, j int) bool { return r[i].Key < r[j].Key }

func sortValidationErrors(values []*ErrorDetail) {
	sort.Sort(sortByKey(values))
}

func NewJSONError(errorList []gojsonschema.ResultError) *Error {
	errors := make([]*ErrorDetail, len(errorList), len(errorList))
	for index, error := range errorList {
		switch error.(type) {
		case *gojsonschema.RequiredError:
			if property, ok := error.Details()["property"]; ok {
				prop := property.(string)
				errors[index] = &ErrorDetail{prop, "", "required attribute"}
			} else {
				errors[index] = &ErrorDetail{error.Field(), error.Value(), error.Description()}
			}
		default:
			errors[index] = &ErrorDetail{error.Field(), error.Value(), error.Description()}
		}
	}

	sortValidationErrors(errors)
	return &Error{http.StatusBadRequest, "Invalid JSON", errors}
}
