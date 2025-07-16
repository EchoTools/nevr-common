package codec

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/gofrs/uuid/v5"
)

func TestXPID_UUID(t *testing.T) {
	type fields struct {
		PlatformCode PlatformCode
		AccountId    uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   uuid.UUID
	}{
		{
			name: "valid UUID",
			fields: fields{
				PlatformCode: 1,
				AccountId:    1,
			},
			want: uuid.FromStringOrNil("496d8944-6159-5c53-bdc8-1cab22f9d28d"),
		},
		{
			name: "invalid PlatformCode",
			fields: fields{
				PlatformCode: 0,
				AccountId:    12341234,
			},
			want: uuid.Nil,
		},
		{
			name: "invalid AccountId",
			fields: fields{
				PlatformCode: 4,
				AccountId:    0,
			},
			want: uuid.Nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xpi := &XPID{
				PlatformCode: tt.fields.PlatformCode,
				AccountId:    tt.fields.AccountId,
			}
			if got := xpi.UUID(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("XPID.UUID = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestXPID_Equal(t *testing.T) {

	xpid1 := XPID{
		PlatformCode: 1,
		AccountId:    1,
	}
	xpid2 := XPID{
		PlatformCode: 1,
		AccountId:    1,
	}

	if xpid1 != xpid2 {
		t.Errorf("XPID.Equal() = %v, want %v", xpid1, xpid2)
	}
}

func TestXPID_Equals(t *testing.T) {
	type fields struct {
		PlatformCode PlatformCode
		AccountId    uint64
	}
	type args struct {
		other XPID
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "valid",
			fields: fields{
				PlatformCode: 1,
				AccountId:    1,
			},
			args: args{
				XPID{
					PlatformCode: 1,
					AccountId:    1,
				},
			},
			want: true,
		},
		{
			name: "invalid PlatformCode",
			fields: fields{
				PlatformCode: 0,
				AccountId:    1,
			},
			args: args{
				XPID{
					PlatformCode: 1,
					AccountId:    1,
				},
			},
			want: false,
		},
		{
			name: "invalid AccountId",
			fields: fields{
				PlatformCode: 1,
				AccountId:    0,
			},
			args: args{
				XPID{
					PlatformCode: 1,
					AccountId:    1,
				},
			},
			want: false,
		},
		{
			name: "invalid",
			fields: fields{
				PlatformCode: 1,
				AccountId:    1,
			},
			args: args{
				XPID{
					PlatformCode: 2,
					AccountId:    2,
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xpi := &XPID{
				PlatformCode: tt.fields.PlatformCode,
				AccountId:    tt.fields.AccountId,
			}
			if got := xpi.Equals(tt.args.other); got != tt.want {
				t.Errorf("XPID.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestXPID_IsNil(t *testing.T) {
	type fields struct {
		PlatformCode PlatformCode
		AccountId    uint64
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "valid",
			fields: fields{
				PlatformCode: 0,
				AccountId:    0,
			},
			want: true,
		},
		{
			name: "invalid PlatformCode",
			fields: fields{
				PlatformCode: 0,
				AccountId:    1,
			},
			want: false,
		},
		{
			name: "invalid AccountId",
			fields: fields{
				PlatformCode: 1,
				AccountId:    0,
			},
			want: false,
		},
		{
			name: "invalid",
			fields: fields{
				PlatformCode: 1,
				AccountId:    1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xpi := &XPID{
				PlatformCode: tt.fields.PlatformCode,
				AccountId:    tt.fields.AccountId,
			}
			if got := xpi.IsNil(); got != tt.want {
				t.Errorf("XPID.IsNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestXPID_MarshalText(t *testing.T) {
	type fields struct {
		PlatformCode PlatformCode
		AccountId    uint64
	}
	tests := []struct {
		name    string
		fields  fields
		want    []byte
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				PlatformCode: 1,
				AccountId:    1,
			},
			want:    []byte("STM-1"),
			wantErr: false,
		},
		{
			name: "invalid PlatformCode",
			fields: fields{
				PlatformCode: 0,
				AccountId:    1,
			},
			want:    []byte("UNK-1"),
			wantErr: false,
		},
		{
			name: "invalid AccountId",
			fields: fields{
				PlatformCode: 1,
				AccountId:    0,
			},
			want:    []byte("STM-0"),
			wantErr: false,
		},
		{
			name: "invalid",
			fields: fields{
				PlatformCode: 0,
				AccountId:    0,
			},
			want:    []byte(""),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := XPID{
				PlatformCode: tt.fields.PlatformCode,
				AccountId:    tt.fields.AccountId,
			}
			got, err := e.MarshalText()
			if (err != nil) != tt.wantErr {
				t.Errorf("XPID.MarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("XPID.MarshalText() = `%v`, want `%v`", string(got), string(tt.want))
			}
		})
	}
}

func TestXPID_UnmarshalJSON(t *testing.T) {
	type fields struct {
		PlatformCode PlatformCode
		AccountId    uint64
	}
	type args struct {
		b []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "valid",
			fields: fields{
				PlatformCode: 1,
				AccountId:    1,
			},
			args: args{
				b: []byte(`"STM-1"`),
			},
			wantErr: false,
		},
		{
			name: "invalid PlatformCode",
			fields: fields{
				PlatformCode: 0,
				AccountId:    1,
			},
			args: args{
				b: []byte(`"UNK-1"`),
			},
			wantErr: false,
		},
		{
			name: "invalid AccountId",
			fields: fields{
				PlatformCode: 1,
				AccountId:    0,
			},
			args: args{
				b: []byte(`"STM-0"`),
			},
			wantErr: false,
		},
		{
			name: "invalid",
			fields: fields{
				PlatformCode: 0,
				AccountId:    0,
			},
			args: args{
				b: []byte(`""`),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &XPID{
				PlatformCode: tt.fields.PlatformCode,
				AccountId:    tt.fields.AccountId,
			}
			if err := json.Unmarshal(tt.args.b, e); (err != nil) != tt.wantErr {
				t.Errorf("XPID.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseXPID(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    *XPID
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				s: "STM-1",
			},
			want: &XPID{
				PlatformCode: 1,
				AccountId:    1,
			},
			wantErr: false,
		},
		{
			name: "invalid PlatformCode",
			args: args{
				s: "UNK-1",
			},
			want: &XPID{
				PlatformCode: 0,
				AccountId:    1,
			},
			wantErr: false,
		},
		{
			name: "invalid AccountId",
			args: args{
				s: "STM-0",
			},
			want: &XPID{
				PlatformCode: 1,
				AccountId:    0,
			},
			wantErr: false,
		},
		{
			name: "OVR_ORG-3963667097037078",
			args: args{
				s: "OVR_ORG-3963667097037078",
			},
			want: &XPID{
				PlatformCode: 4,
				AccountId:    3963667097037078,
			},
			wantErr: false,
		},
		{
			name: "OVR-ORG-3963667097037078",
			args: args{
				s: "OVR-ORG-3963667097037078",
			},
			want: &XPID{
				PlatformCode: 4,
				AccountId:    3963667097037078,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseXPID(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseXPID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseXPID() = %v, want %v", got, tt.want)
			}
		})
	}
}
