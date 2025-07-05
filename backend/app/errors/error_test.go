package errors

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var errPackagePath = "github.com/boost-jp/stock-automation/app/errors"

var (
	customWrappedOnce  = errors.New("wrapped once")
	customWrappedTwice = errors.New("wrapped twice")
)

// Level1Function
// エラー発生元.
func Level1Function() error {
	return errors.New("original")
}

// WrappedOnceFunction
// ハンドリングしたエラーをWrapしたレイヤーの関数.
func WrappedOnceFunction() error {
	return WrapWithCustom(Level1Function(), customWrappedOnce)
}

// WrappedTwiceFunction
// ハンドリングしたカスタムエラーを更にWrapしたレイヤーの関数.
func WrappedTwiceFunction() error {
	return WrapWithCustom(WrappedOnceFunction(), customWrappedTwice)
}

func TestWrapWithCustom(t *testing.T) {
	type args struct {
		original error
		err      error
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "nil",
			args: args{
				original: nil,
				err:      nil,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
		{
			name: "depth 3, wrapped with custom",
			args: args{
				original: WrappedOnceFunction(),
				err:      customWrappedOnce,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				wantStacks := []trace{
					{
						FuncName: errPackagePath + ".WrappedOnceFunction",
						Line:     27,
					},
					{
						FuncName: errPackagePath + ".TestWrapWithCustom",
						Line:     59,
					},
				}
				var cerr *customError
				if errors.As(err, &cerr) {
					c := err.Error()
					assert.Equal(t, c, "wrapped once: wrapped once: original")
					actualStacks := shortStack(cerr)
					assert.Len(t, actualStacks, 3)
					assert.Equal(t, wantStacks[0].FuncName, actualStacks[0].FuncName)
					assert.Equal(t, wantStacks[0].Line, actualStacks[0].Line)
					assert.Equal(t, wantStacks[1].FuncName, actualStacks[1].FuncName)
					assert.Equal(t, wantStacks[1].Line, actualStacks[1].Line)
					// testing.tRunnerからのtraceはパッケージの更新があると差分出るのでこれ以降の深さは無視
					return false
				}
				return true
			},
		},
		{
			name: "depth 3, wrapped with custom twice",
			args: args{
				original: WrappedTwiceFunction(),
				err:      customWrappedTwice,
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				wantStacks := []trace{
					{
						FuncName: errPackagePath + ".WrappedOnceFunction",
						Line:     27,
					},
					{
						FuncName: errPackagePath + ".WrappedTwiceFunction",
						Line:     33,
					},
				}
				var cerr *customError
				if errors.As(err, &cerr) {
					c := err.Error()
					assert.Equal(t, c, "wrapped twice: wrapped twice: wrapped once: original")
					actualStacks := shortStack(cerr)
					assert.Len(t, actualStacks, 3)
					assert.Equal(t, wantStacks[0].FuncName, actualStacks[0].FuncName)
					assert.Equal(t, wantStacks[0].Line, actualStacks[0].Line)
					assert.Equal(t, wantStacks[1].FuncName, actualStacks[1].FuncName)
					assert.Equal(t, wantStacks[1].Line, actualStacks[1].Line)
					// testing.tRunnerからのtraceはパッケージの更新があると差分出るのでこれ以降の深さは無視
				}
				return false
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WrapWithCustom(tt.args.original, tt.args.err)

			if !tt.wantErr(t, err, fmt.Sprintf("WrapWithCustom(%v, %v)", tt.args.original, tt.args.err)) {
				return
			}
		})
	}
}
