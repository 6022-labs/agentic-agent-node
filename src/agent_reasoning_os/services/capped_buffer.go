package services

import "sync"

const truncationMarker = "\n… [output truncated] …\n"

// cappedBuffer keeps the first half and the last half of a stream, so a runaway command can't blow up the transcript.
type cappedBuffer struct {
	mutex     sync.Mutex
	max       int
	head      []byte
	tail      []byte
	truncated bool
}

func newCappedBuffer(max int) *cappedBuffer {
	return &cappedBuffer{max: max}
}

func (b *cappedBuffer) Write(p []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	written := len(p)

	if room := b.max/2 - len(b.head); room > 0 {
		taken := min(room, len(p))
		b.head = append(b.head, p[:taken]...)
		p = p[taken:]
	}

	if len(p) > 0 {
		b.tail = append(b.tail, p...)
		// Only a tail that overflows actually loses bytes; reaching the tail at all does not.
		if overflow := len(b.tail) - b.max/2; overflow > 0 {
			// Shift in place — reslicing would keep the whole backing array alive.
			b.tail = append(b.tail[:0], b.tail[overflow:]...)
			b.truncated = true
		}
	}

	return written, nil
}

func (b *cappedBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	if !b.truncated {
		return string(b.head) + string(b.tail)
	}
	return string(b.head) + truncationMarker + string(b.tail)
}

func (b *cappedBuffer) Truncated() bool {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.truncated
}
