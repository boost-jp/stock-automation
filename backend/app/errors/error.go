package errors

import (
	"errors"
	"fmt"
	"strings"
)

type customError struct {
	Err               error
	Stack             []*trace
	ErrCode           string // グローバルスコープでないエラーコード。例えば特定のAPIだけで使うエラーコード。未設定時は空文字
	ErrDisplayMessage string // フロント利用者向けのエラーメッセージ。未設定時は空文字
}

func New(message string, options ...ErrorOption) error {
	ce := &customError{ //nolint:exhaustruct
		Err:   errors.New(message),
		Stack: StackTrace(),
	}
	for _, option := range options {
		option(ce)
	}

	return ce
}

type ErrorOption func(*customError)

func WithCustomErrorCode(code string) ErrorOption {
	return func(e *customError) {
		e.ErrCode = code
	}
}

func WithErrDisplayMessage(message string) ErrorOption {
	return func(e *customError) {
		e.ErrDisplayMessage = message
	}
}

func (m *customError) Error() string {
	return m.Err.Error()
}

func Wrap(err error, message string, options ...ErrorOption) error {
	if err == nil {
		return nil
	}

	wrapped := fmt.Errorf("%w: %s", err, message)

	var traces []*trace

	var cerr *customError
	if errors.As(err, &cerr) {
		traces = cerr.Stack
	} else {
		traces = StackTrace()
	}

	ce := &customError{ //nolint:exhaustruct
		Err:   wrapped,
		Stack: traces,
	}
	for _, option := range options {
		option(ce)
	}

	return ce
}

// WrapWithCustom
// originalのstacktraceを継承し、カスタムエラーでWrapする
// Wrapを二回通しても最下層のエラーのstacktraceが継承される.
func WrapWithCustom(original error, err error, options ...ErrorOption) error {
	if original == nil {
		return nil
	}

	wrapped := fmt.Errorf("%w: %s", err, original) //nolint:errorlint

	var traces []*trace

	var cerr *customError
	if errors.As(original, &cerr) {
		traces = cerr.Stack
	} else {
		traces = StackTrace()
	}

	ce := &customError{ //nolint:exhaustruct
		Err:   wrapped,
		Stack: traces,
	}
	for _, option := range options {
		option(ce)
	}

	return ce
}

var shortStackDepth = 3

func shortStack(err error) []*trace {
	if err == nil {
		return []*trace{}
	}

	var ce *customError

	if errors.As(err, &ce) {
		var stacks []*trace

		var depth int

		for _, t := range ce.Stack {
			stacks = append(stacks, t)

			depth++
			if depth >= shortStackDepth {
				break
			}
		}

		return stacks
	}

	return []*trace{}
}

func PrintStack(traces []*trace) string {
	var stacks []string
	for _, t := range traces {
		stacks = append(stacks, fmt.Sprintf("\t%s\n\t\t%s:%d", t.FuncName, t.FileName, t.Line))
	}

	return strings.ReplaceAll(strings.TrimLeft(strings.Join(stacks, "\n"), "\t"), "\n\t\t", ";")
}

func UnWrap(err error) error {
	if err == nil {
		return nil
	}

	var ce *customError
	if errors.As(err, &ce) {
		return ce.Err
	}

	return err
}

func CustomErrorCode(err error) string {
	if err == nil {
		return ""
	}

	var ce *customError
	if errors.As(err, &ce) {
		return ce.ErrCode
	}

	return ""
}

func PrintShortStack(err error) string {
	if err == nil {
		return ""
	}

	var ce *customError
	if !errors.As(err, &ce) {
		return fmt.Errorf("unwrapped error: %w", err).Error()
	}

	return PrintStack(shortStack(ce))
}

func ErrDisplayMessage(err error) string {
	if err == nil {
		return ""
	}

	var ce *customError
	if errors.As(err, &ce) {
		return ce.ErrDisplayMessage
	}

	return ""
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

func IsInvalidArgument(err error) bool {
	return errors.Is(err, ErrInvalidArgument)
}

func IsAlreadyExists(err error) bool {
	return errors.Is(err, ErrAlreadyExists)
}

func IsUnauthenticated(err error) bool {
	return errors.Is(err, ErrUnauthenticated)
}

func IsPermissionDenied(err error) bool {
	return errors.Is(err, ErrPermissionDenied)
}

func IsPreconditionFailed(err error) bool {
	return errors.Is(err, ErrPreconditionFailed)
}

func IsExternal(err error) bool {
	return errors.Is(err, ErrExternal)
}

func IsUnknown(err error) bool {
	return errors.Is(err, ErrUnknown)
}

func IsWarning(err error) bool {
	return IsNotFound(err) || IsInvalidArgument(err) || IsAlreadyExists(err) || IsUnauthenticated(err) || IsPermissionDenied(err) || IsPreconditionFailed(err)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}

func (m *customError) Is(target error) bool {
	return errors.Is(m.Err, target)
}

func NewNotFound(message string) error {
	return Wrap(ErrNotFound, message)
}

func NewInvalidArgument(message string) error {
	return Wrap(ErrInvalidArgument, message)
}

func NewUnauthenticated(message string) error {
	return Wrap(ErrUnauthenticated, message)
}

func NewPermissionDenied(message string) error {
	return Wrap(ErrPermissionDenied, message)
}

func NewPreconditionFailed(message string) error {
	return Wrap(ErrPreconditionFailed, message)
}

func NewAlreadyExists(message string) error {
	return Wrap(ErrAlreadyExists, message)
}

func NewExternalError(message string) error {
	return Wrap(ErrExternal, message)
}

func NewUnknown(message string) error {
	return Wrap(ErrUnknown, message)
}
