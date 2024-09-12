package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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
var blackListedTokens []string = make([]string, 100)

func loadUsers() {
	userFile, err := os.Open("users.json")
	if err == nil {
		defer userFile.Close()
		byteValue, _ := io.ReadAll(userFile)
		json.Unmarshal(byteValue, &users)
	}
}

func saveUsers() {
	userFile, _ := json.Marshal(users)
	_ = os.WriteFile("users.json", userFile, 0644)
}

func saveBlacklistedTokens() {
	blackListedTokensFile, _ := json.Marshal(blackListedTokens)
	_ = os.WriteFile("blacklisted-tokens.json", blackListedTokensFile, 0644)
}

func loadBlacklistedTokens() {
	blackListedTokensFile, err := os.Open("blacklisted-tokens.json")
	if err == nil {
		defer blackListedTokensFile.Close()
		byteValue, _ := io.ReadAll(blackListedTokensFile)
		json.Unmarshal(byteValue, &blackListedTokens)
	}
}

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
	saveUsers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func logout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	refreshTokenString := r.Header.Get("X-Refresh-Token")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Authorization header missing")
		return
	}

	if refreshTokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "X-Refresh-Token header missing")
		return
	}

	accessTokenString := strings.TrimPrefix(authHeader, "Bearer ")
	_, err := validateAccessToken(accessTokenString)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid Authorization token")
		return
	}

	_, err = validateRefreshToken(refreshTokenString)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid X-Refresh-Token")
		return
	}

	blackListedTokens = append(blackListedTokens, accessTokenString, refreshTokenString)
	saveBlacklistedTokens()

	fmt.Fprint(w, "logged out")
}

func refresh(w http.ResponseWriter, r *http.Request) {
	oldRefreshTokenString := r.Header.Get("X-Refresh-Token")
	if oldRefreshTokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "X-Refresh-Token header missing")
		return
	}

	email, err := validateRefreshToken(oldRefreshTokenString)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid refresh token")
		return
	}

	for i := 0; i < len(blackListedTokens); i++ {
		if blackListedTokens[i] == oldRefreshTokenString {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Invalid refresh token")
			return
		}
	}

	var user User
	for i := 0; i < len(users); i++ {
		if users[i].Email == email {
			user = users[i]
			break
		}
	}

	if user.Email == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "User not found")
		return
	}

	accessToken, err := createAccessToken(user.Email, user.Roles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to generate access token")
		return
	}

	refreshToken, err := createRefreshToken(user.Email)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "failed to generate refresh token")
		return
	}

	blackListedTokens = append(blackListedTokens, oldRefreshTokenString)
	saveBlacklistedTokens()

	token := Token{
		Refresh: refreshToken,
		Access:  accessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(token)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
}

func me(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Authorization header missing")
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid token")
		return
	}

	email, err := validateAccessToken(tokenString)

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, "Invalid token")
		return
	}

	for i := 0; i < len(blackListedTokens); i++ {
		if blackListedTokens[i] == tokenString {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Invalid token")
			return
		}
	}

	var user User
	for i := 0; i < len(users); i++ {
		if users[i].Email == email {
			user = users[i]
			break
		}
	}

	if user.Email == "" {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, "User not found")
		return
	}

	user.Password = "" // Do not expose the password

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func main() {
	loadUsers()
	loadBlacklistedTokens()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/signup", signup)
	mux.HandleFunc("POST /api/login", login)
	mux.HandleFunc("POST /api/logout", logout)
	mux.HandleFunc("GET /api/me", me)
	mux.HandleFunc("POST /api/refresh", refresh)

	fmt.Println("Starting server at :8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
