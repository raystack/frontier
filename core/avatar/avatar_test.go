package avatar

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	jpeg := jpegDataURLPrefix + base64.StdEncoding.EncodeToString([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10})
	png := pngDataURLPrefix + base64.StdEncoding.EncodeToString([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	pngBytesAsJPEG := jpegDataURLPrefix + base64.StdEncoding.EncodeToString([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	notAnImage := jpegDataURLPrefix + base64.StdEncoding.EncodeToString([]byte("hello there"))
	oversized := jpegDataURLPrefix + strings.Repeat("A", DefaultMaxBytes)

	tests := []struct {
		name    string
		value   string
		cfg     Config
		wantErr bool
	}{
		{name: "empty is allowed", value: "", wantErr: false},
		{name: "valid jpeg base64", value: jpeg, wantErr: false},
		{name: "valid png base64", value: png, wantErr: false},
		{name: "zero limit falls back to default", value: jpeg, cfg: Config{MaxSizeBytes: 0}, wantErr: false},
		{name: "valid jpeg rejected under tiny limit", value: jpeg, cfg: Config{MaxSizeBytes: 8}, wantErr: true},
		{name: "rejects svg data url", value: "data:image/svg+xml;base64,PHN2ZyBvbmxvYWQ9ImFsZXJ0KDEpIi8+", wantErr: true},
		{name: "rejects javascript scheme", value: "javascript:alert(document.cookie)", wantErr: true},
		{name: "rejects http url", value: "http://169.254.169.254/latest/meta-data/", wantErr: true},
		{name: "rejects https url", value: "https://attacker.example/collect?u=admin", wantErr: true},
		{name: "rejects oversized payload", value: oversized, wantErr: true},
		{name: "rejects invalid base64", value: jpegDataURLPrefix + "@@@not-base64@@@", wantErr: true},
		{name: "rejects png bytes declared as jpeg", value: pngBytesAsJPEG, wantErr: true},
		{name: "rejects non-image bytes", value: notAnImage, wantErr: true},
		{name: "rejects data url without base64", value: "data:image/jpeg,rawdata", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.value, tt.cfg)
			if tt.wantErr {
				assert.ErrorIs(t, err, ErrInvalid)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
