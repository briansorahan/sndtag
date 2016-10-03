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

// expectChunkID reads a chunk ID from an io.Reader and checks it
// against an expected value.
func (w wav) expectChunkID(r io.Reader, expected string) error {
	chunkID, err := w.readChunkID(r)
	if err != nil {
		return err
	}
	if expected != string(chunkID) {
		return fmt.Errorf("expected chunk ID %s, got %s", expected, chunkID)
	}
	return nil
}

// readChunkID reads a chunk ID.
func (w wav) readChunkID(r io.Reader) ([]byte, error) {
	chunkID := make([]byte, 4)
	bytesRead, err := r.Read(chunkID)
	if err != nil {
		return nil, err
	}
	if expected, got := 4, bytesRead; expected != got {
		return nil, fmt.Errorf("expected to read %d bytes, actually read %d", expected, got)
	}
	return chunkID, nil
}

// readFormat reads the format chunk.
// It also stores the formatting information as properties.
func (w wav) readFormat(r io.Reader) error {
	// Read the format identifier, should be "fmt ".
	if err := w.expectChunkID(r, "fmt "); err != nil {
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
	if err := w.readInt16(r, "NumChannels"); err != nil {
		return err
	}

	// Read sample rate.
	if err := w.readInt32(r, "SampleRate"); err != nil {
		return err
	}

	// Read byte rate.
	if err := w.readInt32(r, "ByteRate"); err != nil {
		return err
	}

	// Read block align.
	if err := w.readInt16(r, "BlockAlign"); err != nil {
		return err
	}

	// Read bit rate.
	if err := w.readInt16(r, "BitRate"); err != nil {
		return err
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

// readInt16 reads an int16 from an io.Reader and stores it as a property.
func (w wav) readInt16(r io.Reader, prop string) error {
	var val int16
	if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
		return err
	}
	w.store[prop] = strconv.FormatInt(int64(val), 10)
	return nil
}

// readInt32 reads an int32 from an io.Reader and stores it as a property.
func (w wav) readInt32(r io.Reader, prop string) error {
	var val int32
	if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
		return err
	}
	w.store[prop] = strconv.Itoa(int(val))
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
	if err := getter.expectChunkID(r, "WAVE"); err != nil {
		return nil, err
	}

	// Read the fmt chunk.
	if err := getter.readFormat(r); err != nil {
		return nil, err
	}

	return getter, nil
}
