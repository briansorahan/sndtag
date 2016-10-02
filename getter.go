package sndtag

import (
	"fmt"
	"io"
)

// Getter gets a named property from an audio file's metadata.
// If the property does not exist, then an error is returned.
type Getter interface {
	Get(string) (string, error)
}

// NewGetter creates a new Getter.
// If the type is not one of the supported types then an error is returned.
func NewGetter(r io.Reader) (Getter, error) {
	// Read the first 3 bytes.
	header := make([]byte, 3)

	bytesRead, err := r.Read(header)
	if err != nil {
		return nil, err
	}
	if expected, got := 3, bytesRead; expected != got {
		return nil, fmt.Errorf("expected to read %d bytes, actually read %d", expected, got)
	}

	// Figure out the type.
	switch x := string(header); x {
	default:
		return nil, fmt.Errorf("unrecognized header: %s", x)
	case "TAG":
		// TODO: handle id3
		return newID3(r)
	case "RIF":
		// Read one more byte for RIFF type.
		headerLastByte := make([]byte, 1)

		bytesRead, err := r.Read(headerLastByte)
		if err != nil {
			return nil, err
		}
		if expected, got := 1, bytesRead; expected != got {
			return nil, fmt.Errorf("expected to read %d bytes, actually read %d", expected, got)
		}

		if headerLastByte[0] != 'F' {
			hdr := string(append(header, headerLastByte...))
			return nil, fmt.Errorf("expected RIFF, got %s", hdr)
		}
		return newWav(r)
	}
}
