package main

import (
	"bytes"
	"errors"
	"time"

	"github.com/echotools/nevr-common/gameapi"
	"github.com/echotools/nevr-common/rtapi"
	"google.golang.org/protobuf/encoding/protojson"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

var (
	protojsonMarshaler = &protojson.MarshalOptions{
		UseProtoNames:   false,
		UseEnumNumbers:  true,
		EmitUnpopulated: false,
	}
	protojsonUnmarshaler = &protojson.UnmarshalOptions{
		DiscardUnknown: false,
	}
)

func MarshalReplay(buf *bytes.Buffer, frames []*rtapi.SessionUpdateMessage) error {

	for _, f := range frames {
		if _, err := buf.WriteString(f.Timestamp.AsTime().Format("2006/01/02 15:04:05.000")); err != nil {
			return errors.New("failed to write timestamp: " + err.Error())
		}
		if _, err := buf.WriteString("\t"); err != nil {
			return errors.New("failed to write tab after timestamp: " + err.Error())
		}
		data, err := protojsonMarshaler.Marshal(f.Session)
		if err != nil {
			return errors.New("failed to marshal session data: " + err.Error())
		}

		if _, err := buf.Write(data); err != nil {
			return errors.New("failed to write session data: " + err.Error())
		}
		if _, err := buf.WriteString("\n"); err != nil {
			return errors.New("failed to write newline after session data: " + err.Error())
		}
	}
	return nil
}

func UnmarshalReplay(data []byte, dst *[]*rtapi.SessionUpdateMessage) error {
	// Parse a line at a time
	lines := bytes.Split(data, []byte("\n"))
	*dst = make([]*rtapi.SessionUpdateMessage, 0, len(lines))
	for _, line := range lines {
		if len(line) == 0 {
			continue // Skip empty lines
		}
		parts := bytes.Split(line, []byte("\t"))
		if len(parts) < 2 {
			return errors.New("invalid frame format: expected at least 2 parts separated by tab")
		}
		// Parse timestamp
		// format is "2025/05/30 02:36:27.789"
		timestampStr := parts[0]
		ts, err := time.Parse("2006/01/02 15:04:05.000", string(timestampStr))
		if err != nil {
			return errors.New("invalid timestamp format: " + string(timestampStr))
		}

		sessionResponse := &gameapi.SessionResponse{}
		if err := protojsonUnmarshaler.Unmarshal(parts[1], sessionResponse); err != nil {
			return errors.New("failed to unmarshal session data: " + err.Error())
		}

		/*
			// Parse user bones if present (assuming JSON)
			if len(parts) > 2 {
				if err := f.UserBones.UnmarshalJSON(parts[2]); err != nil {
					return errors.New("failed to unmarshal user bones: " + err.Error())
				}
			}
		*/

		*dst = append(*dst, &rtapi.SessionUpdateMessage{
			Timestamp: &timestamppb.Timestamp{
				Seconds: ts.Unix(),
				Nanos:   int32(ts.Nanosecond()),
			},
			Session: sessionResponse,
			// UserBones: gameapi.UserBones{}, // Uncomment if UserBones is used
		})
	}

	return nil
}
