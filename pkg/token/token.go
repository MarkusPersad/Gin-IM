package token

import (
	"Gin-IM/pkg/exception"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
)

var apiSecret []byte

func init() {
	apiSecret = []byte(os.Getenv("JWT_SECRET"))
	if len(apiSecret) == 0 {
		log.Logger.Panic().Msg("JWT_SECRET is not set")
	}
}

func GernerateToken(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(apiSecret)
}

func TokenValid(ctx *gin.Context) error {
	tokenString := ExtractToken(ctx)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return apiSecret, nil
	})
	if err != nil {
		return err
	}
	if _, ok := token.Claims.(jwt.Claims); ok && token.Valid {
		return nil
	}
	return exception.ErrInvalidToken
}

func ExtractToken(ctx *gin.Context) string {
	bearerToken := ctx.GetHeader("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ExtractClaims(ctx *gin.Context, claims jwt.Claims) error {
	tokenString := ExtractToken(ctx)
	if tokenString == "" {
		return exception.ErrTokenEmpty
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, exception.ErrUnknownAlg
		}
		return apiSecret, nil
	})
	if err != nil {
		return exception.ErrInvalidToken
	}
	claims, ok := token.Claims.(jwt.Claims)
	if ok && token.Valid {
		return nil
	}
	return exception.ErrInvalidToken
}
