package state

import (
	"reflect"
	"testing"
)

func TestObjectUpdate(t *testing.T) {
	for _, v := range []struct {
		Name   string
		Dest   Object
		Src    Object
		Output Object
	}{
		{
			Name:   "basic assignment",
			Dest:   Object{},
			Src:    Object{"1": "2"},
			Output: Object{"1": "2"},
		},
		{
			Name:   "assign new object",
			Dest:   Object{},
			Src:    Object{"1": Object{"2": "3"}},
			Output: Object{"1": Object{"2": "3"}},
		},
		{
			Name:   "overwrite scalar with object",
			Dest:   Object{"1": "2"},
			Src:    Object{"1": Object{"2": "3"}},
			Output: Object{"1": Object{"2": "3"}},
		},
		{
			Name:   "merge nested object",
			Dest:   Object{"1": Object{"a": 1, "b": 2}},
			Src:    Object{"1": Object{"a": 2}},
			Output: Object{"1": Object{"a": 2, "b": 2}},
		},
	} {
		v.Dest.Update(v.Src)
		if !reflect.DeepEqual(v.Dest, v.Output) {
			t.Fatalf("%s: %s", v.Name, "v.Dest != v.Output")
		}
	}
}
