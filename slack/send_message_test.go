package slack

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSendMessage(t *testing.T) {
	err := SendMessage("Hello from github.com/dittotrade/internal/slack unit test!")
	require.NoError(t, err)
}
