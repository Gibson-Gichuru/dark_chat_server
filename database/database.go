package database

import (
	"context"
	"darkchat/monitor"
	"darkchat/protocol"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client
var databaseMonitor *monitor.Monitor

const (
	StreamNamePrefix   = "stream"
	GroupNamePrefix    = "group"
	ChatsPrefix        = "chats"
	ConsumerNamePrefix = "consumer"
)

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

// RegisterClientChat creates a Redis Stream and Consumer Group for the given chatId
// if they do not already exist. It returns an error if the Redis command fails.
// If the error is that the Consumer Group name already exists, the function logs
// an info message and returns the error. The function times out after 30 seconds.
// If the timeout is exceeded, the context is canceled and an error is returned.
// If the chatId is successfully registered, the function returns nil.
func RegisterClientChat(chatId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	err := redisClient.XGroupCreateMkStream(
		ctx,
		fmt.Sprintf("%s:%s", StreamNamePrefix, chatId),
		fmt.Sprintf("%s:%s", GroupNamePrefix, chatId),
		"0",
	).Err()

	if err != nil && strings.Contains(err.Error(), "BUSYGROUP Consumer Group name already exists") {
		databaseMonitor.Info(fmt.Sprintf("Stream already exists: %s", chatId))
		return err
	}

	_, err = redisClient.SAdd(
		ctx,
		fmt.Sprintf("%s:online", ChatsPrefix),
		fmt.Sprintf("stream:%s", chatId),
	).Result()

	if err != nil {
		return err
	}

	return nil

}

// DeleteClientChat removes a Redis Stream for the given chatId, and removes the
// chatId from the set of online chats. The function times out after 5 seconds,
// and returns an error if there was an error communicating with Redis.
func DeleteClientChat(chatId string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	err := redisClient.XGroupDestroy(
		ctx,
		fmt.Sprintf("%s:%s", StreamNamePrefix, chatId),
		fmt.Sprintf("%s:%s", GroupNamePrefix, chatId),
	).Err()

	if err != nil {
		return err
	}

	err = redisClient.Del(ctx, fmt.Sprintf("%s:%s", StreamNamePrefix, chatId)).Err()

	if err != nil {
		return err
	}

	err = redisClient.SRem(
		ctx,
		fmt.Sprintf("%s:online", ChatsPrefix),
		fmt.Sprintf("stream:%s", chatId),
	).Err()

	if err != nil {
		return err
	}
	return nil
}

// StreamChat reads messages from Redis streams and sends them to the given channel. It subscribes to
// the given streams and reads messages from them. If the subscribe channel is closed, the function
// returns. If there is an error communicating with Redis, the error is logged to the database log.
// The function will continue to run until the subscribe channel is closed or there is an error
// communicating with Redis. The function times out after 100 milliseconds if there are no messages
// in any of the streams.
func StreamChat(ctx context.Context, chatChannel chan<- protocol.Payload, subscribe <-chan string, chatId string) {
	databaseCTX, cancel := context.WithCancel(context.Background())
	defer func() {
		close(chatChannel)
		cancel()
	}()

	activeStreams := make(map[string]bool)
	consumerName := fmt.Sprintf("%s:%s", ConsumerNamePrefix, uuid.NewString())
	groupName := fmt.Sprintf("%s:%s", GroupNamePrefix, chatId)

	for {

		select {
		case newSub := <-subscribe:
			streamName := fmt.Sprintf("%s:%s", StreamNamePrefix, newSub)
			if !activeStreams[streamName] {
				activeStreams[streamName] = true
			}

		case <-ctx.Done():
			return

		default:

			streams := make([]string, 0, len(activeStreams))

			for stream := range activeStreams {
				streams = append(streams, stream, ">")
			}

			if len(streams) == 0 {
				time.Sleep(100 * time.Millisecond)
				continue
			}

			args := &redis.XReadGroupArgs{
				Group:    groupName,
				Consumer: consumerName,
				Streams:  streams,
				Block:    100 * time.Millisecond,
			}

			result, err := redisClient.XReadGroup(databaseCTX, args).Result()

			if err != nil && err != redis.Nil {
				databaseMonitor.Error(err.Error())
				continue
			}
			if len(result) == 0 || len(result[0].Messages) == 0 {
				continue
			}

			messageString, ok := result[0].Messages[0].Values["message"].(string)
			if !ok {
				databaseMonitor.Error("Expected message to be a string")
				continue
			}

			var message protocol.Message

			err = json.Unmarshal([]byte(messageString), &message)
			if err != nil {
				databaseMonitor.Error(err.Error())
				continue
			}

			chatChannel <- &message

			err = redisClient.XAck(databaseCTX,
				result[0].Stream,
				fmt.Sprintf("%s:%s", GroupNamePrefix, chatId),
				result[0].Messages[0].ID).Err()

			if err != nil {
				databaseMonitor.Error(err.Error())
				return
			}

		}
	}

}

// PostToChat sends a message to a Redis Stream identified by the given chatId.
// If an error occurs while communicating with Redis, the error is returned.
// If the timeout (5 seconds) is exceeded, the context is canceled and an error is returned.
// If the message is successfully sent, the function returns nil.
func PostToChat(message string, chatId string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	_, err := redisClient.XAdd(ctx, &redis.XAddArgs{
		Stream: fmt.Sprintf("%s:%s", StreamNamePrefix, chatId),
		Values: map[string]interface{}{
			"message": message,
		},
	}).Result()

	if err != nil {
		return err
	}
	return nil
}

// CheckChatExists returns true if the given chatId exists in the Redis set "chats",
// and false otherwise. If an error occurs while communicating with Redis, the
// error is logged and false is returned. The function times out after 5 seconds.
func CheckChatExists(chatId string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	result, err := redisClient.SIsMember(
		ctx,
		fmt.Sprintf("%s:online", ChatsPrefix),
		fmt.Sprintf("stream:%s", chatId),
	).Result()

	if err != nil {
		databaseMonitor.Error(err.Error())
		return false
	}

	return result

}
