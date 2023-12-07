package app

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gitlab.jiebu.com/base/log"
	"net/http"
)

func logIdInjector(c *gin.Context) {
	log.InjectLogID(c)
}

func corsMiddleware() gin.HandlerFunc {
	corsCfg := cors.DefaultConfig()
	corsCfg.AllowAllOrigins = true
	corsCfg.AllowHeaders = append(corsCfg.AllowHeaders, "Authorization")
	corsCfg.OptionsResponseStatusCode = http.StatusOK
	return cors.New(corsCfg)
}
