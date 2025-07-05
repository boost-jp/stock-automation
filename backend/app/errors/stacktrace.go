package errors

import (
	"runtime"
	"strings"
)

const (
	TraceStart int = 2
	TraceDepth int = 32
)

type trace struct {
	FuncName string
	FileName string
	Line     int
}

func newTrace(
	funcName string,
	fileName string,
	line int,
) *trace {
	return &trace{
		FuncName: funcName,
		FileName: fileName,
		Line:     line,
	}
}

func StackTrace() []*trace {
	st := make([]*trace, 0, TraceDepth)

	for depth := TraceStart; depth <= TraceDepth; depth++ {
		pc, file, line, ok := runtime.Caller(depth)
		if !ok {
			break
		}

		fn := runtime.FuncForPC(pc)

		if strings.EqualFold(file, "") {
			break
		}

		st = append(st, newTrace(fn.Name(), file, line))
	}

	return st
}
