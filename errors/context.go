package errors

type errorWithContext interface {
	Set(string, interface{})
	Get(string) interface{}
	Del(string)
	Context() map[string]interface{}
}

// ErrorSetContext sets a value in the error's context. If the error has not
// been wrapped, it does nothing.
func ErrorSetContext(err error, key string, value interface{}) {
	if e, ok := err.(errorWithContext); ok {
		e.Set(key, value)
	}
}

// ErrorGetContext gets a value from the error's context. If the error has not
// been wrapped, it returns an empty string.
func ErrorGetContext(err error, key string) interface{} {
	if e, ok := err.(errorWithContext); ok {
		return e.Get(key)
	}
	return ""
}

// ErrorDelContext removes a value from the error's context. If the error has
// not been wrapped, it does nothing.
func ErrorDelContext(err error, key string) {
	if e, ok := err.(errorWithContext); ok {
		e.Del(key)
	}
}

// ErrorContext returns the context map for an error if it is a wrappedError.
// If it is not a wrappedError it will return an empty map.
func ErrorContext(err error) map[string]interface{} {
	if e, ok := err.(errorWithContext); ok {
		return e.Context()
	}
	return nil
}
