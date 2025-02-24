package tokens

import "github.com/golang-jwt/jwt/v5"

type JWTAccessClaims struct {
	jwt.RegisteredClaims
	Sign string `json:"sign"`
	Type string `json:"type"`
}

type JWTRefreshClaims struct {
	jwt.RegisteredClaims
	Sign    string `json:"sign"`
	Type    string `json:"type"`
	Version int    `json:"version"`
}
