package actor

import "context"

type Context interface {
	context.Context

	referencer

	supervisor

	talker

	System() System
	Inner() context.Context

	WithSender(sender Ref) Context
}

type actorContext struct {
	context.Context

	system System

	self   Ref
	sender Ref
}

var _ Context = (*actorContext)(nil)

func newActorContext(ctx context.Context, system System, self Ref) Context {
	return &actorContext{
		Context: ctx,
		system:  system,
		self:    self,
	}
}

func (c *actorContext) WithSender(sender Ref) Context {
	return &actorContext{
		Context: c.Context,
		system:  c.system,
		self:    c.self,
		sender:  sender,
	}
}

func (c *actorContext) System() System {
	return c.system
}

func (c *actorContext) Self() Ref {
	return c.self
}

func (c *actorContext) Sender() Ref {
	return c.sender
}

func (c *actorContext) Inner() context.Context {
	return c.Context
}

func (c *actorContext) Tell(whom Ref, what Message, opts ...TalkOptions) error {
	return c.system.Tell(whom, what, opts...)
}

func (c *actorContext) Ask(whom Ref, what Message, opts ...TalkOptions) (reply Message, err error) {
	return c.system.Ask(whom, what, opts...)
}

func (c *actorContext) Spawn(actor Actor) Ref {
	return c.system.Spawn(actor)
}
func (c *actorContext) Kill(ref Ref, graceful bool) error {
	return c.system.Kill(ref, graceful)
}
