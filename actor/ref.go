package actor

import (
	"fmt"
	"time"
)

// Ref to an actor, might be local, remote or cluster
type Ref interface {
	String() string
}

/* local ref */

var _ Ref = (*localRef)(nil)

type localRef struct {
	id uint64
}

func newLocalRef(id uint64) localRef {
	return localRef{
		id: id,
	}
}

func (lr *localRef) String() string {
	return fmt.Sprintf("local#%d", lr.id)
}

/* channel ref */

type channelRef struct {
	Ref
	id int64
	ch chan *Envelope
}

func newChannelRef(ch chan *Envelope) channelRef {
	return channelRef{
		id: time.Now().UnixNano(),
		ch: ch,
	}
}

func (cr *channelRef) String() string {
	return fmt.Sprintf("channel#%d", cr.id)
}
