package utils

/*
#include <stdlib.h>
*/
import "C"
import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
	"unsafe"
)

func FreeCString(ptr unsafe.Pointer) {
	if ptr != nil {
		C.free(ptr)
	}
}

// BuildErrJSON builds a structured JSON string for FFI-safe error reporting.
// Fields include:
//   - code: always -1 for error
//   - message: err.Error()
//   - type: concrete error type
//   - timestamp: RFC3339Nano in UTC
//   - where: first stack frame (file:line)
//   - causes: unwrap chain of error messages
//   - stack: concise frames with func/file/line
func BuildErrJSON(err error) string {
	if err == nil {
		return `{"code":-1,"message":"unknown error"}`
	}

	// Cause chain
	causes := make([]string, 0, 4)
	for e := err; e != nil; e = errors.Unwrap(e) {
		causes = append(causes, e.Error())
	}

	// Stack trace (trim runtime.*)
	const maxDepth = 32
	pcs := make([]uintptr, maxDepth)
	n := runtime.Callers(2, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	stack := make([]map[string]interface{}, 0, n)
	for {
		frame, more := frames.Next()
		if !strings.HasPrefix(frame.Function, "runtime.") {
			stack = append(stack, map[string]interface{}{
				"func": frame.Function,
				"file": frame.File,
				"line": frame.Line,
			})
		}
		if !more {
			break
		}
	}
	where := ""
	if len(stack) > 0 {
		where = fmt.Sprintf("%s:%d", stack[0]["file"], stack[0]["line"])
	}

	data := map[string]interface{}{
		"code":      -1,
		"message":   err.Error(),
		"type":      fmt.Sprintf("%T", err),
		"timestamp": time.Now().UTC().Format(time.RFC3339Nano),
		"where":     where,
		"causes":    causes,
		"stack":     stack,
	}
	b, mErr := json.Marshal(data)
	if mErr != nil {
		return `{"code":-1,"message":"marshal error"}`
	}
	return string(b)
}
