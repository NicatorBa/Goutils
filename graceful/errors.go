package graceful

import (
	"errors"
	"fmt"
)

var ErrContextClosed = errors.New("context closed")

type PanicError struct {
	Value any
	Stack []byte
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", p.Value)
}

func (p *PanicError) Unwrap() error {
	err, _ := p.Value.(error)
	return err
}
