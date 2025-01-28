package main

import (
	"context"
	"log"
	"time"

	"github.com/thlcodes/go-actress/actor"
)

type countingActor struct {
	count   int
	control actor.Ref
}

type counterAdd struct {
	actor.Message
	amount int
}

type counterHi struct {
	actor.Message
}

type counterSub struct {
	actor.Message
	amount int
}

type counterState struct {
	actor.Message
	count int
}

func (a *countingActor) Handle(ctx actor.Context, msg actor.Message) (reply actor.Message, err error) {
	log.Printf("msg %T %s", msg, ctx.Sender())
	if msg == nil {
		print("nil")
	}
	check := false
	switch msg := msg.(type) {
	case counterHi:
		log.Printf("hi from %s", ctx.Self())
		_ = ctx.Tell(a.control, controlTap{})
	case counterAdd:
		a.count += msg.amount
		check = true
		_ = ctx.Tell(a.control, controlTap{})
	case counterSub:
		a.count -= msg.amount
		check = true
		_ = ctx.Tell(a.control, controlTap{})
	}

	if check && a.count < 0 {
		if err = ctx.Tell(a.control, controlQuit{}); err != nil {
			return
		}
	}

	return counterState{count: a.count}, nil
}

type controlActor struct {
	close chan struct{}
}

type controlQuit struct {
	actor.Message
}

type controlTap struct {
	actor.Message
}

func (ca *controlActor) Handle(ctx actor.Context, msg actor.Message) (reply actor.Message, err error) {
	switch msg.(type) {
	case controlQuit:
		log.Printf("%s triggered control quit", ctx.Sender())
		ca.close <- struct{}{}
	case controlTap:
		log.Println("control tapped ...")
	}
	return
}

func main() {
	sys := actor.NewSystem(context.TODO())
	defer sys.Stop()

	control := &controlActor{close: make(chan struct{})}
	controlRef := sys.Spawn(control)
	counter := &countingActor{control: controlRef}
	counterRef := sys.Spawn(counter)

	// non-waiting tell
	if err := sys.Tell(counterRef, counterHi{}); err != nil {
		panic(err)
	}

	time.Sleep(1 * time.Millisecond)
	// ask  waiting for a response
	if state, err := sys.Ask(counterRef, counterAdd{amount: 150}); err != nil {
		panic(err)
	} else {
		log.Printf("count %d should be 150 now", state.(counterState).count)
	}

	time.Sleep(1 * time.Millisecond)
	// ask  waiting for a response
	if state, err := sys.Ask(counterRef, counterSub{amount: 70}); err != nil {
		panic(err)
	} else {
		log.Printf("count %d should be 80 now", state.(counterState).count)
	}

	time.Sleep(1 * time.Millisecond)
	// remove enough to trigger control quit
	if state, err := sys.Ask(counterRef, counterSub{amount: 81}); err != nil {
		panic(err)
	} else {
		log.Printf("count %d should be -1 now, this control should quit", state.(counterState).count)
	}

	<-control.close

	log.Println("quit")
}
