package network

import (
	"blockchain/core"
	"encoding/gob"
	"io"
)

type Status struct {
	// ID of the server
	ID      string
	Version uint32
	Height  uint32
}

func (s *Status) Encode(enc core.Encoder[*Status]) error {
	return enc.Encode(s)
}

func (s *Status) Decode(dec core.Decoder[*Status]) error {
	return dec.Decode(s)
}

type GobStatusEncoder struct {
	w io.Writer
}

func NewGobStatusEncoder(w io.Writer) *GobStatusEncoder {
	return &GobStatusEncoder{
		w: w,
	}
}

func (e *GobStatusEncoder) Encode(s *Status) error {
	return gob.NewEncoder(e.w).Encode(s)
}

type GobStatusDecoder struct {
	r io.Reader
}

func NewGobStatusDecoder(r io.Reader) *GobStatusDecoder {
	return &GobStatusDecoder{
		r: r,
	}
}

func (d *GobStatusDecoder) Decode(s *Status) error {
	return gob.NewDecoder(d.r).Decode(s)
}

type EmptyMessage struct {
	Type MessageType
}

func (m *EmptyMessage) Encode(enc core.Encoder[*EmptyMessage]) error {
	return enc.Encode(m)
}

func (m *EmptyMessage) Decode(dec core.Decoder[*EmptyMessage]) error {
	return dec.Decode(m)
}

type GobEmptyMessageEncoder struct {
	w io.Writer
}

func NewGobEmptyMessageEncoder(w io.Writer) *GobEmptyMessageEncoder {
	return &GobEmptyMessageEncoder{
		w: w,
	}
}

func (e *GobEmptyMessageEncoder) Encode(m *EmptyMessage) error {
	return gob.NewEncoder(e.w).Encode(m)
}

type GobEmptyMessageDecoder struct {
	r io.Reader
}

func NewGobEmptyMessageDecoder(r io.Reader) *GobEmptyMessageDecoder {
	return &GobEmptyMessageDecoder{
		r: r,
	}
}

func (d *GobEmptyMessageDecoder) Decode(m *EmptyMessage) error {
	return gob.NewDecoder(d.r).Decode(m)
}

type SyncBlocksRequest struct {
	FromHeight uint32
	ToHeight   uint32
}

func (r *SyncBlocksRequest) Encode(enc core.Encoder[*SyncBlocksRequest]) error {
	return enc.Encode(r)
}

func (r *SyncBlocksRequest) Decode(dec core.Decoder[*SyncBlocksRequest]) error {
	return dec.Decode(r)
}

type GobSyncBlocksRequestEncoder struct {
	w io.Writer
}

func NewGobSyncBlocksRequestEncoder(w io.Writer) *GobSyncBlocksRequestEncoder {
	return &GobSyncBlocksRequestEncoder{
		w: w,
	}
}

func (e *GobSyncBlocksRequestEncoder) Encode(r *SyncBlocksRequest) error {
	return gob.NewEncoder(e.w).Encode(r)
}

type GobSyncBlocksRequestDecoder struct {
	r io.Reader
}

func NewGobSyncBlocksRequestDecoder(r io.Reader) *GobSyncBlocksRequestDecoder {
	return &GobSyncBlocksRequestDecoder{
		r: r,
	}
}

func (d *GobSyncBlocksRequestDecoder) Decode(r *SyncBlocksRequest) error {
	return gob.NewDecoder(d.r).Decode(r)
}

type Blocks []*core.Block

func (b *Blocks) Encode(enc core.Encoder[*Blocks]) error {
	return enc.Encode(b)
}

func (b *Blocks) Decode(dec core.Decoder[*Blocks]) error {
	return dec.Decode(b)
}

type GobBlocksEncoder struct {
	w io.Writer
}

func NewGobBlocksEncoder(w io.Writer) *GobBlocksEncoder {
	return &GobBlocksEncoder{
		w: w,
	}
}

func (e *GobBlocksEncoder) Encode(b *Blocks) error {
	return gob.NewEncoder(e.w).Encode(b)
}

type GobBlocksDecoder struct {
	r io.Reader
}

func NewGobBlocksDecoder(r io.Reader) *GobBlocksDecoder {
	return &GobBlocksDecoder{
		r: r,
	}
}

func (d *GobBlocksDecoder) Decode(b *Blocks) error {
	return gob.NewDecoder(d.r).Decode(b)
}
