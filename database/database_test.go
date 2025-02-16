package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

// TestClientRegister tests the RegisterClientChat function by creating a new client ID,
// registering the client chat, and verifying that the Redis		 stream group is created
// with the expected name. It checks for errors during registration and retrieval of
// stream group information, and asserts that the group name matches the expected format.

func TestClientRegister(t *testing.T) {

	ctx := context.Background()

	clientID := uuid.NewString()

	err := RegisterClientChat(clientID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	groups, err := redisClient.XInfoGroups(
		ctx,
		fmt.Sprintf("stream:%s", clientID),
	).Result()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	for _, group := range groups {

		if group.Name != fmt.Sprintf("group:%s", clientID) {
			t.Errorf("Expected group name to be %s, got %s", fmt.Sprintf("group:%s", clientID), group.Name)
		}
	}

	if !CheckChatExists(clientID) {
		t.Errorf("Expected chat to exist, but it does not")
	}
}

// TestClientDelete tests the DeleteClientChat function by registering a new
// client chat, deleting the chat, and verifying that the Redis stream group
// is deleted. It checks for errors during registration, deletion, and retrieval
// of stream group information, and asserts that the group is deleted.
func TestClientDelete(t *testing.T) {
	ctx := context.Background()

	clientID := uuid.NewString()

	err := RegisterClientChat(clientID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	err = DeleteClientChat(clientID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// make sure that the consumer group is deleted
	_, err = redisClient.XInfoGroups(
		ctx,
		fmt.Sprintf("stream:%s", clientID),
	).Result()

	if err == nil {
		t.Error("Expected ERR no such key but hot nil")
	}

	// make sure that the stream is removed from the online chats set

	if CheckChatExists(clientID) {
		t.Errorf("Expected chat to not exist, but it does")
	}

	// make sure that the stream does not exist

	if redisClient.Exists(ctx, fmt.Sprintf("stream:%s", clientID)).Val() == 1 {
		t.Errorf("Expected stream to not exist, but it does")
	}
}
