package nevrcap

import (
	"bytes"
	"io"
	"testing"
)

func TestZstdCodec_writeDelimitedMessage(t *testing.T) {
	tests := []struct {
		name    string
		message []byte
	}{
		{"empty message", []byte{}},
		{"short message", []byte{0x01, 0x02, 0x03}},
		{"long message", bytes.Repeat([]byte{0xAB}, 300)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			codec := &ZstdCodec{writer: &buf}
			err := codec.writeDelimitedMessage(tt.message)
			if err != nil {
				t.Fatalf("writeDelimitedMessage() error = %v", err)
			}

			// Now decode the varint length
			var (
				length uint64
				shift  uint
			)
			for i := 0; i < 10; i++ {
				b, err := buf.ReadByte()
				if err != nil {
					t.Fatalf("failed to read varint: %v", err)
				}
				length |= uint64(b&0x7F) << shift
				if b&0x80 == 0 {
					break
				}
				shift += 7
			}
			if int(length) != len(tt.message) {
				t.Errorf("length mismatch: got %d, want %d", length, len(tt.message))
			}
			got := make([]byte, length)
			_, err = io.ReadFull(&buf, got)
			if err != nil {
				t.Fatalf("failed to read message: %v", err)
			}
			if !bytes.Equal(got, tt.message) {
				t.Errorf("message mismatch: got %v, want %v", got, tt.message)
			}
		})
	}
}

func BenchmarkZstdCodec_writeDelimitedMessage(b *testing.B) {
	msg := bytes.Repeat([]byte{0x42}, 1024)
	var buf bytes.Buffer
	codec := &ZstdCodec{writer: &buf}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		if err := codec.writeDelimitedMessage(msg); err != nil {
			b.Fatal(err)
		}
	}
}
