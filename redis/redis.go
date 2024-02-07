package redis

import (
	"context"
	config "github.com/nguyenbatam/example_websocket_server/common"
	"github.com/redis/go-redis/v9"
	"strconv"
)

var rdb *redis.Client
var redisEmpty = "redis: nil"

func Init(config config.Redis) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
	})
	return rdb.Ping(context.Background()).Err()
}

func Close() error {
	return rdb.Close()
}

func AddItemToXStream(ctx context.Context, key string, data map[string]interface{}, maxLen int64) error {
	_, err := rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: key,
		MaxLen: maxLen,
		Values: data,
	}).Result()
	if err != nil {
		return err
	}
	return nil
}

func GetItemsByXStream(ctx context.Context, key string, start int64, limit int64) ([]redis.XMessage, error) {
	data, err := rdb.XRevRangeN(ctx, key, strconv.FormatInt(start, 10), "-", limit).Result()
	if err != nil {
		if err.Error() == redisEmpty {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

func PublishMsg(ctx context.Context, channel string, message string) error {
	_, err := rdb.Publish(ctx, channel, message).Result()
	if err != nil {
		return err
	}
	return nil
}

func SubscribeChannel(ctx context.Context, channels ...string) <-chan *redis.Message {
	return rdb.Subscribe(ctx, channels...).Channel()
}
