package state

import (
	"reflect"
	"testing"
)

func TestStateUpdate(t *testing.T) {
	for _, v := range []struct {
		Name     string
		UpdateFn func(s *State)
		Data     map[string]Object
	}{
		{
			Name: "populate with no role",
			UpdateFn: func(s *State) {
				s.Update(Object{}, []string{"test"})
				s.Update(Object{"1": "2"}, nil)
			},
			Data: map[string]Object{
				"test": {"1": "2"},
			},
		},
		{
			Name: "populate with one role",
			UpdateFn: func(s *State) {
				s.Update(Object{}, []string{"test1"})
				s.Update(Object{"1": "2"}, []string{"test2"})
			},
			Data: map[string]Object{
				"test1": {},
				"test2": {"1": "2"},
			},
		},
	} {
		s := New(nil)
		v.UpdateFn(s)
		if !reflect.DeepEqual(s.data, v.Data) {
			t.Fatalf("%s: %#v != %#v", v.Name, s.data, v.Data)
		}
	}
}
