package models

import (
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/nguyenbatam/example_websocket_server/common"
	"github.com/nguyenbatam/example_websocket_server/models/protobuf"
	"github.com/nguyenbatam/example_websocket_server/utils/skiplist"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

const maxHash = 100

type Room struct {
	Id          string
	Connections map[int]*skiplist.Skiplist[string, *WsConnection]
	Users       map[uint64]struct{}
	mu          sync.RWMutex
	Close       bool
}

func InitANewRoom(roomId string) *Room {
	connections := make(map[int]*skiplist.Skiplist[string, *WsConnection])
	for i := 0; i < maxHash; i++ {
		connections[i] = skiplist.New[string, *WsConnection](comparator.StringComparator)
	}
	room := &Room{
		Id:          roomId,
		Connections: connections,
		Users:       make(map[uint64]struct{}),
	}
	go room.updateListUser()
	return room
}
func (room *Room) updateListUser() {
	for !room.Close {
		time.Sleep(10 * time.Second)
		updated := map[uint64]struct{}{}
		for _, connections := range room.Connections {
			connections.Traversal(func(connId string, conn *WsConnection) bool {
				updated[conn.UserId] = struct{}{}
				return true
			})
		}
		room.Users = updated
	}
}

func (room *Room) AddConnection(conn *WsConnection) {
	if conn == nil {
		return
	}
	hashValue := common.HashToRange(conn.Id)
	room.Connections[hashValue].Insert(conn.Id, conn)
}

func (room *Room) RemoveConnection(connId string) {
	hashValue := common.HashToRange(connId)
	room.Connections[hashValue].Remove(connId)
}

func (room *Room) SendMsgToAll(msg *protobuf.MessageChat) {
	payload, _ := proto.Marshal(msg)
	for _, connections := range room.Connections {
		connections.Traversal(func(connId string, conn *WsConnection) bool {
			conn.AddWaitingMsg(payload)
			return true
		})
	}
}
