package main

import (
	"fmt"
	"net/http"
	"os"
)

func main() {
	loadUsers()
	loadBlacklistedTokens()

	mainRouter := http.NewServeMux()

	v1 := http.NewServeMux()

	v1.HandleFunc("POST /auth/signup", signup)
	v1.HandleFunc("POST /auth/login", login)
	v1.HandleFunc("GET /auth/me", me)
	v1.HandleFunc("POST /auth/refresh", refresh)
	v1.HandleFunc("POST /auth/logout", logout)

	api := http.NewServeMux()
	api.Handle("/v1/", http.StripPrefix("/v1", v1))

	mainRouter.Handle("/api/", http.StripPrefix("/api", api))

	fmt.Println("Starting server at :8080")
	err := http.ListenAndServe(":8080", mainRouter)
	if err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
