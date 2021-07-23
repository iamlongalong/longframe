package auth

import (
	"errors"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var ErrInvalidToken = errors.New("invalid token")

var sk = []byte("FAnrKbNawqhX3pTpC9FKUsm4hYXpVsHfRddtTuAkn4CYimAp94zwapbxzvFvvEVw")

type JWTClaims struct {
	ID    string `json:"IDs"`
	Name  string `json:"name"`
	Email string `json:"email"`
	jwt.StandardClaims
}

func NewClaim() *JWTClaims {
	nowTimeStamp := time.Now().Unix()
	return &JWTClaims{
		StandardClaims: jwt.StandardClaims{
			NotBefore: nowTimeStamp - 5,
			ExpiresAt: nowTimeStamp + 60*60*1, // 1 小时过期时间
			Issuer:    "night-fury",
		},
	}
}

func GenJwtToken(ID, name, email string) (string, error) {
	c := NewClaim()
	c.Name = name
	c.Email = email
	c.ID = ID

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	t, err := token.SignedString(sk)
	return t, err
}

func JwtTokenValidate(token string) (*JWTClaims, error) {
	jwtToken, err := jwt.ParseWithClaims(token, &JWTClaims{}, func(t *jwt.Token) (interface{}, error) {
		return sk, nil
	})

	if err != nil {
		return nil, err
	}

	if m, ok := jwtToken.Claims.(*JWTClaims); ok && jwtToken.Valid {
		return m, nil
	}
	return nil, ErrInvalidToken
}
