package sndtag

import (
	"encoding/binary"
	"fmt"
	"io"
)

// wav parses RIFF tags from wav files.
type wav struct {
	length int32

	store map[string]string
}

// Get returns the tag specified by key.
// If the tag doesn't exist an error is returned.
func (w wav) Get(key string) (string, error) {
	v, ok := w.store[key]
	if !ok {
		return "", fmt.Errorf("key does not exist: %s", key)
	}
	return v, nil
}

// sniffFormat reads the format bytes from an io.Reader and returns
// an error if they aren't "WAVE".
func (w wav) sniffFormat(r io.Reader) error {
	format := make([]byte, 4)
	bytesRead, err := r.Read(format)
	if err != nil {
		return err
	}
	if expected, got := 4, bytesRead; expected != got {
		return fmt.Errorf("expected to read %d bytes, actually read %d", expected, got)
	}
	if expected, got := "WAVE", string(format); expected != got {
		return fmt.Errorf("expected format %s, got %s", expected, got)
	}
	return nil
}

// newWav creates a new Getter that can return properties for WAV files.
// Note that the "RIFF" chunk identifier has already been read
// by the time this function is called.
func newWav(r io.Reader) (Getter, error) {
	getter := wav{
		store: map[string]string{},
	}

	// Get the length.
	if err := binary.Read(r, binary.LittleEndian, &getter.length); err != nil {
		return nil, err
	}

	// Sniff the format.
	if err := getter.sniffFormat(r); err != nil {
		return nil, err
	}

	return getter, nil
}

// chunk is a RIFF chunk.
type chunk struct {
	ID   [4]byte
	Len  [4]byte
	Data []byte
}
