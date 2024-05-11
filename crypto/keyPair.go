package crypto

import (
	"blockchain/types"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

type PrivateKey struct {
	D              *big.Int
	PublicKeyBytes []byte
}

func GeneratePrivateKey() *PrivateKey {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		panic("couldn't convert public key to ECDSA public key")
	}

	return &PrivateKey{
		D:              privateKey.D,
		PublicKeyBytes: PublicKeyToBytes(publicKeyECDSA),
	}
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
		PublicKey: *BytesToPublicKey(priv.PublicKeyBytes),
		D:         priv.D,
	}
}

func (priv PrivateKey) PublicKey() PublicKey {
	return PublicKey{
		PublicKeyBytes: priv.PublicKeyBytes,
	}
}

func (priv PrivateKey) Sign(data []byte) (*Signature, error) {
	r, s, err := ecdsa.Sign(rand.Reader, priv.Key(), data)
	if err != nil {
		return nil, err
	}

	return &Signature{r, s}, nil
}

type PublicKey struct {
	PublicKeyBytes []byte
}

func (pub PublicKey) Key() *ecdsa.PublicKey {
	return BytesToPublicKey(pub.PublicKeyBytes)
}

func (pub PublicKey) Address() types.Address {
	hash := sha256.Sum256(pub.PublicKeyBytes)
	// return last 20 elements of the array
	return types.Address(hash[12:])
}

type Signature struct {
	R, S *big.Int
}

func (s *Signature) Verify(pub PublicKey, data []byte) bool {
	return ecdsa.Verify(pub.Key(), data, s.R, s.S)
}
