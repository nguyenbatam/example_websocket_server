package models

import (
	"fmt"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

type TokenJWTInfo struct {
	Uid uint64 `json:"uid,omitempty"`
	Exp int64  `json:"exp,omitempty"`
}

var jWTCache = expirable.NewLRU[string, *TokenJWTInfo](30000, nil, -1)
var JwtSecretKey []byte

const TokenHourLifespan = 2

func InitJwtSecretKey(key string) {
	JwtSecretKey = []byte(key)
}
func GenerateToken(userId uint) (string, error) {
	claims := jwt.MapClaims{}
	claims["uid"] = userId
	claims["exp"] = time.Now().Add(time.Hour * time.Duration(TokenHourLifespan)).Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	println("JwtSecretKey", string(JwtSecretKey))
	return token.SignedString(JwtSecretKey)
}

func TokenStringIsValid(tokenString string) error {
	_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return JwtSecretKey, nil
	})
	if err != nil {
		return err
	}
	return nil
}

func ExtractToken(c *gin.Context) string {
	token := c.Query("token")
	if token != "" {
		return token
	}
	bearerToken := c.Request.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func ExtractTokenJWT(tokenString string) (*TokenJWTInfo, error) {
	if len(tokenString) == 0 {
		return nil, errors.New("token is empty")
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return JwtSecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("token is not jwt.MapClaims")
	}
	var tokenInfo TokenJWTInfo
	tokenInfo.Uid, err = strconv.ParseUint(fmt.Sprintf("%v", claims["uid"]), 10, 64)
	if err != nil {
		return nil, err
	}
	exp, err := strconv.ParseFloat(fmt.Sprintf("%v", claims["exp"]), 64)
	if err != nil {
		return nil, err
	}
	tokenInfo.Exp = int64(exp)
	return &tokenInfo, nil
}

func GetTokenJWT(c *gin.Context) (*TokenJWTInfo, error) {
	tokenString := ExtractToken(c)
	if len(tokenString) == 0 {
		return nil, errors.New("token is empty")
	}
	tokenInfo, _ := jWTCache.Get(tokenString)
	if tokenInfo != nil {
		return tokenInfo, nil
	}
	tokenInfo, err := ExtractTokenJWT(tokenString)
	if err != nil {
		return nil, err
	}
	if tokenInfo != nil {
		jWTCache.Add(tokenString, tokenInfo)
	}
	return tokenInfo, nil
}
