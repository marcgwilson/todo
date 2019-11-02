package main

import (
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"

	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
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

		if q, err := ParseFilter(req.URL.Query()); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(err)
		} else {
			if list, err := r.TM.List(q); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				e := []*APIError{&APIError{Code: http.StatusBadRequest, Message: err.Error()}}
				json.NewEncoder(w).Encode(e)
			} else {

				log.Printf("req: %#v\n", req.Host)
				log.Printf("protocol: %s\n", req.Proto)
				
				next := ""
				previous := ""

				countQuery := q.Copy()
				countQuery.Page = 1
				countQuery.Limit = 0

				pageNum := q.Page

				if count, err := r.TM.Count(countQuery); err != nil {
					log.Printf("Error querying count: %s", err)
				} else {
					if q.Offset()+q.Limit < count {
						urlQuery := CopyValues(req.URL.Query())
						urlQuery.Set("page", strconv.Itoa(pageNum+1))
						next = urlQuery.Encode()
					}

					if q.Offset() > 0 {
						urlQuery := CopyValues(req.URL.Query())
						urlQuery.Set("page", strconv.Itoa(pageNum-1))
						previous = urlQuery.Encode()
					}
				}

				result := &PaginatedResult{Next: next, Previous: previous, Results: list}
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(result)
			}
		}
	}
}

func (r *Handler) CreateFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		defer req.Body.Close()

		var body []byte
		var err error

		if body, err = ioutil.ReadAll(req.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			e := []*APIError{&APIError{http.StatusBadRequest, err.Error()}}
			json.NewEncoder(w).Encode(e)
			return
		}

		var data map[string]interface{}
		if err = json.Unmarshal(body, &data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			e := []*APIError{&APIError{http.StatusBadRequest, err.Error()}}
			json.NewEncoder(w).Encode(e)
			return
		}

		doc := gojsonschema.NewGoLoader(data)

		if result, err := r.CreateValidator.Validate(doc); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			e := &APIError{http.StatusInternalServerError, err.Error()}
			json.NewEncoder(w).Encode(e)
		} else {
			if len(result.Errors()) != 0 {
				validationErrors := []*ValidationError{}
				for _, e := range result.Errors() {
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

				apiError := &APIValidationError{http.StatusBadRequest, "Validator Error", validationErrors}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(apiError)
			} else {
				if t, err := r.TM.Create(data); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					e := []*APIError{&APIError{http.StatusBadRequest, err.Error()}}
					json.NewEncoder(w).Encode(e)
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

		var body []byte
		var err error

		if body, err = ioutil.ReadAll(req.Body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			e := []*APIError{&APIError{http.StatusBadRequest, err.Error()}}
			json.NewEncoder(w).Encode(e)
			return
		}

		var data map[string]interface{}
		if err = json.Unmarshal(body, &data); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			e := []*APIError{&APIError{http.StatusBadRequest, err.Error()}}
			json.NewEncoder(w).Encode(e)
			return
		}

		doc := gojsonschema.NewGoLoader(data)

		if result, err := r.UpdateValidator.Validate(doc); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			e := &APIError{http.StatusInternalServerError, err.Error()}
			json.NewEncoder(w).Encode(e)
		} else {
			if len(result.Errors()) != 0 {
				validationErrors := []*ValidationError{}
				for _, e := range result.Errors() {
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

				apiError := &APIValidationError{http.StatusBadRequest, "Validator Error", validationErrors}
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(apiError)
			} else {
				var todo *Todo = nil
				if todo, err = r.TM.Update(id, data); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					e := []*APIError{&APIError{http.StatusBadRequest, err.Error()}}
					json.NewEncoder(w).Encode(e)
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
