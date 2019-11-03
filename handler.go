package main

import (
	"github.com/marcgwilson/todo/query"

	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"

	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"log"
)

type Handler struct {
	TM              *TodoManager
	CreateValidator *gojsonschema.Schema
	UpdateValidator *gojsonschema.Schema
}

func NewHandler(tm *TodoManager) *Handler {
	var createSchema *gojsonschema.Schema
	var updateSchema *gojsonschema.Schema
	var err error
	var loader = gojsonschema.NewStringLoader(CreateSchema)
	if createSchema, err = gojsonschema.NewSchema(loader); err != nil {
		panic(err)
	}

	loader = gojsonschema.NewStringLoader(UpdateSchema)
	if updateSchema, err = gojsonschema.NewSchema(loader); err != nil {
		panic(err)
	}

	return &Handler{
		tm,
		createSchema,
		updateSchema,
	}
}

func (r *Handler) RetrieveFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, _ := strconv.ParseInt(mux.Vars(req)["id"], 10, 64)

		if t, err := r.TM.Get(id); err != nil {
			w.WriteHeader(http.StatusNotFound)
			e := []*APIError{&APIError{http.StatusNotFound, "Not found"}}
			json.NewEncoder(w).Encode(e)
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(t)
		}
	}
}

func (r *Handler) ListFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()
		result := query.ParseValues(req.URL.Query())
		if result.Errors() != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(result.Errors()[0])
		} else {
			result.Paginate()

			q := result.Query()
			if list, err := r.TM.Query(q); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				e := []*APIError{&APIError{Code: http.StatusBadRequest, Message: err.Error()}}
				json.NewEncoder(w).Encode(e)
			} else {
				pr := &PaginatedResult{Results: list}
				rc := result.ShallowCopy().Depaginate()
				qc := rc.Query()
				if count, err := r.TM.Count(qc); err != nil {
					log.Printf("ERROR: r.TM.Count: %s", err)
				} else {
					pr.Next = query.NextPage(result, req.URL, count)
				}
				
				pr.Previous = query.PrevPage(result, req.URL)
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(pr)
			}
		}
	}
}

func (r *Handler) CreateFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var data TodoMap
		var err *APIError

		if data, err = UnmarshalJSONRequest(req); err != nil {
			w.WriteHeader(err.Code)
			json.NewEncoder(w).Encode([]*APIError{err})
			return
		}	

		doc := gojsonschema.NewGoLoader(data)

		if result, err := r.CreateValidator.Validate(doc); err != nil {
			e := &APIError{http.StatusInternalServerError, err.Error()}
			w.WriteHeader(e.Code)
			json.NewEncoder(w).Encode(e)
		} else {
			if len(result.Errors()) != 0 {
				apiError := NewAPIValidationError(result.Errors())
				w.WriteHeader(apiError.Code)
				json.NewEncoder(w).Encode(apiError)
			} else {
				if t, err := r.TM.Create(data); err != nil {
					apiError := &APIError{http.StatusBadRequest, err.Error()}
					w.WriteHeader(apiError.Code)
					json.NewEncoder(w).Encode([]*APIError{apiError})
				} else {
					w.WriteHeader(http.StatusCreated)
					json.NewEncoder(w).Encode(t)
				}
			}
		}
	}
}

func (r *Handler) UpdateFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id, _ := strconv.ParseInt(mux.Vars(req)["id"], 10, 64)

		defer req.Body.Close()

		var data TodoMap
		var err *APIError

		if data, err = UnmarshalJSONRequest(req); err != nil {
			w.WriteHeader(err.Code)
			json.NewEncoder(w).Encode([]*APIError{err})
			return
		}

		doc := gojsonschema.NewGoLoader(data)

		if result, err := r.UpdateValidator.Validate(doc); err != nil {
			e := &APIError{http.StatusInternalServerError, err.Error()}
			w.WriteHeader(e.Code)
			json.NewEncoder(w).Encode(e)
		} else {
			if len(result.Errors()) != 0 {
				apiError := NewAPIValidationError(result.Errors())
				w.WriteHeader(apiError.Code)
				json.NewEncoder(w).Encode(apiError)
			} else {
				if todo, err := r.TM.Update(id, data); err != nil {
					apiError := &APIError{http.StatusBadRequest, err.Error()}
					w.WriteHeader(apiError.Code)
					json.NewEncoder(w).Encode([]*APIError{apiError})
				} else {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(todo)
				}
			}
		}
	}
}

func (r *Handler) DeleteFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		id, _ := strconv.ParseInt(mux.Vars(req)["id"], 10, 64)

		if _, err := r.TM.Get(id); err != nil {
			w.WriteHeader(http.StatusNotFound)
			e := []*APIError{&APIError{http.StatusNotFound, "Not found"}}
			json.NewEncoder(w).Encode(e)
			return
		}

		if err := r.TM.Delete(id); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			e := []*APIError{&APIError{http.StatusInternalServerError, err.Error()}}
			json.NewEncoder(w).Encode(e)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func UnmarshalJSONRequest(req *http.Request) (TodoMap, *APIError) {
	var body []byte
	var err error

	if body, err = ioutil.ReadAll(req.Body); err != nil {
		return nil, &APIError{http.StatusBadRequest, err.Error()}
	}

	var data TodoMap
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, &APIError{http.StatusBadRequest, err.Error()}
	}

	return data, nil
}

func NewAPIValidationError(errorList []gojsonschema.ResultError) *APIValidationError {
	validationErrors := []*ValidationError{}
	for _, e := range errorList {
		switch e.(type) {
		case *gojsonschema.RequiredError:
			if property, ok := e.Details()["property"]; ok {
				prop := property.(string)
				validationErrors = append(validationErrors, &ValidationError{prop, "", "required attribute"})
			}
		default:
			validationErrors = append(validationErrors, &ValidationError{e.Field(), e.Value(), e.Description()})
		}
	}

	return &APIValidationError{http.StatusBadRequest, "Validator Error", validationErrors}
}
