package core

import (
	"encoding/gob"
	"io"
)

type Encoder[T any] interface {
	Encode(T) error
}

type Decoder[T any] interface {
	Decode(T) error
}

type GobTransactionEncoder struct {
	w io.Writer
}

func NewGobTransactionEncoder(w io.Writer) *GobTransactionEncoder {
	return &GobTransactionEncoder{
		w: w,
	}
}

func (e *GobTransactionEncoder) Encode(tx *Transaction) error {
	return gob.NewEncoder(e.w).Encode(tx)
}

type GobTransactionDecoder struct {
	r io.Reader
}

func NewGobTransactionDecoder(r io.Reader) *GobTransactionDecoder {
	return &GobTransactionDecoder{
		r: r,
	}
}

func (d *GobTransactionDecoder) Decode(tx *Transaction) error {
	return gob.NewDecoder(d.r).Decode(tx)
}

type GobBlockEncoder struct {
	w io.Writer
}

func NewGobBlockEncoder(w io.Writer) *GobBlockEncoder {
	return &GobBlockEncoder{
		w: w,
	}
}

func (e *GobBlockEncoder) Encode(b *Block) error {
	return gob.NewEncoder(e.w).Encode(b)
}

type GobBlockDecoder struct {
	r io.Reader
}

func NewGobBlockDecoder(r io.Reader) *GobBlockDecoder {
	return &GobBlockDecoder{
		r: r,
	}
}

func (d *GobBlockDecoder) Decode(b *Block) error {
	return gob.NewDecoder(d.r).Decode(b)
}
