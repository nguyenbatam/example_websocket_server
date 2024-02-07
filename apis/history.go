package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/nguyenbatam/example_websocket_server/serializers"
	"github.com/nguyenbatam/example_websocket_server/services"
	"github.com/nguyenbatam/example_websocket_server/utils"
	"net/http"
	"time"
)

func (hdl *apiHandler) GetListMessageByRoom(c *gin.Context) {
	room := c.Param("room")
	start, _ := utils.QueryInt64Context(c, "start", 0)
	length, _ := utils.QueryInt64Context(c, "length", 20)
	if start == 0 {
		start = time.Now().UnixMilli()
	}
	data, next, _ := services.GetMessageByRoomChat(c.Request.Context(), room, start, length)
	c.AbortWithStatusJSON(http.StatusOK, serializers.NewListMessageResponse(data, next))
}
