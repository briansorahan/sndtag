package sndtag

import (
	"io"
)

// newID3 returns a new getter that returns properties from ID3 metadata.
func newID3(r io.Reader) (Getter, error) {
	return nil, nil
}
