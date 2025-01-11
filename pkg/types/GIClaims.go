package types

import "github.com/golang-jwt/jwt/v5"

type GIClaims struct {
	UserId string `json:"userId"`
	Admin  bool   `json:"admin"`
	jwt.RegisteredClaims
}
