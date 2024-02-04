package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/ginutils"
)

const (
	NoBasicAuth = "No basic auth provided"
	AuthFailure = "Basic auth failure"
)

type AuthGetter interface {
	GetAuth(string) (string, error)
	SetAuth(string) error
}

func BasicAuthMiddleware(authGetter AuthGetter) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, pass, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			ginutils.AbortWithError(c, http.StatusUnauthorized, NoBasicAuth)
			return
		}

		storePwd, err := authGetter.GetAuth(user)
		if err != nil || pass != storePwd {
			ginutils.AbortWithError(c, http.StatusInternalServerError, AuthFailure)
			return
		}

		c.Next()
	}
}
