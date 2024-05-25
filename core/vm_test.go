package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStackPushPop(t *testing.T) {
	stack := NewStack(8)
	stack.Push(1)
	stack.Push(2)
	stack.Push("hey")
	assert.Equal(t, "hey", stack.Pop())
	assert.Equal(t, 2, stack.Pop())
}

func TestVM_InstrAdd(t *testing.T) {
	ins := new(Instr)
	ins.Add(1, 2)
	contractState := NewState()
	vm := NewVM(ins.Bytes(), contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop()
	assert.Equal(t, 3, result)
}

func TestVM_InstrSub(t *testing.T) {
	ins := new(Instr)
	ins.Sub(5, 3)
	contractState := NewState()
	vm := NewVM(ins.Bytes(), contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop()
	assert.Equal(t, 2, result)
}

func TestVM_InstrMul(t *testing.T) {
	ins := new(Instr)
	ins.Mul(4, 2)
	contractState := NewState()
	vm := NewVM(ins.Bytes(), contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop()
	assert.Equal(t, 8, result)
}

func TestVM_InstrDiv(t *testing.T) {
	ins := new(Instr)
	ins.Div(8, 2)
	contractState := NewState()
	vm := NewVM(ins.Bytes(), contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop()
	assert.Equal(t, 4, result)
}

func TestVM_InstrPushByteInstrPack(t *testing.T) {
	ins := new(Instr)
	ins.String("hey")
	contractState := NewState()
	vm := NewVM(ins.Bytes(), contractState)
	assert.Nil(t, vm.Run())

	result := vm.stack.Pop().([]byte)
	assert.Equal(t, "hey", string(result))
}

func TestVM_InstrStore(t *testing.T) {
	ins := new(Instr)
	ins.Add(5, 2).String("hey").Store()
	contractState := NewState()
	vm := NewVM(ins.Bytes(), contractState)
	assert.Nil(t, vm.Run())

	value, err := contractState.Get([]byte("hey"))
	assert.Nil(t, err)
	assert.Equal(t, int64(7), deserializeInt64(value))
}

func TestVM_InstrGet(t *testing.T) {
	ins := new(Instr)
	ins.Add(2, 3).String("hey").Store().Get("hey")
	contractState := NewState()
	vm := NewVM(ins.Bytes(), contractState)
	assert.Nil(t, vm.Run())
	assert.Equal(t, int64(5), deserializeInt64(vm.stack.Pop().([]byte)))
}
