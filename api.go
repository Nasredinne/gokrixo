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

	router.HandleFunc("/CreateCommand", corsMiddleware(makeHTTPHandleFunc(s.handleCreateCommand)))
	router.HandleFunc("/GetCommands", makeHTTPHandleFunc(s.handleGetCommands))
	router.HandleFunc("/CreateWorker", corsMiddleware(makeHTTPHandleFunc(s.handleCreateWorker)))
	router.HandleFunc("/GetWorkers", makeHTTPHandleFunc(s.handleGetWorkers))
	router.HandleFunc("/Regestration", corsMiddleware(makeHTTPHandleFunc(s.handleRegestration)))
	router.HandleFunc("/account/{id}", corsMiddleware(withJWTAuth(makeHTTPHandleFunc(s.handleGetWorkerByID), s.store)))
	router.HandleFunc("/UpdateCommand", corsMiddleware(makeHTTPHandleFunc(s.handleUpdateCommand)))
	router.HandleFunc("/UpdateWorker", corsMiddleware(makeHTTPHandleFunc(s.handleUpdateWorker)))
	router.HandleFunc("/DeleteCommand", corsMiddleware(makeHTTPHandleFunc(s.handleDeleteCommand)))
	router.HandleFunc("/DeleteDataBaseTables", corsMiddleware(makeHTTPHandleFunc(s.handleDeleteDBTables)))
	log.Println("JSON API server running on port: ", s.listenAddr)

	http.ListenAndServe(s.listenAddr, router)
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*") // Or your frontend origin
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Allowed Origin
		w.Header().Set("Access-Control-Allow-Origin", "*") // Change this to your frontend origin
		// Allowed Methods
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT")
		// Allowed Headers
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle Preflight Request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Proceed to actual request
		next.ServeHTTP(w, r)
	}
}

func (s *APIServer) handleDeleteDBTables(w http.ResponseWriter, r *http.Request) error {
	err := s.store.DropAllTables()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusAccepted, "TABLES DELETED")
}

func (s *APIServer) handleCreateCommand(w http.ResponseWriter, r *http.Request) error {
	req := new(CreateCommandRequest)
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		return err
	}

	command, err := NewCommand(req.FullName, req.Number, req.Flor, req.Itemtype, req.Service, req.Workers, req.Start, req.Distination, req.Prix, req.IsAccepted)
	if err != nil {
		return err
	}
	if err := s.store.CreateCommand(command); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, command)
}

func (s *APIServer) handleGetCommands(w http.ResponseWriter, r *http.Request) error {
	enableCors(&w)

	commands, err := s.store.GetCommands()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, commands)
}

func (s *APIServer) handleCreateWorker(w http.ResponseWriter, r *http.Request) error {

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

	enableCors(&w)

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
	enableCors(&w)
	req := new(LoginRequest)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}

	if req.Email == "Krixo" || req.Password == "Nasro1234" {
		return WriteJSON(w, http.StatusOK, "Welcome Admin")
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

	// resp := LoginResponse{
	// 	Token: token,
	// 	ID:    worker.ID,
	// }
	http.SetCookie(w, &http.Cookie{
		Name:     "x-jwt-token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true, // Prevent JavaScript access
		Secure:   true, // Only send over HTTPS
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})

	return WriteJSON(w, http.StatusOK, worker)
}

func (s *APIServer) handleUpdateCommand(w http.ResponseWriter, r *http.Request) error {
	req := new(Command)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	err := s.store.UpdateCommand(req)
	if err != nil {
		return WriteJSON(w, http.StatusResetContent, err)
	}
	return WriteJSON(w, http.StatusAccepted, "Commend Updates Corectly")
}

func (s *APIServer) handleUpdateWorker(w http.ResponseWriter, r *http.Request) error {
	req := new(Worker)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	err := s.store.UpdateWorker(req)
	if err != nil {
		return WriteJSON(w, http.StatusResetContent, err)
	}
	return WriteJSON(w, http.StatusAccepted, "Worker Updates Corectly")
}

func (s *APIServer) handleDeleteCommand(w http.ResponseWriter, r *http.Request) error {
	req := new(Command)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return err
	}
	err := s.store.DeleteCommand(req)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Command Not Deleted")

		return err

	}
	return WriteJSON(w, http.StatusAccepted, "Command Deleted")

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
