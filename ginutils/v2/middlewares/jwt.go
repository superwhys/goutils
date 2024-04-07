package middlewares

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
	"github.com/superwhys/goutils/ginutils/v2"
	"github.com/superwhys/goutils/lg"
)

const (
	AuthHeaderKey = "Authorization"

	UnAuthInfo          = "Authorization failure"
	TokenExpired        = "Authorization: Token is expired"
	TokenNoBearerPrefix = "Authorization: Bearer your_access_token"
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

func jwtTokenCheck(token string) (bool, string, string) {
	if token == "" {
		return false, "", UnAuthInfo
	}

	if !strings.HasPrefix(token, "Bearer ") {
		return false, "", TokenNoBearerPrefix
	}
	return true, strings.Replace(token, "Bearer ", "", 1), ""
}

func JWTMiddleware(signKey string, claimsTmp jwt.Claims) gin.HandlerFunc {
	return func(c *gin.Context) {
		headerToken := c.GetHeader(AuthHeaderKey)
		valid, tokenString, errMsg := jwtTokenCheck(headerToken)
		if !valid {
			ginutils.AbortWithError(c, http.StatusUnauthorized, errMsg)
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
			message := UnAuthInfo
			if errors.Is(err, jwt.ErrTokenExpired) {
				message = TokenExpired
			}
			ginutils.AbortWithError(c, http.StatusUnauthorized, message)
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

func JWTMiddlewareHandler(signKey string, claimsTmp jwt.Claims) ginutils.Handler {
	return ginutils.WrapGinHandlerFunc(JWTMiddleware(signKey, claimsTmp))
}
