package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listerAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listerAddr,
		store:      store,
	}
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method == "GET" {
		return s.handleGetAccount(w, r)
	}
	if r.Method == "POST" {
		return s.handleCreateSegment(w, r)

	}
	if r.Method == "DELETE" {
		return s.handleDeleteSegment(w, r)
	}
	if r.Method == "PUT" {
		return s.handleAddUserToSegment(w, r)
	}
	return fmt.Errorf("method now allowed %s", r.Method)

}

func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return fmt.Errorf("invalid id given %s", idStr)
	}

	user, err := s.store.GetActiveSegments(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, user)
}

func (s *APIServer) handleCreateSegment(w http.ResponseWriter, r *http.Request) error {
	createSegmentReq := new(User)
	if err := json.NewDecoder(r.Body).Decode(createSegmentReq); err != nil {
		return err
	}
	segment, _ := NewSegment(createSegmentReq.Segment)
	if err := s.store.CreateSegment(segment); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, segment)
}

func (s *APIServer) handleDeleteSegment(w http.ResponseWriter, r *http.Request) error {
	segment := mux.Vars(r)["slug"]
	if err := s.store.DeleteSegment(segment); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, "segment deleted sucsessfully")
}

func (s *APIServer) handleAddUserToSegment(w http.ResponseWriter, r *http.Request) error {
	createSegmentRequest := new(Request)
	if err := json.NewDecoder(r.Body).Decode(createSegmentRequest); err != nil {
		return err
	}
	segment := NewRequest(createSegmentRequest.UserID, createSegmentRequest.AddSegments, createSegmentRequest.RemoveSegments)
	if err := s.store.AddUserToSegment(segment); err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, segment)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

func (s *APIServer) Run() {
	router := mux.NewRouter()
	router.HandleFunc("/segment", makeHTTPHandleFunc(s.handleCreateSegment))
	router.HandleFunc("/segments/{id}", makeHTTPHandleFunc(s.handleGetAccount))
	router.HandleFunc("/segment/{slug}", makeHTTPHandleFunc(s.handleDeleteSegment))
	router.HandleFunc("/segmentUpdate", makeHTTPHandleFunc(s.handleAddUserToSegment))
	http.ListenAndServe(s.listenAddr, router)

	log.Panicln("API server running on port", s.listenAddr)

}

type ApiError struct {
	Error string `json:"error"`
}
