package apis

import "github.com/nguyenbatam/example_websocket_server/services"

type apiHandler struct {
	repo *services.SocketServer
}

func NewApiHandler(repo *services.SocketServer) *apiHandler {
	return &apiHandler{
		repo: repo,
	}
}
