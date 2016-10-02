package sndtag

import (
	"encoding/binary"
	"fmt"
	"io"
	"strconv"
)

// wav parses RIFF tags from wav files.
// See http://soundfile.sapp.org/doc/WaveFormat/ for more info.
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

// readFormat reads the format chunk.
// It also stores the formatting information as properties.
func (w wav) readFormat(r io.Reader) error {
	// Read the format identifier, should be "fmt ".
	if err := w.readFormatID(r); err != nil {
		return err
	}

	// Read the length, should be 16 for PCM.
	if err := w.readFormatLength(r); err != nil {
		return err
	}

	// Read the audio format.
	if err := w.readAudioFormat(r); err != nil {
		return err
	}

	// Read the number of channels.
	if err := w.readNumChannels(r); err != nil {
		return err
	}

	// Read sample rate.

	return nil
}

// readFormatID reads the fmt chunk identifier.
func (w wav) readFormatID(r io.Reader) error {
	id := make([]byte, 4)
	bytesRead, err := r.Read(id)
	if err != nil {
		return err
	}
	if expected, got := 4, bytesRead; expected != got {
		return fmt.Errorf("expected to read %d bytes, actually read %d", expected, got)
	}
	if expected, got := "fmt ", string(id); expected != got {
		return fmt.Errorf("expected format %s, got %s", expected, got)
	}
	return nil
}

// readFormatLength reads the length of the fmt chunk.
func (w wav) readFormatLength(r io.Reader) error {
	var length int32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return err
	}
	if expected, got := int32(16), length; expected != got {
		return fmt.Errorf("expected fmt chunk length %d, got %d", expected, got)
	}
	return nil
}

// readAudioFormat reads the audio format from the fmt chunk
// and stores it as the "AudioFormat" property.
func (w wav) readAudioFormat(r io.Reader) error {
	var audioFormat int16
	if err := binary.Read(r, binary.LittleEndian, &audioFormat); err != nil {
		return err
	}
	if expected, got := int16(1), audioFormat; expected != got {
		return fmt.Errorf("expected pcm audio format %d, got %d", expected, got)
	}
	w.store["AudioFormat"] = strconv.FormatInt(int64(audioFormat), 10)
	return nil
}

// readNumChannels reads the number of channels from the fmt chunk
// and stores it as the "NumChannels" property.
func (w wav) readNumChannels(r io.Reader) error {
	var numChannels int16
	if err := binary.Read(r, binary.LittleEndian, &numChannels); err != nil {
		return err
	}
	w.store["NumChannels"] = strconv.FormatInt(int64(numChannels), 10)
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
