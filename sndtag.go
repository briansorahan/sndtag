package sndtag

import (
	"fmt"
	"io"
)

// Types of tags that are supported.
// TODO: support id3
const (
	RIFF = iota
	ID3v1
	ID3v2
)

// New creates a new map with metadata read from an io.Reader.
// If the type is not one of the supported types then an error is returned.
func New(r io.Reader) (map[string]string, error) {
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
		if err := checkRIFFLastByte(r, header); err != nil {
			return nil, err
		}

		getter, err := newWav(r)
		if err != nil && err != io.EOF {
			return nil, err
		}
		return getter, nil
	}
}

// checkRIFFLastByte checks that the 4th byte of a RIFF file is 'F'.
func checkRIFFLastByte(r io.Reader, header []byte) error {
	// Read one more byte for RIFF type.
	headerLastByte := make([]byte, 1)

	bytesRead, err := r.Read(headerLastByte)
	if err != nil {
		return err
	}
	if expected, got := 1, bytesRead; expected != got {
		return fmt.Errorf("expected to read %d bytes, actually read %d", expected, got)
	}

	if headerLastByte[0] != 'F' {
		hdr := string(append(header, headerLastByte...))
		return fmt.Errorf("expected RIFF, got %s", hdr)
	}
	return nil
}
