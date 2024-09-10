package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
)

type User struct {
	Email      string `json:"email"`
	FirstName  string `json:"first_name"`
	MiddleName string `json:"middle_name"`
	LastName   string `json:"last_name"`
	Password   string `json:"password"`
}

type Config struct {
	DbName     string
	DbUser     string
	DbPassword string
	DbHost     string
	DbPort     string
}

var config Config = Config{
	DbName:     os.Getenv("DB_NAME"),
	DbUser:     os.Getenv("DB_USER"),
	DbPassword: os.Getenv("DB_PASSWORD"),
	DbHost:     os.Getenv("DB_HOST"),
	DbPort:     os.Getenv("DB_PORT"),
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

func login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "logged up")
}

func logout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "logged out")
}

type Test struct {
	id         int
	name       string
	created_at string
}

var psqlInfo string = fmt.Sprintf("host=%s port=%s user=%s "+
	"password=%s dbname=%s sslmode=disable",
	config.DbHost, config.DbPort, config.DbUser, config.DbPassword, config.DbName)

func main() {
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/signup", signup)
	mux.HandleFunc("/api/login", login)
	mux.HandleFunc("/api/logout", logout)

	rows, err := db.Query("select * from test;")

	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {
		var title Test
		if err := rows.Scan(&title.id, &title.created_at, &title.name); err != nil {
			panic(err)
		}
		fmt.Println(title.id, title.created_at, title.name)
	}
	if err := rows.Err(); err != nil {
		panic(err)
	}

	fmt.Println("Starting server at :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
