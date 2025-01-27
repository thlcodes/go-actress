package actor_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/thlcodes/go-actress/actor"
	"github.com/thlcodes/go-actress/log"
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
		ack: make(chan ackMsg, 10),
	}
	ref := sys.Spawn(a)
	require.NotNil(t, ref)

	n := 5
	for i := 0; i < n; i++ {
		require.NoError(t, sys.Tell(ref, ackMsg{i: i}))
	}
	_ = sys.Kill(ref, true)

	i := 0
loop:
	for {
		select {
		case ack, ok := <-a.ack:
			if !ok {
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

type countActor struct {
	cnt uint64
}

var _ actor.Actor = (*countActor)(nil)

func (ca *countActor) Handle(ctx actor.Context, msg actor.Message) (reply actor.Message, err error) {
	switch msg.(type) {
	case *actor.Stop:
	// noop
	default:
		ca.cnt++
	}
	return nil, nil
}

func Benchmark_System_Ack(b *testing.B) {
	b.StopTimer()
	sys := newSystem()
	sys.SetLogger(log.NewStdLogger().WithLevel(log.INFO))
	defer sys.Stop()
	act := &ackActor{ack: make(chan ackMsg, b.N)}
	ref := sys.Spawn(act)
	b.StartTimer()
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for r := range act.ack {
			_ = r
		}
		wg.Done()
	}()
	for i := 0; i < b.N; i++ {
		_ = sys.Tell(ref, nil)
	}
	_ = sys.Kill(ref, true)
	wg.Wait()
	b.StopTimer()
}

func Benchmark_System_Counter(b *testing.B) {
	b.StopTimer()
	sys := newSystem()
	sys.SetLogger(log.NewStdLogger().WithLevel(log.INFO))
	defer sys.Stop()
	act := &countActor{}
	ref := sys.Spawn(act)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		_, err := sys.Ask(ref, nil)
		require.NoError(b, err)
	}
	_ = sys.Kill(ref, true)
	b.StopTimer()
	require.Equal(b, uint64(b.N), act.cnt)
}
