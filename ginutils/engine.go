package ginutils

import (
	"github.com/gin-gonic/gin"
	"github.com/superwhys/goutils/lg"
)

func New() *gin.Engine {
	engine := gin.New()

	if !lg.IsDebug() {
		gin.SetMode(gin.ReleaseMode)
	}
	engine.MaxMultipartMemory = 100 << 20
	engine.Use(lg.LoggerMiddleware(), gin.Recovery())

	return engine
}
