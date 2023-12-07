package keepalive

import (
	"fmt"
	"sync"
	"time"
)

// KeepAlive is a keep alive tracker for detecting defunct connections.
type KeepAlive struct {
	setting, duration time.Duration
	expiration        *time.Timer

	quit   chan struct{}
	closed bool
	mu     sync.Mutex
}

// NewKeepAlive creates a KeepAlive tracker for the given CONNECT keepalive setting of expiration
// in seconds. As per the spec, the connection is considered expired after 1.5x that setting. For
// example, keepalive=60 means that the connection expires after 90s after the last control message
// and should be close by the server. The client can send PING as a nop.
func NewKeepAlive(keepaliveSeconds uint16) *KeepAlive {
	if keepaliveSeconds == 0 {
		return &KeepAlive{quit: make(chan struct{}), closed: true} // nop: never expires
	}

	ret := &KeepAlive{
		setting:  time.Duration(keepaliveSeconds) * time.Second,
		duration: (time.Duration(keepaliveSeconds) * time.Second * 15) / 10,
		quit:     make(chan struct{}),
	}
	ret.expiration = time.AfterFunc(ret.duration, ret.expire)

	return ret
}

// Setting returns the norminal setting, such as 30s.
func (k *KeepAlive) Setting() time.Duration {
	return k.setting
}

// Touch indicates that a packet has been received that resets the expiration.
func (k *KeepAlive) Touch() {
	k.mu.Lock()
	defer k.mu.Unlock()

	if k.closed {
		return
	}

	if k.expiration.Stop() {
		k.expiration.Reset(k.duration)
	} else {
		// create a new timer if existing one has expired, so k.expire does not run twice
		k.expiration = time.AfterFunc(k.duration, k.expire)
	}
}

// IsExpired returns true iff keepalive is expired.
func (k *KeepAlive) IsExpired() bool {
	select {
	case <-k.quit:
		return true
	default:
		return false
	}
}

// Expired returns a chan, which is closed iff keepalive expires.
func (k *KeepAlive) Expired() <-chan struct{} {
	return k.quit
}

func (k *KeepAlive) expire() {
	k.mu.Lock()
	defer k.mu.Unlock()

	if !k.closed {
		close(k.quit)
		k.closed = true
	}
}

func (k *KeepAlive) String() string {
	k.mu.Lock()
	defer k.mu.Unlock()

	return fmt.Sprintf("keepalive[closed=%v, duration=%v]", k.closed, k.duration)
}
