package network

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessage_EmptyMessageEncodeDecode(t *testing.T) {
	msg := &EmptyMessage{
		Type: MessageTypeStatusRequest,
	}
	buf := new(bytes.Buffer)
	assert.Nil(t, msg.Encode(NewGobEmptyMessageEncoder(buf)))

	msgDecoded := new(EmptyMessage)
	assert.Nil(t, msgDecoded.Decode(NewGobEmptyMessageDecoder(buf)))
	assert.Equal(t, msg, msgDecoded)
}
