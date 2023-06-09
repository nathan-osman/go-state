package state

import (
	"net/http"
	"sync"

	"github.com/lampctl/go-sse"
)

const (
	TypeSync  = "sync"
	TypeDelta = "delta"
)

// Config provides the configuration for initializing a State instance.
type Config struct {

	// RoleFn is invoked to determine the role that should be assigned to the
	// connecting client, based on the request.
	RoleFn func(r *http.Request) string
}

// State provides a thread-safe way to manage, update, and synchronize
// application state. Clients are assigned roles and objects are sent to them
// based on their roles.
type State struct {
	mutex   sync.Mutex
	cfg     *Config
	handler *sse.Handler
	data    map[string]Object
}

func (s *State) connectedFn(r *http.Request) any {
	return s.cfg.RoleFn(r)
}

func (s *State) initFn(v any) []*sse.Event {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	o, ok := s.data[v.(string)]
	if !ok {
		return nil
	}
	e, err := o.Event()
	if err != nil {
		// TODO: log error
		return nil
	}
	e.Type = TypeSync
	return []*sse.Event{e}
}

func (s *State) filterFn(v any, e *sse.Event) bool {
	var (
		role  = v.(string)
		roles = e.UserData.([]string)
	)
	if roles == nil {
		return true
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func (s *State) getOrCreateObject(r string) Object {
	o, ok := s.data[r]
	if !ok {
		o = Object{}
		s.data[r] = o
	}
	return o
}

func (s *State) sendDeltaUpdate(o Object, roles []string) {
	e, err := o.Event()
	if err != nil {
		// TODO: log error
		return
	}
	e.Type = TypeDelta
	e.UserData = roles
	s.handler.Send(e)
}

func (s *State) updateAndSend(o, newObj Object, r string) {
	o.Update(newObj)
	s.sendDeltaUpdate(newObj, []string{r})
}

func New(cfg *Config) *State {
	s := &State{
		cfg:  cfg,
		data: make(map[string]Object),
	}
	s.handler = sse.NewHandler(&sse.HandlerConfig{
		NumEventsToKeep:   sse.DefaultHandlerConfig.NumEventsToKeep,
		ChannelBufferSize: sse.DefaultHandlerConfig.ChannelBufferSize,
		ConnectedFn:       s.connectedFn,
		InitFn:            s.initFn,
		FilterFn:          s.filterFn,
	})
	return s
}

// Update merges the provided object into the object for the provided roles. If
// roles is set to nil, all roles will be assumed.
func (s *State) Update(newObj Object, roles []string) {
	defer s.mutex.Unlock()
	s.mutex.Lock()

	// Update the object for each of the specified roles
	if roles == nil {
		for _, o := range s.data {
			o.Update(newObj)
		}
	} else {
		for _, r := range roles {
			s.getOrCreateObject(r).Update(newObj)
		}
	}

	// Send the delta update to the connected clients; FilterFn will ensure
	// that only the clients with that role receive it
	s.sendDeltaUpdate(newObj, roles)
}

// UpdateFunc provides a way to atomically update values in the provided roles.
// The provided callback is invoked with the value of each role and is expected
// to return an object of delta updates.
func (s *State) UpdateFunc(fn func(o Object, r string) Object, roles []string) {
	defer s.mutex.Unlock()
	s.mutex.Lock()

	// Because each role's object could get different updates, send a delta
	// update for each individual change
	if roles == nil {
		for r, o := range s.data {
			s.updateAndSend(o, fn(o, r), r)
		}
	} else {
		for _, r := range roles {
			o := s.getOrCreateObject(r)
			s.updateAndSend(o, fn(o, r), r)
		}
	}
}

func (s *State) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}

// Close shuts down the state and waits for all connections to close.
func (s *State) Close() {
	s.handler.Close()
}
