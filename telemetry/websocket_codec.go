package telemetry

import (
	"context"
	"encoding/json"
	"log"
	"net/url"
	"time"

	"nhooyr.io/websocket"
	"google.golang.org/protobuf/proto"
)

// WebSocketCodecOptions configures what data to send over WebSocket
type WebSocketCodecOptions struct {
	SendSessionData    bool // Send raw session data as bytes
	SendUserBones      bool // Send raw user bones data as bytes
	SendEvents         bool // Process events and send LobbySessionStateFrame
	ProcessFrames      bool // Whether to process frames for event detection
}

// WebSocketCodec handles streaming LobbySessionStateFrame objects over WebSocket
type WebSocketCodec struct {
	conn           *websocket.Conn
	ctx            context.Context
	options        WebSocketCodecOptions
	frameProcessor *FrameProcessor
}

// NewWebSocketCodec creates a new WebSocket codec
func NewWebSocketCodec(serverURL string, options WebSocketCodecOptions) (*WebSocketCodec, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	conn, _, err := websocket.Dial(ctx, u.String(), nil)
	if err != nil {
		return nil, err
	}

	var frameProcessor *FrameProcessor
	if options.ProcessFrames {
		frameProcessor = NewFrameProcessor()
	}

	return &WebSocketCodec{
		conn:           conn,
		ctx:            ctx,
		options:        options,
		frameProcessor: frameProcessor,
	}, nil
}

// StreamRawData sends raw session and user bones data
func (ws *WebSocketCodec) StreamRawData(sessionData, userBonesData []byte, timestamp time.Time) error {
	if ws.options.SendSessionData {
		message := map[string]interface{}{
			"type":      "session_data",
			"timestamp": timestamp.Unix(),
			"data":      sessionData,
		}
		
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		
		err = ws.conn.Write(ws.ctx, websocket.MessageText, data)
		if err != nil {
			return err
		}
	}

	if ws.options.SendUserBones && len(userBonesData) > 0 {
		message := map[string]interface{}{
			"type":      "user_bones",
			"timestamp": timestamp.Unix(),
			"data":      userBonesData,
		}
		
		data, err := json.Marshal(message)
		if err != nil {
			return err
		}
		
		err = ws.conn.Write(ws.ctx, websocket.MessageText, data)
		if err != nil {
			return err
		}
	}

	return nil
}

// StreamProcessedFrame processes data into LobbySessionStateFrame and sends it
func (ws *WebSocketCodec) StreamProcessedFrame(sessionData, userBonesData []byte, timestamp time.Time) error {
	if !ws.options.ProcessFrames || ws.frameProcessor == nil {
		return nil
	}

	frame, err := ws.frameProcessor.ProcessFrame(sessionData, userBonesData, timestamp)
	if err != nil {
		return err
	}

	return ws.StreamFrame(frame)
}

// StreamFrame sends a processed LobbySessionStateFrame
func (ws *WebSocketCodec) StreamFrame(frame *LobbySessionStateFrame) error {
	if ws.options.SendEvents {
		// Send as protobuf binary
		data, err := proto.Marshal(frame)
		if err != nil {
			return err
		}

		return ws.conn.Write(ws.ctx, websocket.MessageBinary, data)
	}

	return nil
}

// ReceiveFrame receives a LobbySessionStateFrame from the WebSocket
func (ws *WebSocketCodec) ReceiveFrame() (*LobbySessionStateFrame, error) {
	_, data, err := ws.conn.Read(ws.ctx)
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

// Close closes the WebSocket connection
func (ws *WebSocketCodec) Close() error {
	return ws.conn.Close(websocket.StatusNormalClosure, "")
}

// WebSocketServer provides a simple WebSocket server for testing
type WebSocketServer struct {
	port string
}

// NewWebSocketServer creates a new WebSocket server
func NewWebSocketServer(port string) *WebSocketServer {
	return &WebSocketServer{port: port}
}

// Start starts the WebSocket server (simplified implementation for testing)
func (s *WebSocketServer) Start() error {
	// This is a simplified implementation
	// In a real implementation, you'd use a proper HTTP server with WebSocket upgrade
	log.Printf("WebSocket server would start on port %s", s.port)
	return nil
}