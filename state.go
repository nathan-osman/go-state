package state

import (
	"encoding/json"
	"net/http"
	"reflect"
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
// application state. Clients are assigned roles and data is sent to them based
// on their roles.
type State struct {
	mutex   sync.Mutex
	cfg     *Config
	handler *sse.Handler
	data    map[string]reflect.Value
}

func (s *State) dataToEvent(v any) (*sse.Event, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return &sse.Event{
		Data: string(b),
	}, nil
}

func (s *State) connectedFn(r *http.Request) any {
	return s.cfg.RoleFn(r)
}

func (s *State) initFn(v any) []*sse.Event {
	defer s.mutex.Unlock()
	s.mutex.Lock()
	iVal, ok := s.data[v.(string)]
	if !ok {
		return nil
	}
	e, err := s.dataToEvent(iVal.Interface())
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
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func New(cfg *Config) *State {
	s := &State{
		cfg: cfg,
	}
	s.handler = sse.NewHandler(&sse.HandlerConfig{
		ConnectedFn: s.connectedFn,
		InitFn:      s.initFn,
		FilterFn:    s.filterFn,
	})
	return s
}

// Update merges the provided value into the value for the provided roles.
func (s *State) Update(v any, roles []string) {

	// TODO: loop through each of the roles, updating the data in each

	// Send the delta update to the connected clients; FilterFn will ensure
	// that only the clients with that role receive it
	e, err := s.dataToEvent(v)
	if err != nil {
		// TODO: log error
		return
	}
	e.Type = TypeDelta
	e.UserData = roles
	s.handler.Send(e)
}
