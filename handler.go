package main

import (
	"github.com/marcgwilson/todo/apierror"
	"github.com/marcgwilson/todo/query"

	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"

	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
)

type PaginatedResponse struct {
	Next     string   `json:"next"`
	Previous string   `json:"previous"`
	Results  TodoList `json:"results"`
}

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
			e := &apierror.Error{Code: http.StatusNotFound, Message: "Not found"}
			w.WriteHeader(e.Code)
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
		if result, err := query.ParseValues(req.URL.Query()); err != nil {
			w.WriteHeader(err.Code)
			json.NewEncoder(w).Encode(err)
		} else {
			result.Paginate()

			q := result.Query()
			if list, err := r.TM.Query(q); err != nil {
				e := &apierror.Error{Code: http.StatusBadRequest, Message: err.Error()}
				w.WriteHeader(e.Code)
				json.NewEncoder(w).Encode(e)
			} else {
				pr := &PaginatedResponse{Results: list}
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
		var err *apierror.Error

		if data, err = UnmarshalJSONRequest(req); err != nil {
			w.WriteHeader(err.Code)
			json.NewEncoder(w).Encode([]*apierror.Error{err})
			return
		}

		doc := gojsonschema.NewGoLoader(data)

		if result, err := r.CreateValidator.Validate(doc); err != nil {
			e := &apierror.Error{Code: http.StatusInternalServerError, Message: err.Error()}
			w.WriteHeader(e.Code)
			json.NewEncoder(w).Encode(e)
		} else {
			if len(result.Errors()) != 0 {
				ae := apierror.NewJSONError(result.Errors())
				w.WriteHeader(ae.Code)
				json.NewEncoder(w).Encode(ae)
			} else {
				if t, err := r.TM.Create(data); err != nil {
					ae := &apierror.Error{Code: http.StatusBadRequest, Message: err.Error()}
					w.WriteHeader(ae.Code)
					json.NewEncoder(w).Encode(ae)
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
		var ae *apierror.Error

		if data, ae = UnmarshalJSONRequest(req); ae != nil {
			w.WriteHeader(ae.Code)
			json.NewEncoder(w).Encode(ae)
			return
		}

		doc := gojsonschema.NewGoLoader(data)

		if result, err := r.UpdateValidator.Validate(doc); err != nil {
			e := &apierror.Error{Code: http.StatusInternalServerError, Message: err.Error()}
			w.WriteHeader(e.Code)
			json.NewEncoder(w).Encode(e)
		} else {
			if len(result.Errors()) != 0 {
				ae = apierror.NewJSONError(result.Errors())
				w.WriteHeader(ae.Code)
				json.NewEncoder(w).Encode(ae)
			} else {
				if todo, err := r.TM.Update(id, data); err != nil {
					ae = &apierror.Error{Code: http.StatusBadRequest, Message: err.Error()}
					w.WriteHeader(ae.Code)
					json.NewEncoder(w).Encode(ae)
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
			e := &apierror.Error{Code: http.StatusNotFound, Message: "Not found"}
			w.WriteHeader(e.Code)
			json.NewEncoder(w).Encode(e)
			return
		}

		if err := r.TM.Delete(id); err != nil {
			e := &apierror.Error{Code: http.StatusInternalServerError, Message: err.Error()}
			w.WriteHeader(e.Code)
			json.NewEncoder(w).Encode(e)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

func UnmarshalJSONRequest(req *http.Request) (TodoMap, *apierror.Error) {
	var body []byte
	var err error

	if body, err = ioutil.ReadAll(req.Body); err != nil {
		return nil, &apierror.Error{Code: http.StatusBadRequest, Message: err.Error()}
	}

	var data TodoMap
	if err = json.Unmarshal(body, &data); err != nil {
		return nil, &apierror.Error{Code: http.StatusBadRequest, Message: err.Error()}
	}

	return data, nil
}
