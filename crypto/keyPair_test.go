package crypto

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSignVerifySuccess(t *testing.T) {
	privateKey := GeneratePrivateKey()
	publicKey := privateKey.PublicKey()

	msg := []byte("hey")
	sig, err := privateKey.Sign(msg)
	assert.Nil(t, err)
	assert.True(t, sig.Verify(publicKey, msg))
}

func TestSignVerifyFail(t *testing.T) {
	privateKey1 := GeneratePrivateKey()
	publicKey1 := privateKey1.PublicKey()
	privateKey2 := GeneratePrivateKey()
	publicKey2 := privateKey2.PublicKey()

	msg := []byte("hey")
	sig, err := privateKey1.Sign(msg)
	assert.Nil(t, err)
	assert.False(t, sig.Verify(publicKey2, msg))
	assert.False(t, sig.Verify(publicKey1, []byte("hey!")))
}
