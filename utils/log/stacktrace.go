package log

import (
	"bytes"
	"github.com/pkg/errors"
)

var (
	colon = []byte(":")
)

type state struct {
	b []byte
}

// Write implement fmt.Formatter interface.
func (s *state) Write(b []byte) (n int, err error) {
	s.b = b
	return len(s.b), nil
}

// Width implement fmt.Formatter interface.
func (s *state) Width() (wid int, ok bool) {
	return 0, false
}

// Precision implement fmt.Formatter interface.
func (s *state) Precision() (prec int, ok bool) {
	return 0, false
}

// Flag implement fmt.Formatter interface.
func (s *state) Flag(c int) bool {
	return true
}

func frameField(f errors.Frame, s *state) string {
	var array bytes.Buffer
	f.Format(s, 's')
	array.Write(s.b)
	array.Write(colon)
	f.Format(s, 'd')
	array.Write(s.b)
	return array.String()
}

// MarshalStack implements pkg/ stack trace marshaling.
//
// zerolog.ErrorStackMarshaler = MarshalStack
func MarshalStack(err error) interface{} {
	type stackTracer interface {
		StackTrace() errors.StackTrace
	}
	var sterr stackTracer
	var ok bool
	for err != nil {
		sterr, ok = err.(stackTracer)
		if ok {
			break
		}

		u, ok := err.(interface {
			Unwrap() error
		})
		if !ok {
			return nil
		}

		err = u.Unwrap()
	}
	if sterr == nil {
		return nil
	}

	st := sterr.StackTrace()
	s := &state{}
	maxTrace := 3
	if maxTrace > len(st) {
		maxTrace = len(st)
	}
	out := make([]string, 0, maxTrace)
	for i := 0; i < maxTrace; i++ {
		frame := st[i]
		out = append(out, frameField(frame, s))
	}
	return out
}
