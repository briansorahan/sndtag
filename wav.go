package sndtag

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

// wav parses RIFF tags from wav files.
// See http://soundfile.sapp.org/doc/WaveFormat/ for more info.
type wav struct {
	length   int32
	metadata map[string]string
}

// newWav creates a new map that contains properties for WAV files.
// Note that the "RIFF" chunk identifier has already been read
// by the time this function is called.
func newWav(r io.Reader) (map[string]string, error) {
	w := wav{
		metadata: map[string]string{},
	}

	// Get the length.
	if err := binary.Read(r, binary.LittleEndian, &w.length); err != nil {
		return nil, err
	}

	// Sniff the format.
	if err := expectFourCC(r, "WAVE"); err != nil {
		return nil, err
	}

	// Read subchunks of the RIFF chunk.
	if err := w.readSubchunks(r); err != nil {
		return nil, err
	}

	return w.metadata, nil
}

// readSubchunks reads the subchunks of the RIFF chunk.
func (w wav) readSubchunks(r io.Reader) error {
	id, length, data, err := readChunk(r)
	if err != nil {
		return err
	}

	switch id {
	case "fmt ":
		// Read the wav format chunk data.
		return w.readFormat(data)
	case "data":
		// Discard the audio data.
		_, err := io.CopyN(ioutil.Discard, r, int64(length))
		return err
	case "LIST":
		// Read a LIST chunk (can contain subchunks).
		return w.readList(data)
	case "INFO":
		// Read an INFO chunk (can contain exif tags).
		// Not sure if the INFO always appears in a LIST, or if it
		// can sometimes appear on its own (briansorahan).
		return w.readInfo(data)
	default:
		return fmt.Errorf("unrecognized chunk ID: %s", id)
	}
}

// readFormat reads the fmt chunk data.
// It also stores the formatting information as properties.
func (w wav) readFormat(r io.Reader) error {
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
	w.metadata["AudioFormat"] = strconv.FormatInt(int64(audioFormat), 10)
	return nil
}

// readInt16 reads an int16 from an io.Reader and stores it as a property.
func (w wav) readInt16(r io.Reader, prop string) error {
	var val int16
	if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
		return err
	}
	w.metadata[prop] = strconv.FormatInt(int64(val), 10)
	return nil
}

// readInt32 reads an int32 from an io.Reader and stores it as a property.
func (w wav) readInt32(r io.Reader, prop string) error {
	var val int32
	if err := binary.Read(r, binary.LittleEndian, &val); err != nil {
		return err
	}
	w.metadata[prop] = strconv.Itoa(int(val))
	return nil
}

// readList reads a LIST chunk, which can contain subchunks.
func (w wav) readList(r io.Reader) error {
	// TODO: Do not force the format to INFO.
	return expectFourCC(r, "INFO")
}

// readInfo reads an INFO chunk.
func (w wav) readInfo(r io.Reader) error {
	return nil
}

// readChunk reads a chunk from an io.Reader and returns the
// chunk identifier, the chunk length, the chunk data, and an error.
func readChunk(r io.Reader) (id string, length int32, data io.Reader, err error) {
	idb, err := readFourCC(r)
	if err != nil {
		return
	}
	id = string(idb)

	if err = binary.Read(r, binary.LittleEndian, &length); err != nil {
		return
	}
	data = io.LimitReader(r, int64(length))
	return
}

// expectFourCC reads a chunk ID from an io.Reader and checks it
// against an expected value.
func expectFourCC(r io.Reader, expected string) error {
	chunkID, err := readFourCC(r)
	if err != nil {
		return err
	}
	if expected != string(chunkID) {
		return fmt.Errorf("expected chunk ID %s, got %s", expected, chunkID)
	}
	return nil
}

// readFourCC reads a chunk ID.
func readFourCC(r io.Reader) ([]byte, error) {
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
