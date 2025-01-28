package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"

	"github.com/thlcodes/go-actress/actor"
	logger "github.com/thlcodes/go-actress/log"
)

// models

type User struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	buf  []byte
}

// actor

type addUser struct {
	actor.Message
	Name string
}

type deleteUser struct {
	actor.Message
	Id int
}

type getUsers struct {
	actor.Message
}

type userList struct {
	actor.Message `json:"-"`
	Users         []User `json:"users"`
}

type usersStatus struct {
	actor.Message `json:"-"`
	UserCount     int `json:"user_count"`
}

type UsersActor struct {
	users []User
}

func (u *UsersActor) Handle(ctx actor.Context, msg actor.Message) (reply actor.Message, err error) {
	host, _ := os.Hostname()
	switch msg := msg.(type) {
	case *actor.Start:
		log.Printf("UsersActor(%s@%s) started successfully", ctx.Self(), host)
	case *actor.Stop:
		log.Printf("UsersActor(%s@%s) received STOP message ...", ctx.Self(), host)
	case getUsers:
		return userList{Users: u.users}, nil
	case addUser:
		if slices.ContainsFunc(u.users, func(user User) bool { return user.Name == msg.Name }) {
			return &actor.Error{Error: fmt.Errorf("we already have a %s", msg.Name), Code: 409}, nil
		}
		u.users = append(u.users, User{Id: len(u.users) + 1, Name: msg.Name, buf: make([]byte, 1000)})
		return usersStatus{UserCount: len(u.users)}, nil
	case deleteUser:
		found := false
		for i := len(u.users) - 1; i >= 0; i-- {
			if u.users[i].Id == msg.Id {
				found = true
				u.users = slices.Delete(u.users, i, i+1)
			}
		}
		if !found {
			return &actor.Error{Error: fmt.Errorf("could not find user with id: %d", msg.Id), Code: 404}, nil
		}
		return usersStatus{UserCount: len(u.users)}, nil
	}
	return nil, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer stop()

	log.Printf("starting with PID %d", os.Getpid())
	sys := actor.NewSystem(ctx)
	sys.SetLogger(logger.NewStdLogger().WithLevel(logger.INFO))

	usersActor := sys.Spawn(&UsersActor{users: []User{}})

	http.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("GET /users")
		reply, err := sys.Ask(usersActor, getUsers{})
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "could get add user: %s", err.Error())
			return
		}

		log.Printf("%T", reply)

		switch reply := reply.(type) {
		case userList:
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(reply)
		case *actor.Error:
			w.WriteHeader(reply.Code)
			fmt.Fprint(w, reply.Error.Error())
		}
	})

	http.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("POST /users")
		user := User{}
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			w.WriteHeader(500)
			log.Print(err.Error())
			fmt.Fprintf(w, "could not decode body to user: %s", err.Error())
			return
		}

		reply, err := sys.Ask(usersActor, addUser{Name: user.Name})
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "could not add user: %s", err.Error())
			return
		}

		switch reply := reply.(type) {
		case usersStatus:
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(reply)
		case *actor.Error:
			w.WriteHeader(reply.Code)
			fmt.Fprint(w, reply.Error.Error())
		}
	})

	http.HandleFunc("DELETE /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("DELETE /users/{id}")
		userId, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			w.WriteHeader(400)
			fmt.Fprintf(w, "could not parse user id %s: %s", r.PathValue("id"), err.Error())
			return
		}

		reply, err := sys.Ask(usersActor, deleteUser{Id: userId})
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprintf(w, "could not delete user: %s", err.Error())
			return
		}

		switch reply := reply.(type) {
		case usersStatus:
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(reply)
		case *actor.Error:
			w.WriteHeader(reply.Code)
			fmt.Fprint(w, reply.Error.Error())
		}
	})

	go func() { _ = http.ListenAndServe("localhost:8080", nil) }()

	<-ctx.Done()
	os.Exit(0)
}
