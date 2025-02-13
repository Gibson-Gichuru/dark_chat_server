package database

import (
	"context"
	"darkchat/monitor"
	"darkchat/protocol"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var databaseMonitor *monitor.Monitor

// init initializes the Redis client and database monitor. It pings the Redis
// server to test the connection, and logs an error if there is one. If the
// connection is successful, it logs an info message to the database log.

func init() {
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	databaseMonitor = monitor.New("database.log")

	_, err := redisClient.Ping(context.Background()).Result()

	if err != nil {
		databaseMonitor.Error(err.Error())
	}

	databaseMonitor.Info("Connected to Redis")
}

// RegisterClientChat creates a new Redis Stream for the given chatId if it does
// not exist, and does nothing if it does exist. The function times out after
// 5 seconds, and returns an error if there was an error communicating with
// Redis.
func RegisterClientChat(chatId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	_, err := redisClient.XInfoStream(ctx, chatId).Result()

	if err != nil {
		if strings.Contains(err.Error(), "no such key") {
			_, err = redisClient.XAdd(
				ctx,
				&redis.XAddArgs{
					Stream: chatId,
					Values: map[string]interface{}{
						"chatId": chatId,
					},
				},
			).Result()

			if err != nil {
				return err
			}
			databaseMonitor.Info(fmt.Sprintf("Created stream: %s", chatId))
		} else {
			return errors.New("failed to create stream")
		}
	}

	return nil

}

// DeleteClientChat deletes a Redis Stream for the given chatId. The function
// times out after 5 seconds, and returns an error if there was an error
// communicating with Redis. If the stream does not exist, the function does
// nothing and does not return an error.
func DeleteClientChat(chatId string) error {
	_, err := redisClient.Del(context.Background(), chatId).Result()

	if err != nil {
		return err
	}
	databaseMonitor.Info(fmt.Sprintf("Deleted stream: %s", chatId))

	return nil
}

// StreamChat listens for new messages on the specified Redis stream (chatId)
// and sends them to the provided chatChannel as protocol.Payloads.
// If an error occurs during reading from the stream, it logs the error and stops execution.

func StreamChat(chatChannel chan<- protocol.Payload, chatId string) {
	ctx := context.Background()
	defer close(chatChannel)

	for {
		msg, err := redisClient.XRead(ctx, &redis.XReadArgs{
			Streams: []string{chatId, "$"},
			Count:   1,
			Block:   0,
		}).Result()

		if err != nil {
			databaseMonitor.Error(err.Error())
			return
		}

		if len(msg) == 0 || len(msg[0].Messages) == 0 {
			continue
		}

		messageString, ok := msg[0].Messages[0].Values["message"].(string)
		if !ok {
			databaseMonitor.Error("Failed to convert message to string")
			return
		}

		var message protocol.Message

		err = json.Unmarshal([]byte(messageString), &message)

		if err != nil {
			databaseMonitor.Error(err.Error())
			return
		}
		chatChannel <- &message

	}
}

// PostToChat sends a message to a Redis Stream identified by the given chatId.
// If an error occurs while communicating with Redis, the error is returned.
// If the timeout (5 seconds) is exceeded, the context is canceled and an error is returned.
// If the message is successfully sent, the function returns nil.
func PostToChat(message string, chatId string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	_, err := redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: chatId,
		Values: map[string]interface{}{
			"message": message,
		},
	}).Result()

	if err != nil {
		return err
	}
	return nil
}
