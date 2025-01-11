package token

import (
	"Gin-IM/pkg/exception"
	"Gin-IM/pkg/types"
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

func GernerateToken(claims jwt.Claims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	if tokenString, err := token.SignedString(apiSecret); err == nil {
		return tokenString
	} else {
		log.Logger.Error().Err(err).Msg("GernerateToken error")
		return ""
	}
}

func TokenValid(ctx *gin.Context) error {
	tokenString := ExtractToken(ctx)
	if tokenString == "" {
		return exception.ErrTokenEmpty
	}
	tokens, err := jwt.ParseWithClaims(tokenString, &types.GIClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, exception.ErrUnknownAlg
		}
		return apiSecret, nil
	})
	if err != nil {
		return err
	}
	if _, ok := tokens.Claims.(*types.GIClaims); ok && tokens.Valid {
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

func ExtractClaims(ctx *gin.Context) (*types.GIClaims, error) {
	tokenString := ExtractToken(ctx)
	if tokenString == "" {
		return nil, exception.ErrTokenEmpty
	}
	tokens, err := jwt.ParseWithClaims(tokenString, &types.GIClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, exception.ErrUnknownAlg
		}
		return apiSecret, nil
	})
	if err != nil {
		return nil, exception.ErrInvalidToken
	}
	claim, ok := tokens.Claims.(*types.GIClaims)
	if ok && tokens.Valid {
		return claim, nil
	}
	return nil, exception.ErrInvalidToken
}
