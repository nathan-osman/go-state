package state

import (
	"encoding/json"

	"github.com/lampctl/go-sse"
)

// Object provides a simple means of storing, updating, and serializing data.
// Use Object instances as values to nest data.
type Object map[string]any

// Update merges the provided object into this one. This method may use itself
// recursively to add nested Objects.
func (o Object) Update(newObject Object) {
	for newKey, newVal := range newObject {

		// new value is not an object, assign it directly to this object
		newObj, ok := newVal.(Object)
		if !ok {
			o[newKey] = newVal
			continue
		}

		// new value is an object, see if it exists; otherwise assign
		oldVal, ok := o[newKey]
		if !ok {
			o[newKey] = newVal
			continue
		}

		// new key exists in this object, see if it exists as an object;
		// otherwise, assign as per usual
		oldObj, ok := oldVal.(Object)
		if !ok {
			o[newKey] = newVal
			continue
		}

		// They are both objects, recurse!
		oldObj.Update(newObj)
	}
}

// Event creates an *sse.Event from the Object.
func (o Object) Event() (*sse.Event, error) {
	b, err := json.Marshal(o)
	if err != nil {
		return nil, err
	}
	return &sse.Event{
		Data: string(b),
	}, nil
}
