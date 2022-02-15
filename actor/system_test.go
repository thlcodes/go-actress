package actor_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thlcoes/go-actress/actor"
	"github.com/thlcoes/go-actress/log"
)

func newSystem() (sys actor.System) {
	sys = actor.NewSystem()
	sys.SetLogger(log.NewStdLogger().WithLevel(log.TRACE).WithPrefix("testsystem"))
	return
}

func TestSystemDoesNotPanic(t *testing.T) {
	require.NotPanics(t, func() {
		sys := newSystem()
		sys.Stop()
	})
}

func TestSystemSpawn(t *testing.T) {
	sys := newSystem()
	a := &simpleActor{}
	ref := sys.Spawn(a)
	require.NotNil(t, ref)
	sys.Stop()
	time.Sleep(time.Millisecond)
}

type ackMsg struct {
	actor.Message
	i int
}

type ackActor struct {
	ack chan ackMsg
}

var _ actor.Actor = (*ackActor)(nil)

func (a *ackActor) Handle(ctx actor.Context, msg actor.Message) (actor.Message, error) {
	if a.ack != nil {
		switch msg := msg.(type) {
		case ackMsg:
			a.ack <- ackMsg{i: msg.i}
		case *actor.Stop:
			close(a.ack)
		}
		return nil, nil
	} else {
		switch msg := msg.(type) {
		case ackMsg:
			return ackMsg{i: msg.i}, nil
		}
		return nil, nil
	}
}

func TestSystemTell(t *testing.T) {
	sys := newSystem()
	defer sys.Stop()
	a := &ackActor{
		ack: make(chan ackMsg),
	}
	ref := sys.Spawn(a)
	require.NotNil(t, ref)

	n := 5
	for i := 0; i < n; i++ {
		require.NoError(t, sys.Tell(ref, ackMsg{i: i}))
	}
	sys.Kill(ref, true)

	i := 0
loop:
	for {
		select {
		case ack, open := <-a.ack:
			if !open {
				break loop
			}
			require.Equal(t, i, ack.i)
			i++
		case <-time.After(1 * time.Second):
			t.Error("timeout")
			break loop
		}
	}
	require.Equal(t, n, i)
}

func TestSystemAsk(t *testing.T) {
	sys := newSystem()
	defer sys.Stop()
	a := &ackActor{}
	ref := sys.Spawn(a)
	require.NotNil(t, ref)

	n := 5
	for i := 0; i < n; i++ {
		reply, err := sys.Ask(ref, ackMsg{i: i})
		require.NoError(t, err)
		require.IsType(t, ackMsg{}, reply)
		require.Equal(t, i, reply.(ackMsg).i)
	}
}
