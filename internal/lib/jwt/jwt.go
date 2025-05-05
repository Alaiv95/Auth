package jwt

import (
	"Auth/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

func New(user models.User, app models.App, ttl time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = user.Id
	claims["email"] = user.Email
	claims["exp"] = time.Now().Add(ttl).Unix()
	claims["app_id"] = app.Id

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
