package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

type APIServer struct {
	ListenAddr string
	Store      Storage
}

func NewAPIServer(listenAddr string, store Storage) *APIServer {
	return &APIServer{
		ListenAddr: listenAddr,
		Store:      store,
	}
}

// --------------------------------------------------
func (s *APIServer) run() error {
	router := mux.NewRouter()

	router.HandleFunc("/account/", makeItHandleFunc(s.handleAccount))
	router.HandleFunc("/account/{id}", WithJWTAuth(makeItHandleFunc(s.handleAccountByID), s.Store))
	router.HandleFunc("/transfer/", makeItHandleFunc(s.handleTransfer))

	log.Println("server running on port", s.ListenAddr)
	return http.ListenAndServe(s.ListenAddr, router)
}

// --------------------------------------------------
func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	default:
		return fmt.Errorf("method not allowed")
	}
}

// --------------------------------------------------
func (s *APIServer) handleAccountByID(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccountByID(w, r)
	case "DELETE":
		return s.handleDeleteAccountByID(w, r)
	default:
		return fmt.Errorf("method not allowed")
	}
}

// --------------------------------------------------
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.Store.GetAccounts()
	if err != nil {
		return err
	}
	return writeToJSON(w, http.StatusOK, accounts)
}

// --------------------------------------------------
func (s *APIServer) handleGetAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, _ := getParamID(r)
	account, err := s.Store.GetAccountByID(id)
	if err != nil {
		return err
	}
	return writeToJSON(w, http.StatusOK, account)
}

// --------------------------------------------------
func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	createAccountReq := new(CreateAccountRequest)
	if err := decodeToStruct(r, createAccountReq); err != nil {
		return err
	}
	account := NewAccount(createAccountReq.FirstName, createAccountReq.LastName)
	if err := s.Store.CreateNewAccount(account); err != nil {
		return err
	}

	tokenString, err := createJWT(account)
	if err != nil {
		return writeToJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
	}

	fmt.Println("token string: ", tokenString)

	return writeToJSON(w, http.StatusOK, account)
}

// --------------------------------------------------
func (s *APIServer) handleDeleteAccountByID(w http.ResponseWriter, r *http.Request) error {
	id, _ := getParamID(r)

	err := s.Store.DeleteAccount(id)
	if err != nil {
		return writeToJSON(w, http.StatusBadRequest, err)
	}

	return writeToJSON(w, http.StatusOK, "Done")
}

// --------------------------------------------------
func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {
	treq := new(TransferRequest)
	if err := decodeToStruct(r, treq); err != nil {
		return err
	}
	return writeToJSON(w, http.StatusOK, treq)
}

// --------------------------------------------------
func writeToJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

// this is our JWT middleware
func WithJWTAuth(handlerfunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("using JWT Auth MiddleWare")
		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		// fmt.Println("token: ", token)
		if err != nil {
			writeToJSON(w, http.StatusForbidden, APIError{Error: "permission denied"})
			return
		}
		if !token.Valid {
			writeToJSON(w, http.StatusForbidden, APIError{Error: "permission denied"})
			return
		}

		id, err := getParamID(r)
		if err != nil {
			writeToJSON(w, http.StatusForbidden, APIError{Error: "invalid token"})
			return
		}
		account, err := s.GetAccountByID(id)
		fmt.Println(account)
		if err != nil {
			writeToJSON(w, http.StatusInternalServerError, APIError{Error: err.Error()})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		fmt.Println(claims)
		if account.Number != int64(claims["accountNumber"].(float64)) {
			writeToJSON(w, http.StatusForbidden, APIError{Error: "permission denied"})
			return
		}
		handlerfunc(w, r)
	}
}

type apiFunc func(w http.ResponseWriter, r *http.Request) error

func makeItHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			// handle error
			writeToJSON(w, http.StatusBadRequest, APIError{Error: err.Error()})
		}
	}
}

func getParamID(r *http.Request) (int, error) {
	id := mux.Vars(r)["id"]
	return strconv.Atoi(id)
}

func decodeToStruct(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func createJWT(account *Account) (string, error) {
	secret := os.Getenv("JWT_SECRET")

	claims := &jwt.MapClaims{
		"accountNumber": account.Number,
		"ExpiresAt":     time.Now().Add(10 * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method:%v", token.Header["alg"])
		}
		return []byte(secret), nil
	})
}
