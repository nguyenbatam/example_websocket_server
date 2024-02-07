package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nguyenbatam/example_websocket_server/apis"
	"github.com/nguyenbatam/example_websocket_server/common"
	"github.com/nguyenbatam/example_websocket_server/middlewares"
	"github.com/nguyenbatam/example_websocket_server/models"
	"github.com/nguyenbatam/example_websocket_server/redis"
	"github.com/nguyenbatam/example_websocket_server/services"
	"github.com/nguyenbatam/example_websocket_server/utils/log"
	"github.com/pelletier/go-toml/v2"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	httpPort = flag.String("http_port", "9082", "http_port listen")
	conf     = flag.String("conf", "./config.toml", "config run file *.toml")
	c        = common.Config{}
)

func main() {
	log.Init(0)
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
	err = redis.Init(c.Redis)
	if err != nil {
		fmt.Println("err when connect redis", err)
	}
	defer redis.Close()
	models.InitJwtSecretKey(c.JwtSecretKey)

	subscriberChanel := redis.SubscribeChannel(context.Background(), redis.PubSubChannel)
	skyRepo := services.NewServer()
	for i := 0; i < c.NumberWorker; i++ {
		go skyRepo.HandleSubscribeMessage(subscriberChanel)
	}
	handler := apis.NewApiHandler(skyRepo)
	router := gin.New()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	ws := router.Group("/ws")
	ws.Use(middlewares.JwtAuthMiddleware())
	{
		ws.GET("/:userId", handler.OpenConnection)
	}
	restApi := router.Group("/api")
	{
		restApi.GET("/room/:room", handler.GetListMessageByRoom)
	}
	router.Run(":" + *httpPort)
}
