package types

import "encoding/hex"

type Address [20]uint8

func (a Address) String() string {
	return hex.EncodeToString(a[:])
}
