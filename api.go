package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type apiError struct {
	Error string
}

// eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2NvdW50TnVtYmVyIjo4NDM1OSwiZXhwaXJlc0F0IjoxNTAwMH0.Q7h11gJ-US3dea3Owr_LpD38CX8DKhnYH_YnbRTUj1o

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)

}
func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, apiError{Error: err.Error()})
		}
	}

}

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

	router.HandleFunc("/account", makeHTTPHandleFunc(s.handleAccount))
	router.HandleFunc("/accounts", makeHTTPHandleFunc(s.handleGetAccounts))
	router.HandleFunc("/account/{id}", withJWTAuth(makeHTTPHandleFunc(s.handleGetAccountById), s.store))
	router.HandleFunc("/delete-account/{id}", makeHTTPHandleFunc(s.handleDeleteAccount))

	fmt.Printf("API server listening on %s\n", s.listenAddr)
	http.ListenAndServe(s.listenAddr, router)
}

func (s *APIServer) handleAccount(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return s.handleGetAccount(w, r)
	case "POST":
		return s.handleCreateAccount(w, r)
	default:
		return fmt.Errorf("unsupported method %s", r.Method)
	}
}

func (s *APIServer) handleGetAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := s.store.GetAccounts()
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, accounts)
}
func (s *APIServer) handleGetAccount(w http.ResponseWriter, r *http.Request) error {
	id := mux.Vars(r)["id"]
	account := NewAccount("Amier", "Eid")
	log.Println("id:", id)
	return WriteJSON(w, http.StatusOK, account)
}

func (s *APIServer) handleCreateAccount(w http.ResponseWriter, r *http.Request) error {
	defer r.Body.Close()
	createAccountRequest := new(CreateAccountRequest)
	if err := json.NewDecoder(r.Body).Decode(createAccountRequest); err != nil {
		return err
	}

	acc := NewAccount(createAccountRequest.FirstName, createAccountRequest.LastName)
	err := s.store.CreateAccount(acc)
	if err != nil {
		return err
	}

	tokenString, err := createJWT(acc)
	if err != nil {
		// log.Println("error:", err)
		return err
	}

	fmt.Println("JWT token:", tokenString)

	return WriteJSON(w, http.StatusOK, acc)
}

func (s *APIServer) handleDeleteAccount(w http.ResponseWriter, r *http.Request) error {
	if r.Method != "DELETE" {
		return fmt.Errorf("unsupported method %s", r.Method)
	}
	id, err := getId(r)
	if err != nil {
		return err
	}
	acc, _ := s.store.GetAccountById(id)

	err = s.store.DeleteAccount(id)
	if err != nil {
		return fmt.Errorf("error deleting account", err)
	}

	return WriteJSON(w, http.StatusOK, acc)
}

func (s *APIServer) handleTransfer(w http.ResponseWriter, r *http.Request) error {

	return nil
}

func (s *APIServer) handleGetAccountById(w http.ResponseWriter, r *http.Request) error {
	id, err := getId(r)
	if err != nil {
		return err
	}
	account, err := s.store.GetAccountById(id)
	if err != nil {
		return err
	}
	return WriteJSON(w, http.StatusOK, account)
}

func getId(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("invalid id 0x0001")
	}
	return id, nil
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("This is JWT WRAPPER!	")
		tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "permission denied"})
			// log.Println("error:", err)
			return
		}

		if !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "permission denied"})
			return
		}

		userId, err := getId(r)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "permission denied"})
			return
		}
		account, err := s.GetAccountById(userId)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "permission denied"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		log.Println("test:", account.Number != int64(claims["accountNumber"].(float64)))
		log.Println("test:", account.Number)
		log.Printf("test:%T , %v", int64(claims["accountNumber"].(float64)), int64(claims["accountNumber"].(float64)))
		if account.Number != int64(claims["accountNumber"].(float64)) {
			WriteJSON(w, http.StatusUnauthorized, apiError{Error: "invalid token"})
			return
		}

		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := os.Getenv("JWT_SECRET")
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(secret), nil
	})
}

func createJWT(account *Account) (string, error) {
	claims := &jwt.MapClaims{
		"expiresAt":     15000,
		"accountNumber": account.Number,
	}

	sec := os.Getenv("JWT_SECRET")
	// log.Println("secret:", sec)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(sec))
}
