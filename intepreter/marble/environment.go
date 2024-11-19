package marble

type environment struct {
	store map[string]object
	outer *environment
}

func NewEnvironment() *environment {
	return &environment{store: make(map[string]object)}
}

func newEnclosedEnvironment(outer *environment) *environment {
	return &environment{store: make(map[string]object), outer: outer}
}

func (e *environment) set(key string, value object) {
	e.store[key] = value
}

func (e *environment) get(key string) (object, bool) {
	value, ok := e.store[key]
	if !ok && e.outer != nil {
		value, ok = e.outer.get(key)
	}
	return value, ok
}
