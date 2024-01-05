package ginutils

import "github.com/gin-gonic/gin"

func SetDevMode() {
	gin.SetMode(gin.DebugMode)
}

func SetProdMode() {
	gin.SetMode(gin.ReleaseMode)
}
