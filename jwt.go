package main

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

const accessTokenSecret = "should be secuere enough"
const refreshTokenSecret = "should be secuere enough 2"

func createAccessToken(email string, roles []string) (string, error) {
	claims := jwt.MapClaims{
		"exp":   time.Now().Add(time.Minute * 15).Unix(), // Token expires in 15 minutes
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

func validateAccessToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(accessTokenSecret), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		email := claims["email"].(string)
		return email, nil
	} else {
		return "", err
	}
}
