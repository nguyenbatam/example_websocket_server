package apis

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nguyenbatam/example_websocket_server/common"
	"github.com/nguyenbatam/example_websocket_server/models"
	"github.com/nguyenbatam/example_websocket_server/utils"
	"github.com/nguyenbatam/example_websocket_server/utils/log"
	"time"
)

func (hdl *apiHandler) OpenConnection(c *gin.Context) {
	userId, _ := utils.ParamUint64Context(c, "userId", 0)
	log.Logger.Info().Uint64("userId", userId).Msg("open connection ")
	jwtTokenObject, ok := c.Get("jwt_info")
	var tokenInfo *models.TokenJWTInfo
	if ok && jwtTokenObject != nil {
		tokenInfo = jwtTokenObject.(*models.TokenJWTInfo)
	}
	if tokenInfo == nil {
		errJWT, _ := c.Get("err")
		ws, err := common.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		ws.WriteMessage(websocket.TextMessage, []byte(errJWT.(error).Error()))
		return
	}
	if tokenInfo.Uid != userId {
		ws, err := common.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error when decode token : got %d , wanted %d", tokenInfo.Uid, userId)))
		return
	}
	now := time.Now().Unix()
	if tokenInfo.Exp < now {
		ws, err := common.Upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer ws.Close()
		ws.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Token is expired : now %d , exp %d", now, tokenInfo.Exp)))
		return
	}
	hdl.repo.CreateConnection(c.Writer, c.Request, userId, tokenInfo)
}
