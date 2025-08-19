package telemetry

import (
	"io"
	"os"

	"github.com/klauspost/compress/zstd"
	"google.golang.org/protobuf/proto"
)

// ZstdCodec handles streaming to/from Zstd-compressed .nevrcap files
type ZstdCodec struct {
	file     *os.File
	encoder  *zstd.Encoder
	decoder  *zstd.Decoder
	writer   io.Writer
	reader   io.Reader
}

// NewZstdCodecWriter creates a new Zstd codec for writing .nevrcap files
func NewZstdCodecWriter(filename string) (*ZstdCodec, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	encoder, err := zstd.NewWriter(file, zstd.WithEncoderLevel(zstd.SpeedFastest))
	if err != nil {
		file.Close()
		return nil, err
	}

	return &ZstdCodec{
		file:    file,
		encoder: encoder,
		writer:  encoder,
	}, nil
}

// NewZstdCodecReader creates a new Zstd codec for reading .nevrcap files
func NewZstdCodecReader(filename string) (*ZstdCodec, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	decoder, err := zstd.NewReader(file)
	if err != nil {
		file.Close()
		return nil, err
	}

	return &ZstdCodec{
		file:    file,
		decoder: decoder,
		reader:  decoder,
	}, nil
}

// WriteHeader writes the telemetry header to the file
func (z *ZstdCodec) WriteHeader(header *TelemetryHeader) error {
	data, err := proto.Marshal(header)
	if err != nil {
		return err
	}

	// Write length-delimited message
	return z.writeDelimitedMessage(data)
}

// WriteFrame writes a frame to the file
func (z *ZstdCodec) WriteFrame(frame *LobbySessionStateFrame) error {
	data, err := proto.Marshal(frame)
	if err != nil {
		return err
	}

	// Write length-delimited message
	return z.writeDelimitedMessage(data)
}

// ReadHeader reads the telemetry header from the file
func (z *ZstdCodec) ReadHeader() (*TelemetryHeader, error) {
	data, err := z.readDelimitedMessage()
	if err != nil {
		return nil, err
	}

	header := &TelemetryHeader{}
	err = proto.Unmarshal(data, header)
	if err != nil {
		return nil, err
	}

	return header, nil
}

// ReadFrame reads a frame from the file
func (z *ZstdCodec) ReadFrame() (*LobbySessionStateFrame, error) {
	data, err := z.readDelimitedMessage()
	if err != nil {
		return nil, err
	}

	frame := &LobbySessionStateFrame{}
	err = proto.Unmarshal(data, frame)
	if err != nil {
		return nil, err
	}

	return frame, nil
}

// writeDelimitedMessage writes a length-delimited protobuf message
func (z *ZstdCodec) writeDelimitedMessage(data []byte) error {
	// Write varint length
	length := uint64(len(data))
	for length >= 0x80 {
		if _, err := z.writer.Write([]byte{byte(length) | 0x80}); err != nil {
			return err
		}
		length >>= 7
	}
	if _, err := z.writer.Write([]byte{byte(length)}); err != nil {
		return err
	}

	// Write message data
	_, err := z.writer.Write(data)
	return err
}

// readDelimitedMessage reads a length-delimited protobuf message
func (z *ZstdCodec) readDelimitedMessage() ([]byte, error) {
	// Read varint length
	var length uint64
	var shift uint
	for {
		b := make([]byte, 1)
		if _, err := z.reader.Read(b); err != nil {
			return nil, err
		}
		
		length |= uint64(b[0]&0x7F) << shift
		if b[0]&0x80 == 0 {
			break
		}
		shift += 7
		if shift >= 64 {
			return nil, io.ErrUnexpectedEOF
		}
	}

	// Read message data
	data := make([]byte, length)
	_, err := io.ReadFull(z.reader, data)
	return data, err
}

// Close closes the codec and underlying file
func (z *ZstdCodec) Close() error {
	var err error
	
	if z.encoder != nil {
		err = z.encoder.Close()
	}
	
	if z.decoder != nil {
		z.decoder.Close()
	}
	
	if z.file != nil {
		if closeErr := z.file.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}
	
	return err
}