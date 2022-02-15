package actor

type TalkOptions func(*Envelope)

type talker interface {
	Tell(Ref, Message, ...TalkOptions) error
	Ask(Ref, Message, ...TalkOptions) (Message, error)
}

type referencer interface {
	Self() Ref
	Sender() Ref
}

type supervisor interface {
	Spawn(Actor) Ref
	Kill(Ref, bool) error
}
