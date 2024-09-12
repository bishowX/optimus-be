package main

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const accessTokenSecret = "should be secuere enough"
const refreshTokenSecret = "should be secuere enough 2"

func createAccessToken(email string, roles []string) (string, error) {
	claims := jwt.MapClaims{
		"exp":   time.Now().Add(time.Minute * 15).Unix(), // Token expires in 15 hours
		"iat":   time.Now().Unix(),
		"email": email,
		"roles": roles,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(accessTokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func createRefreshToken(email string) (string, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(time.Hour * 24 * 30).Unix(), // Token expires in 30 days
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(refreshTokenSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
