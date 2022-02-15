package actor_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/thlcoes/go-actress/actor"
)

type simpleMessage struct {
	actor.Message
	i int
}

type simpleReply struct {
	actor.Message
	i int
}

type simpleActor struct {
}

func (sa *simpleActor) Handle(_ctx actor.Context, msg actor.Message) (actor.Message, error) {
	switch msg := msg.(type) {
	case *simpleMessage:
		return &simpleReply{i: msg.i}, nil
	}
	return nil, nil
}

func TestActorSimple(t *testing.T) {
	i := 123
	a := &simpleActor{}
	got, err := a.Handle(nil, &simpleMessage{i: i})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, i, got.(*simpleReply).i)
}
