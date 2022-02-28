package actor

type TalkOption func(*Envelope)
type SpawnOption func(*actor)

type talker interface {
	Tell(Ref, Message, ...TalkOption) error
	Ask(Ref, Message, ...TalkOption) (Message, error)
}

type referencer interface {
	Self() Ref
	Sender() Ref
}

type supervisor interface {
	Spawn(Actor, ...SpawnOption) Ref
	Kill(Ref, bool) error
}
