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
			t.Fatalf("%s: v.Dest != v.Output", v.Name)
		}
	}
}

func TestObjectEvent(t *testing.T) {
	for _, v := range []struct {
		Name   string
		Object Object
		Data   string
	}{
		{
			Name:   "empty object",
			Object: Object{},
			Data:   "{}",
		},
		{
			Name:   "simple object",
			Object: Object{"1": "2"},
			Data:   "{\"1\":\"2\"}",
		},
		{
			Name:   "nested object",
			Object: Object{"1": Object{"2": "3"}},
			Data:   "{\"1\":{\"2\":\"3\"}}",
		},
	} {
		e, err := v.Object.Event()
		if err != nil {
			t.Fatalf("%s: %s", v.Name, err)
		}
		if e.Data != v.Data {
			t.Fatalf("%s: \"%s\" != \"%s\"", v.Name, e.Data, v.Data)
		}
	}
}
