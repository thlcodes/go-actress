package actor

import (
	"errors"
	"fmt"
)

// system
var (
	ErrUnsupportedRef = func(ref Ref) error { return fmt.Errorf("system cannot handle ref %s for now", ref) }
)

// talk errors
var (
	ErrUnsupportedRefForTalking = func(ref Ref) error { return fmt.Errorf("talking to ref %s is currently not supported", ref) }
	ErrActorNotFound            = func(ref Ref) error { return fmt.Errorf("could not find local actor %s", ref) }
	ErrChannelRefChannelClosed  = func(ref Ref) error { return fmt.Errorf("somehow the channel of the channel ref %s was closed", ref) }
	ErrTalkTimeout              = errors.New("talk timeout")
)

// actor errors
var (
	ErrActorNotImplemented = errors.New("actor not implemented yet")
)

// mailbox errors
var (
	ErrMailboxFull = func(ref Ref) error { return fmt.Errorf("mailbox of actor %s is full", ref) }
)
