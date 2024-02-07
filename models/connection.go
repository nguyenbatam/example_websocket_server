package models

import (
	"github.com/nguyenbatam/example_websocket_server/utils/log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a protobuf to the peer.
	writeTimeOut = 5 * time.Second
)

type WsConnection struct {
	Id             string
	UserId         uint64
	Ws             *websocket.Conn
	mu             sync.Mutex
	Stop           bool
	Expired        bool
	ExpireTime     int64
	waitingSend    [][]byte
	waitingPongMsg bool
	length         int
	rooms          map[string]*Room
	roomMutex      sync.RWMutex
}

func NewConnection(ws *websocket.Conn, userId uint64, expireTime int64) *WsConnection {
	connect := WsConnection{
		Id:          uuid.NewString(),
		Ws:          ws,
		UserId:      userId,
		ExpireTime:  expireTime,
		Stop:        false,
		rooms:       make(map[string]*Room),
		waitingSend: [][]byte{},
	}
	go connect.loopWrite()
	return &connect
}

func (c *WsConnection) loopWrite() {
	defer func() {
		log.Logger.Info().Str("c.Id", c.Id).Msg("Stop loop write ")
		c.Ws.Close()
	}()
	for !c.Stop {
		waiting, sendPong := c.getListWaitingMsg()
		if sendPong {
			err := c.Ws.WriteMessage(websocket.PongMessage, nil)
			if err != nil {
				log.Logger.Error().Err(err).Str("c.Id", c.Id).Msg("err when write message")
				c.Stop = true
				return
			}
		}
		if len(waiting) == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		c.Ws.SetWriteDeadline(time.Now().Add(writeTimeOut))
		for _, data := range waiting {
			err := c.Ws.WriteMessage(websocket.TextMessage, data)
			if err != nil {
				log.Logger.Error().Err(err).Str("c.Id", c.Id).Msg("err when write message")
				c.Stop = true
				return
			}
		}
	}
}

func (c *WsConnection) getListWaitingMsg() ([][]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	waiting := c.waitingSend
	c.waitingSend = [][]byte{}
	waitingPong := c.waitingPongMsg
	c.waitingPongMsg = false
	return waiting, waitingPong
}

func (c *WsConnection) AddWaitingMsg(payload []byte) {
	if payload == nil || c.Stop {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.waitingSend = append(c.waitingSend, payload)
}

func (c *WsConnection) AddPongMsg() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.waitingPongMsg = true
}
func (conn *WsConnection) SubscribeRoom(room *Room) {
	conn.roomMutex.Lock()
	defer conn.roomMutex.Unlock()
	conn.rooms[room.Id] = room
}

func (conn *WsConnection) UnSubscribeRoom(roomId string) {
	conn.roomMutex.Lock()
	defer conn.roomMutex.Unlock()
	delete(conn.rooms, roomId)
}
func (conn *WsConnection) GetListRoom() []*Room {
	conn.roomMutex.RLock()
	defer conn.roomMutex.RUnlock()
	rooms := make([]*Room, len(conn.rooms))
	i := 0
	for _, room := range conn.rooms {
		rooms[i] = room
		i++
	}
	return rooms
}
