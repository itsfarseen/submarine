package errorspan

import (
	"fmt"
	"slices"
	"strings"
)

type ErrorSpan struct {
	Path    []string
	Message string
}

func NewErrorSpan(message string) *ErrorSpan {
	return &ErrorSpan{Message: message}
}

func (err *ErrorSpan) WithPath(path string) *ErrorSpan {
	if strings.ContainsAny(path, " /") {
		path = fmt.Sprintf("[%s]", path)
	}
	err.Path = append(err.Path, path)
	return err
}

func (err *ErrorSpan) WithPathf(format string, args ...any) *ErrorSpan {
	return err.WithPath(fmt.Sprintf(format, args...))
}

func (err *ErrorSpan) WithPathInt(n int) *ErrorSpan {
	return err.WithPathf("%d", n)
}

func (err *ErrorSpan) Error() string {
	path := slices.Clone(err.Path)
	slices.Reverse(path)
	pathStr := strings.Join(path, "/")
	msg := err.Message
	return fmt.Sprintf("%s: %s", pathStr, msg)
}
