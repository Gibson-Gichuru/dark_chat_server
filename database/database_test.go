package database

import (
	"testing"

	"github.com/google/uuid"
)

func TestClientRegister(t *testing.T) {

	clientID := uuid.NewString()

	err := RegisterClientChat(clientID)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
