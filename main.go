package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/rs/cors"
)

func main() {
	loadUsers()
	loadBlacklistedTokens()
	loadContents()

	mainRouter := http.NewServeMux()

	v1 := http.NewServeMux()

	v1.HandleFunc("POST /auth/signup", signup)
	v1.HandleFunc("POST /auth/login", login)
	v1.HandleFunc("GET /auth/me", me)
	v1.HandleFunc("POST /auth/refresh", refresh)
	v1.HandleFunc("POST /auth/logout", logout)

	v1.HandleFunc("POST /contents", createContent)
	v1.HandleFunc("GET /contents", getContents)
	v1.HandleFunc("GET /contents/{id}", getContent)
	api := http.NewServeMux()
	api.Handle("/v1/", http.StripPrefix("/v1", v1))

	mainRouter.Handle("/api/", http.StripPrefix("/api", api))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:4173", "https://optimus.bishow.xyz"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"*"},
		// Enable Debugging for testing, consider disabling in production
		Debug: true,
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	fmt.Printf("Starting server at :%s", port)
	err := http.ListenAndServe(fmt.Sprintf(":%s", port), c.Handler(mainRouter))
	if err != nil {
		fmt.Println("Error starting server: ", err)
		os.Exit(1)
	}
}
