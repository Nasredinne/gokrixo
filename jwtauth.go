package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
)

// func createJWT(worker *Worker) (string, error) {
// 	claims := &jwt.MapClaims{
// 		"expiresAt": 15000,
// 		"workerid":  &worker.ID,
// 	}

// 	secret := os.Getenv("JWT_SECRET")
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

// 	return token.SignedString([]byte(secret))
// }

//eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5hc3JvQGdtYWlsLmNvbSIsImV4cCI6MTc1MTEyMjYzNCwiaWQiOiJjYTIwNzY5Ni1mZmYzLTQ3YWItYjkxMS1mZGI2ZGYzZGNlMTYifQ.pBdqEbVneQ9EQgIqTSBmAWIbueqs4_7TLsKul-dP9ns

func createJWT(worker *Worker) (string, error) {
	if worker == nil {
		return "", fmt.Errorf("worker is nil")
	}

	// Example JWT creation
	claims := jwt.MapClaims{
		"id":    worker.ID,
		"email": worker.Email,
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("miw_miw"))
}

func withJWTAuth(handlerFunc http.HandlerFunc, s Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("calling JWT auth middleware")
		cookie, err := r.Cookie("x-jwt-token")

		tokenString := cookie.Value
		//tokenString := r.Header.Get("x-jwt-token")
		token, err := validateJWT(tokenString)

		if err != nil {
			WriteJSON(w, http.StatusForbidden, err)
			return
		}
		if !token.Valid {
			permissionDenied(w)
			return
		}
		userID, err := getID(r)
		if err != nil {
			permissionDenied(w)
			return
		}
		account, err := s.GetAccountByID(userID)
		fmt.Println("ACCOUNT : ", account)
		if err != nil {
			permissionDenied(w)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Unauthorized: Invalid claims", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", claims["id"])
		handlerFunc.ServeHTTP(w, r.WithContext(ctx))

		// WriteJSON(w, http.StatusForbidden, ApiError{Error: "invalid token"})
		return

		//handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	// secret := os.Getenv("JWT_SECRET")

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte("miw_miw"), nil
	})
}

func permissionDenied(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission denied"})
}

func permissionAccepted(w http.ResponseWriter) {
	WriteJSON(w, http.StatusForbidden, ApiError{Error: "permission accepted"})
}

func getID(r *http.Request) (string, error) {
	idStr := mux.Vars(r)["id"]

	return idStr, nil
}
