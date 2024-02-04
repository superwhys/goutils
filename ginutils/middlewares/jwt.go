package middlewares

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/ginutils"
	"github.com/superwhys/goutils/lg"
)

const (
	UnAuthInfo = "Authorization failure"
)

func GenerateJWTAuth(signKey string, claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(signKey))
	if err != nil {
		lg.Errorf("jwt sign with key: %v error: %v", signKey, err)
		return "", errors.Wrap(err, "signedToken")
	}

	return tokenStr, nil
}

func JWTMiddleware(signKey string, claimsTmp jwt.Claims) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			ginutils.AbortWithError(c, http.StatusUnauthorized, UnAuthInfo)
			return
		}

		claimsType := reflect.TypeOf(claimsTmp)
		if claimsType.Kind() == reflect.Pointer {
			claimsType = claimsType.Elem()
		}
		claims := reflect.New(claimsType).Interface().(jwt.Claims)

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(signKey), nil
		})

		if err != nil {
			lg.Errorf("jwt parse error: %v", err)
			ginutils.AbortWithError(c, http.StatusUnauthorized, UnAuthInfo)
			return
		}

		if !token.Valid {
			lg.Errorf("auth failure, token validate: %v", token.Valid)
			ginutils.AbortWithError(c, http.StatusUnauthorized, UnAuthInfo)
			return
		}

		c.Set("claims", token.Claims)

		c.Next()
	}
}
