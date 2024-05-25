package core

import (
	"slices"
)

type Instruction byte

const (
	InstrPushInt Instruction = iota + 10 // 0-9 are reserved for digits
	InstrAdd
	InstrSub
	InstrMul
	InstrDiv
	InstrPushByte
	InstrPack
	InstrStore
	InstrGet
)

type Stack struct {
	data    []any
	pointer int
}

func NewStack(size int) *Stack {
	return &Stack{
		data:    make([]any, size),
		pointer: -1,
	}
}

func (s *Stack) Push(v any) {
	s.pointer++
	s.data[s.pointer] = v
}

func (s *Stack) Pop() any {
	value := s.data[s.pointer]
	slices.Delete(s.data, s.pointer, s.pointer+1)
	s.pointer--
	return value
}

// VM is virtual machine
type VM struct {
	data          []byte
	pointer       int
	stack         *Stack
	contractState *State
}

func NewVM(data []byte, contractState *State) *VM {
	return &VM{
		data:          data,
		stack:         NewStack(128),
		contractState: contractState,
	}
}

func (vm *VM) Run() error {
	for {
		instr := Instruction(vm.data[vm.pointer])
		if err := vm.Exec(instr); err != nil {
			return err
		}
		vm.pointer++
		if vm.pointer > len(vm.data)-1 {
			break
		}
	}
	return nil
}

func (vm *VM) Exec(instr Instruction) error {
	switch instr {
	case InstrPushInt:
		// "vm.data[vm.instrPointer]-1" is the byte that is pushed to the stack
		// which comes before "InstrPush" command
		vm.stack.Push(int(vm.data[vm.pointer-1]))
	case InstrAdd:
		b := vm.stack.Pop().(int)
		a := vm.stack.Pop().(int)
		c := a + b
		vm.stack.Push(c)
	case InstrSub:
		b := vm.stack.Pop().(int)
		a := vm.stack.Pop().(int)
		c := a - b
		vm.stack.Push(c)
	case InstrMul:
		b := vm.stack.Pop().(int)
		a := vm.stack.Pop().(int)
		c := a * b
		vm.stack.Push(c)
	case InstrDiv:
		b := vm.stack.Pop().(int)
		a := vm.stack.Pop().(int)
		c := a / b
		vm.stack.Push(c)
	case InstrPushByte:
		vm.stack.Push(vm.data[vm.pointer-1])
	// pack into array
	case InstrPack:
		n := vm.stack.Pop().(int)
		b := make([]byte, n)
		for i := n - 1; i >= 0; i-- {
			b[i] = vm.stack.Pop().(byte)
		}
		vm.stack.Push(b)
	case InstrStore:
		key := vm.stack.Pop().([]byte)
		value := vm.stack.Pop()

		var serializedValue []byte
		switch v := value.(type) {
		case int:
			serializedValue = serializeInt64(int64(v))
		default:
			panic("VM.Exec: unknown type")
		}

		vm.contractState.Add(key, serializedValue)
	case InstrGet:
		key := vm.stack.Pop().([]byte)
		value, err := vm.contractState.Get(key)
		if err != nil {
			return err
		}
		vm.stack.Push(value)
	}
	return nil
}
