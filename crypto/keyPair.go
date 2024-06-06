package crypto

import (
	"blockchain/types"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"math/big"
)

type PrivateKey struct {
	D         *big.Int
	publicKey []byte
}

func NewPrivateKey(r io.Reader) *PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), r)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("couldn't convert public key to ECDSA public key")
	}

	return &PrivateKey{
		D:         privateKey.D,
		publicKey: PublicKeyToBytes(publicKeyECDSA),
	}
}

func GeneratePrivateKey() *PrivateKey {
	return NewPrivateKey(rand.Reader)
}

func PublicKeyToBytes(pub *ecdsa.PublicKey) []byte {
	return elliptic.MarshalCompressed(pub, pub.X, pub.Y)
}

func BytesToPublicKey(b []byte) *ecdsa.PublicKey {
	curve := elliptic.P256()
	x, y := elliptic.UnmarshalCompressed(curve, b)
	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}
}

// Key return decoded private key
func (priv PrivateKey) Key() *ecdsa.PrivateKey {
	return &ecdsa.PrivateKey{
		PublicKey: *BytesToPublicKey(priv.publicKey),
		D:         priv.D,
	}
}

func (priv PrivateKey) PublicKey() PublicKey {
	return priv.publicKey
}

func (priv PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priv.Key(), data)
	if err != nil {
		return nil, err
	}

	return &Signature{r, s}, nil
}

type PublicKey []byte

func (pub PublicKey) Key() *ecdsa.PublicKey {
	return BytesToPublicKey(pub)
}

func (pub PublicKey) Address() types.Address {
	hash := sha256.Sum256(pub)
	// return last 20 elements of the array
	return types.Address(hash[12:])
}

func (pub PublicKey) String() string {
	return hex.EncodeToString(pub)
}

type Signature struct {
	R, S *big.Int
}

func (s *Signature) Verify(pub PublicKey, data []byte) bool {
	return ecdsa.Verify(pub.Key(), data, s.R, s.S)
}
