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
	if e.UserData == nil {
		return true
	}
	var (
		role  = v.(string)
		roles = e.UserData.([]string)
	)
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

func New(cfg *Config) *State {
	s := &State{
		cfg:  cfg,
		data: make(map[string]Object),
	}
	s.handler = sse.NewHandler(&sse.HandlerConfig{
		ConnectedFn: s.connectedFn,
		InitFn:      s.initFn,
		FilterFn:    s.filterFn,
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
