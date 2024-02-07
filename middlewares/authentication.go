package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyenbatam/example_websocket_server/models"
	"github.com/nguyenbatam/example_websocket_server/utils/log"
)

func JwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenInfo, err := models.GetTokenJWT(c)
		log.Logger.Info().Msg("error when parse jwt token")
		c.Set("jwt_info", tokenInfo)
		c.Set("err", err)
		c.Next()
	}
}
