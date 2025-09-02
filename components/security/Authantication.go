package security

import (
	"banking-app-be/components/config"
	"banking-app-be/components/errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

func ValidateToken(_ http.ResponseWriter, r *http.Request, claim *Claims) error {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return errors.NewUnauthorizedError("missing or invalid Authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		return errors.NewUnauthorizedError("empty token in Authorization header")
	}

	// fmt.Println("token string ============>>", tokenString)

	token, err := checkToken(tokenString, claim)
	if err != nil {
		return errors.NewUnauthorizedError("invalid token: " + err.Error())
	}

	if !token.Valid {
		return errors.NewUnauthorizedError("invalid token")
	}

	return nil
}

// Checks Token String
func checkToken(tokenString string, claim *Claims) (*jwt.Token, error) {

	token, err := jwt.ParseWithClaims(tokenString, claim, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTKey.GetStringValue()), nil
	})
	fmt.Println("token err==> ", err)
	return token, err
}
