package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type APIServer struct {
	listenAddr string
	store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		listenAddr: listenAddr,
		store:      store,
	}
}

func (s *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/CreateCommand", makeHTTPHandleFunc(s.handleCreateCommand))
	router.HandleFunc("/GetCommands", makeHTTPHandleFunc(s.handleGetCommands))
	router.HandleFunc("/CreateWorker", makeHTTPHandleFunc(s.handleCreateWorker))
	router.HandleFunc("/GetWorkers", makeHTTPHandleFunc(s.handleGetWorkers))
	router.HandleFunc("/Regestration", makeHTTPHandleFunc(s.handleRegestration))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetWorkerByID), s.store))

	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleCreateCommand(w http.ResponseWriter, r *http.Request) error {
	req := new(CreateCommandRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	command, err := NewCommand(req.FullName, req.Phone, req.Service, req.Workers, req.Start, req.Distination)
	if err != nil {
		return err
	}
	if err := s.store.CreateCommand(command); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, command)
}

func (s *APIServer) handleGetCommands(w http.ResponseWriter, r *http.Request) error {
	commands, err := s.store.GetCommands()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, commands)
}

func (s *APIServer) handleCreateWorker(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "POST" {
		return fmt.Errorf("methode not allowed %s", r.Method)
	}

	req := new(CreateWorkerRequest)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	//get account
	err := s.store.CreateWorker(
		&Worker{
			FullName:   req.FullName,
			Number:     req.Number,
			Email:      req.Email,
			Password:   req.Password,
			Position:   req.Position,
			Experience: req.Experience,
			Message:    req.Message,
			IsAccepted: req.IsAccepted,
		})
	return err
}

func (s *APIServer) handleGetWorkers(w http.ResponseWriter, r *http.Request) error {
	workers, err := s.store.GetWorkers()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, workers)
}
func (s *APIServer) handleGetWorkerByID(w http.ResponseWriter, r *http.Request) error {

	fmt.Println("get to handleGetWorkerByID ")
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := s.store.GetAccountByID(id)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)

}

func (s *APIServer) handleRegestration(w http.ResponseWriter, r *http.Request) error {
	req := new(LoginRequest)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	worker, err := s.store.Register(req.Password, req.Email)
	if err != nil {
		WriteJSON(w, http.StatusNotAcceptable, err)
		fmt.Println("PREBLEME FROM REGISTRATION ///////// WORKER :", worker)

	}

	token, err := createJWT(worker)
	if err != nil {
		return err
	}

	resp := LoginResponse{
		Token: token,
		ID:    worker.ID,
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "x-jwt-token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true, // Prevent JavaScript access
		Secure:   true, // Only send over HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	return WriteJSON(w, http.StatusOK, resp)
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(v)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}
