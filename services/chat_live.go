package services

import (
	"github.com/liyue201/gostl/utils/comparator"
	"github.com/nguyenbatam/example_websocket_server/common"
	"github.com/nguyenbatam/example_websocket_server/models"
	"github.com/nguyenbatam/example_websocket_server/models/protobuf"
	"github.com/nguyenbatam/example_websocket_server/utils/log"
	"github.com/nguyenbatam/example_websocket_server/utils/skiplist"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"net/http"
	"sync"
	"time"
)

type SocketServer struct {
	listUser  *skiplist.Skiplist[uint64, *models.User]
	listRoom  *skiplist.Skiplist[string, *models.Room]
	userMutex sync.RWMutex
	roomMutex sync.RWMutex
}

func NewServer() *SocketServer {
	var socket = SocketServer{
		listUser: skiplist.New[uint64, *models.User](comparator.Uint64Comparator, skiplist.WithGoroutineSafe()),
		listRoom: skiplist.New[string, *models.Room](comparator.StringComparator, skiplist.WithGoroutineSafe()),
	}
	return &socket
}

func (socket *SocketServer) CreateConnection(w http.ResponseWriter, r *http.Request, userId uint64, tokenInfo *models.TokenJWTInfo) (bool, error) {
	log.Logger.Info().Uint64("userId", userId).Msg("create new connection")
	ws, err := common.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Logger.Error().Err(err).Msg("err when try run upgrader")
		return false, err
	}
	conn := models.NewConnection(ws, userId, tokenInfo.Exp)
	user := socket.GetUser(userId)
	if user == nil {
		user = socket.GetOrInitUser(userId)
	}
	user.AddConnection(conn)
	ws.SetCloseHandler(func(code int, text string) error {
		socket.CloseAConnectionData(user, conn)
		return nil
	})
	ws.SetPingHandler(func(appData string) error {
		conn.AddPongMsg()
		ws.SetReadDeadline(time.Now().Add(common.ReadTimeOut))
		return nil
	})
	go socket.handleReceivedData(user, conn)
	return true, nil
}

func (socket *SocketServer) CloseAConnectionData(user *models.User, conn *models.WsConnection) {
	user.RemoveConnection(conn.Id)
	rooms := conn.GetListRoom()
	for _, room := range rooms {
		room.RemoveConnection(conn.Id)
		conn.UnSubscribeRoom(room.Id)
	}
	socket.RemoveUser(user)
}
func (socket *SocketServer) GetRoom(channel string) *models.Room {
	user, err := socket.listRoom.Get(channel)
	if err != nil {
		return nil
	}
	return user
}
func (socket *SocketServer) GetUser(userId uint64) *models.User {
	user, err := socket.listUser.Get(userId)
	if err != nil {
		return nil
	}
	return user
}

func (socket *SocketServer) RemoveUser(user *models.User) {
	socket.userMutex.Lock()
	defer socket.userMutex.Unlock()
	if user.CountConnection() > 0 {
		return
	}
	socket.listUser.Remove(user.Id)
}

func (socket *SocketServer) GetOrInitUser(userId uint64) *models.User {
	user, err := socket.listUser.Get(userId)
	if err == nil && user != nil {
		return user
	}
	socket.userMutex.Lock()
	defer socket.userMutex.Unlock()
	user, err = socket.listUser.Get(userId)
	if err == nil && user != nil {
		return user
	}
	user = &models.User{
		Id:          userId,
		Connections: skiplist.New[string, *models.WsConnection](comparator.StringComparator, skiplist.WithGoroutineSafe()),
	}
	socket.listUser.Insert(userId, user)
	return user
}
func (socket *SocketServer) GetOrInitRoom(channel string) *models.Room {
	room, err := socket.listRoom.Get(channel)
	if room != nil && err == nil {
		return room
	}
	socket.roomMutex.Lock()
	defer socket.roomMutex.Unlock()
	room, err = socket.listRoom.Get(channel)
	if room != nil && err == nil {
		return room
	}
	room = models.InitANewRoom(channel)
	socket.listRoom.Insert(channel, room)
	return room
}

func (socket *SocketServer) Subscribe(channel string, conn *models.WsConnection) bool {
	room := socket.GetOrInitRoom(channel)
	room.AddConnection(conn)
	conn.SubscribeRoom(room)
	return true
}

func (socket *SocketServer) Unsubscribe(channel string, conn *models.WsConnection) {
	room := socket.GetRoom(channel)
	if room == nil {
		return
	}
	conn.UnSubscribeRoom(channel)
	room.RemoveConnection(conn.Id)
}

func (socket *SocketServer) handleReceivedData(user *models.User, conn *models.WsConnection) {
	defer func() {
		log.Logger.Info().Str("c.Id", conn.Id).Msg("Stop loop wait read")
		socket.CloseAConnectionData(user, conn)
		conn.Stop = true
		conn.Ws.Close()
	}()
	con := conn.Ws
	con.SetReadDeadline(time.Now().Add(common.ReadTimeOut))
	for {
		_, data, err := con.ReadMessage()
		if err != nil {
			log.Logger.Error().Err(err).Str("conn.Id", conn.Id).Msg("Error when read data from connection")
			return
		}
		msg := &protobuf.MessageChat{}
		err = proto.Unmarshal(data, msg)
		if err != nil {
			log.Logger.Info().Err(err).Str("data", string(data)).Msg("Error when parser protobuf ")
			return
		}
		con.SetReadDeadline(time.Now().Add(common.ReadTimeOut))

		msg.Sender = conn.Id
		switch msg.Command {
		case common.CommandSubscribe:
			socket.Subscribe(msg.Room, conn)
		case common.CommandUnsubscribe:
			socket.Unsubscribe(msg.Room, conn)
		case common.CommandChat:
			AddMessageToRoomChat(msg)
		}
	}
}

func (socket *SocketServer) HandleSubscribeMessage(c <-chan *redis.Message) {
	for body := range c {
		var msg protobuf.MessageChat
		err := proto.Unmarshal([]byte(body.Payload), &msg)
		if err != nil {
			continue
		}
		room := socket.GetRoom(msg.Room)
		if room != nil {
			room.SendMsgToAll(&msg)
		}
	}
}
