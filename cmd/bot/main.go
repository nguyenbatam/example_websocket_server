package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/nguyenbatam/example_websocket_server/common"
	"github.com/nguyenbatam/example_websocket_server/models"
	"github.com/nguyenbatam/example_websocket_server/models/protobuf"
	"github.com/pelletier/go-toml/v2"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	httpPort = flag.String("http_port", "9082", "http_port listen")
	conf     = flag.String("conf", "./config.toml", "config run file *.toml")
	c        = common.Config{}
)
var userStatus = map[int]bool{}
var mu sync.Mutex

var rooms = []string{}

func createMessageChat(room string, sender int, rand int) *protobuf.MessageChat {
	text := room + strconv.Itoa(sender) + strconv.Itoa(rand)
	md5byte := md5.Sum([]byte(text))
	id, _ := uuid.NewUUID()
	return &protobuf.MessageChat{
		Sender:  strconv.Itoa(sender),
		Id:      id.String(),
		Room:    room,
		Content: []byte("Msg: 網站有中、英文版本，也有繁、簡體版，可通過每頁左上角的連結隨時調整。" + hex.EncodeToString(md5byte[:])),
		Command: common.CommandChat,
	}
}

func createMessageSubscribe(room string, sender int) *protobuf.MessageChat {
	id, _ := uuid.NewUUID()
	return &protobuf.MessageChat{
		Sender:  strconv.Itoa(sender),
		Id:      id.String(),
		Room:    room,
		Command: common.CommandSubscribe,
	}
}
func createMessageUnSubscribe(room string, sender int) *protobuf.MessageChat {
	id, _ := uuid.NewUUID()
	return &protobuf.MessageChat{
		Sender:  strconv.Itoa(sender),
		Id:      id.String(),
		Room:    room,
		Command: common.CommandSubscribe,
	}
}

func createUser(userId int) {
	println("start user ", userId)
	mu.Lock()
	userStatus[userId] = true
	mu.Unlock()
	defer func() {
		mu.Lock()
		userStatus[userId] = false
		mu.Unlock()
		println("stop users ", userId)
	}()
	userToken, err := models.GenerateToken(uint(userId))
	if err != nil {
		println("error when gen token", err.Error())
	}
	err = models.TokenStringIsValid(userToken)
	if err != nil {
		println("verify token error", err.Error())
	}
	url := fmt.Sprintf("ws://localhost:%s/ws/%d?token=%s", *httpPort, userId, userToken)
	fmt.Println(url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		println("error when connect to server:", err.Error())
		return
	}
	stop := false
	defer c.Close()
	// send subcribe
	go func() {
		for !stop {
			mt, body, err := c.ReadMessage()
			if err != nil {
				stop = true
				return
			}
			if mt == websocket.TextMessage {
				var msg protobuf.MessageChat
				proto.Unmarshal(body, &msg)
				println(" user Id ", userId, " received Msg", msg.String())
			}
		}
	}()

	time.Sleep(1 * time.Second)
	room := rooms[userId%len(rooms)]
	message, _ := proto.Marshal(createMessageSubscribe(room, userId))
	err = c.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		println("err when try subscribe room", err.Error())
		stop = true
		return
	}
	//}
	time.Sleep(5 * time.Second)
	sleepTime := rand.Int() % 5

	for !stop {
		if userId < 100 {
			message, _ = proto.Marshal(createMessageChat(room, userId, sleepTime))
			err := c.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				println("err when send protobuf from ", userId, " to room ", room, err.Error())
				break
			}
		} else {
			err := c.WriteMessage(websocket.PingMessage, message)
			if err != nil {
				println("err when send protobuf from ", userId, " to room ", room, err.Error())
				break
			}
		}
		sleepTime = 15
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}

}

func makeRoom(max int) {
	for i := 0; i < max; i++ {
		md5byte := md5.Sum([]byte(strconv.Itoa(i)))
		md5RoomHex := hex.EncodeToString(md5byte[:])
		rooms = append(rooms, md5RoomHex)
	}
}

func main() {
	flag.Parse()
	configBytes, err := os.ReadFile(*conf)
	if err != nil {
		fmt.Println("err when read config file ", err, "file ", *conf)
	}
	err = toml.Unmarshal(configBytes, &c)
	if err != nil {
		fmt.Println("err when pass toml file ", err)
	}
	text, err := json.Marshal(c)
	fmt.Println("Success read config from toml file ", string(text))
	models.InitJwtSecretKey(c.JwtSecretKey)
	makeRoom(5)
	maxUser := 20
	for {
		count := 0
		for i := 0; i < maxUser; i++ {
			status := false
			mu.Lock()
			status = userStatus[i]
			mu.Unlock()
			if status == false {
				go createUser(i)
				count++
			}
		}
		if count > 0 {
			println("start again ", count, " user")
		}
		time.Sleep(1 * time.Second)
	}
}
