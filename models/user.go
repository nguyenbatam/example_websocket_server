package models

import (
	"github.com/nguyenbatam/example_websocket_server/utils/skiplist"
)

type User struct {
	Id          uint64
	Connections *skiplist.Skiplist[string, *WsConnection]
}

func (user *User) AddConnection(conn *WsConnection) {
	user.Connections.Insert(conn.Id, conn)
}

func (user *User) RemoveConnection(connId string) {
	user.Connections.Remove(connId)
}

func (user *User) CountConnection() int {
	return user.Connections.Len()
}
