package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type User struct {
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Password   string `json:"password"`
}

var users []User = []User{}

func signup(w http.ResponseWriter, r *http.Request) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid payload")
		return
	}

	fmt.Println(user)

	if user.Email == "" || user.Password == "" || user.FirstName == "" || user.LastName == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid payload")
		return
	}

	for i := 0; i < len(users); i++ {
		if users[i].Email == user.Email {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "User already exists")
			return
		}
	}

	users = append(users, user)
	fmt.Println(users)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

var tokens map[string]string = make(map[string]string)

func login(w http.ResponseWriter, r *http.Request) {
	var loginCred struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginCred)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "email and password is required")
		return
	}

	var user User

	for i := 0; i < len(users); i++ {
		if users[i].Email == loginCred.Email && users[i].Password == loginCred.Password {
			user = users[i]
			break
		}
	}

	fmt.Println(user.Email)

	if user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "email and password doesn't match")
		return
	}

	tokens[user.Email] = "12345"

	res := map[string]string{
		"refresh": tokens[user.Email],
		"access":  tokens[user.Email],
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)

}

func logout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "logged out")
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/signup", signup)
	mux.HandleFunc("POST /api/login", login)
	mux.HandleFunc("/api/logout", logout)

	fmt.Println("Starting server at :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
