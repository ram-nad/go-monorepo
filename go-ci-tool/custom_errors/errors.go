// Package customerrors contains custom error types
package customerrors

// Empty Struct
type noLogError struct {
	_ int
}

func (e noLogError) Error() string {
	return ""
}

func (e noLogError) Is(err error) bool {
	_, ok := err.(noLogError)
	return ok
}

func NewErrNoLog() error {
	return noLogError{}
}
