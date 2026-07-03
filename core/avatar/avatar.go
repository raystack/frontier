package avatar

import (
	"bytes"
	"encoding/base64"
	"errors"
	"strings"
)

// DefaultMaxBytes caps the stored avatar string at 1 MiB when no limit is
// configured. The limit is measured on the encoded data URL string.
const DefaultMaxBytes = 1 << 20

// Config holds avatar validation settings.
type Config struct {
	// MaxSizeBytes is the largest allowed size of a stored avatar string. An
	// avatar is a base64 data URL, so this is measured on the encoded string.
	MaxSizeBytes int `yaml:"max_size_bytes" mapstructure:"max_size_bytes" default:"1048576"`
}

const (
	jpegDataURLPrefix = "data:image/jpeg;base64,"
	pngDataURLPrefix  = "data:image/png;base64,"
)

// ErrInvalid is returned when an avatar is not an allowed image data URL.
var ErrInvalid = errors.New("avatar must be a base64 JPEG or PNG data URL")

var (
	jpegMagic = []byte{0xFF, 0xD8, 0xFF}
	pngMagic  = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
)

// Validate checks that an avatar is either empty or a base64 JPEG/PNG data
// URL within the size limit.
func Validate(value string, cfg Config) error {
	if value == "" {
		return nil
	}

	maxBytes := cfg.MaxSizeBytes
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}

	// Check the length before decoding so oversized input is rejected cheaply.
	if len(value) > maxBytes {
		return ErrInvalid
	}

	payload, magic, ok := splitDataURL(value)
	if !ok {
		return ErrInvalid
	}

	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return ErrInvalid
	}

	if !bytes.HasPrefix(raw, magic) {
		return ErrInvalid
	}
	return nil
}

// splitDataURL returns the base64 payload and the file signature the decoded
// bytes must start with. ok is false for anything but a JPEG or PNG data URL.
func splitDataURL(value string) (payload string, magic []byte, ok bool) {
	switch {
	case strings.HasPrefix(value, jpegDataURLPrefix):
		return value[len(jpegDataURLPrefix):], jpegMagic, true
	case strings.HasPrefix(value, pngDataURLPrefix):
		return value[len(pngDataURLPrefix):], pngMagic, true
	default:
		return "", nil, false
	}
}
