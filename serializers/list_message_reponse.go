package serializers

import "github.com/nguyenbatam/example_websocket_server/models/protobuf"

type ListMessageResponse struct {
	Next uint64                 `json:"next"`
	Data []protobuf.MessageChat `json:"data"`
}

func NewListMessageResponse(data []protobuf.MessageChat, next uint64) ListMessageResponse {
	return ListMessageResponse{Next: next, Data: data}
}
