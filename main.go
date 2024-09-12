package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type User struct {
	Id         string   `json:"id"`
	Email      string   `json:"email"`
	FirstName  string   `json:"first_name"`
	MiddleName string   `json:"middle_name"`
	LastName   string   `json:"last_name"`
	Password   string   `json:"password"`
	Roles      []string `json:"roles"`
}

type Token struct {
	Access  string `json:"access"`
	Refresh string `json:"refresh"`
}

var users []User = []User{}
var tokens map[string]Token = make(map[string]Token)

func login(w http.ResponseWriter, r *http.Request) {
	var loginCred struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginCred)

	if err != nil {
		fmt.Println(err)
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

	if user.Email == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "email and password don't match")
		return
	}

	accessToken, err := createAccessToken(user.Email, user.Roles)
	refreshToken, err := createRefreshToken(user.Email)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to generate access token")
	}

	token := Token{
		Refresh: refreshToken,
		Access:  accessToken,
	}

	fmt.Println(token)

	tokens[user.Email] = token

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func signup(w http.ResponseWriter, r *http.Request) {
	var user User

	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid payload")
		return
	}

	fmt.Println(user)

	if user.Email == "" || user.Password == "" || user.FirstName == "" || user.LastName == "" {
		fmt.Println(err)
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

	user.Roles = []string{"editor"}
	user.Id = "345"

	users = append(users, user)
	fmt.Println(users)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
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
