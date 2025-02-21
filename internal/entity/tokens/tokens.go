package tokens

import "github.com/golang-jwt/jwt/v5"

type JWTAccessClaims struct {
	jwt.RegisteredClaims
	Sign string
	Type string
}

type JWTRefreshClaims struct {
	jwt.RegisteredClaims
	Sign string
	Type string
}
