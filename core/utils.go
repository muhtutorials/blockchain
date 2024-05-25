package core

import (
	"encoding/binary"
)

type Instr struct {
	data []byte
}

// Bytes returns composed instructions
func (i *Instr) Bytes() []byte {
	return i.data
}

// Add composes "add two integers and push result to the stack" instruction
func (i *Instr) Add(a, b int) *Instr {
	ins := []byte{
		byte(a),
		byte(InstrPushInt),
		byte(b),
		byte(InstrPushInt),
		byte(InstrAdd),
	}
	i.data = append(i.data, ins...)
	return i
}

// Sub composes "subtract two integers and push result to the stack" instruction
func (i *Instr) Sub(a, b int) *Instr {
	ins := []byte{
		byte(a),
		byte(InstrPushInt),
		byte(b),
		byte(InstrPushInt),
		byte(InstrSub),
	}
	i.data = append(i.data, ins...)
	return i
}

// Mul composes "multiply two integers and push result to the stack" instruction
func (i *Instr) Mul(a, b int) *Instr {
	ins := []byte{
		byte(a),
		byte(InstrPushInt),
		byte(b),
		byte(InstrPushInt),
		byte(InstrMul),
	}
	i.data = append(i.data, ins...)
	return i
}

// Div composes "divide two integers and push result to the stack" instruction
func (i *Instr) Div(a, b int) *Instr {
	ins := []byte{
		byte(a),
		byte(InstrPushInt),
		byte(b),
		byte(InstrPushInt),
		byte(InstrDiv),
	}
	i.data = append(i.data, ins...)
	return i
}

// String composes "push string to the stack" instruction
func (i *Instr) String(str string) *Instr {
	for _, char := range []byte(str) {
		i.data = append(i.data, char, byte(InstrPushByte))
	}
	i.data = append(i.data, byte(len(str)), byte(InstrPushInt), byte(InstrPack))
	return i
}

// Store composes "store key-value pair to the state" instruction
func (i *Instr) Store() *Instr {
	i.data = append(i.data, byte(InstrStore))
	return i
}

// Get composes "get value by key from the state ant push it to the stack" instruction
func (i *Instr) Get(str string) *Instr {
	i.String(str)
	i.data = append(i.data, byte(InstrGet))
	return i
}

func serializeInt64(value int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

func deserializeInt64(b []byte) int64 {
	value := binary.LittleEndian.Uint64(b)
	return int64(value)
}
