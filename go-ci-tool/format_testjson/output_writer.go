// Package formattestjson provides a pretty looking output for Go test JSON
package formattestjson

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/ram-nad/go-monorepo/go-ci-tool/v2/color"
)

type PackageTestResult struct {
	PassCount uint
	FailCount uint
	SkipCount uint
}

type TestOutState struct {
	outBuffer     []byte
	PackageOut    map[string][]byte
	PackageResult map[string]PackageTestResult
}

const (
	DefaultTestOutBufferSize    = 4 * 1024 // 4KB
	DefaultPackageOutBufferSize = 1 * 1024 // 1KB
)

func NewBufferSize(len int) int {
	// Size is less than 1 MB
	if len < 1024*1024 {
		//nolint:mnd // double the buffer size
		newLen := 2 * len
		return newLen
	} else {
		// Increase buffer size by 1 MB
		return len + 1024*1024
	}
}

func ParseTestEvent(out []byte) (TestEvent, error) {
	event := TestEvent{}

	err := json.NewDecoder(bytes.NewReader(out)).Decode(&event)
	return event, err
}

func NewTestOutState() *TestOutState {
	return &TestOutState{
		outBuffer:     make([]byte, 0, DefaultTestOutBufferSize),
		PackageOut:    make(map[string][]byte),
		PackageResult: make(map[string]PackageTestResult),
	}
}

/*
Write used to parse the output of `go test -json`

Implement the io.Writer interface
so that we can stream test command output.

In future, we can implement more sophisticated logic here with locks to
make sure that calls to Write is not blocking a lot
*/
func (o *TestOutState) Write(p []byte) (int, error) {
	writeLen := 0
	pLen := len(p)

	for writeLen < pLen {
		canWrite := cap(o.outBuffer) - len(o.outBuffer)

		toWrite := min(writeLen+canWrite, pLen)
		o.outBuffer = append(o.outBuffer, p[writeLen:toWrite]...)
		writeLen = toWrite

		// If we haven't still completed write, try to flush early
		if writeLen < pLen {
			err := o.FlushBuffer()
			if err != nil {
				return writeLen, err
			}

			// This is only possible if single line is more than buffer size,
			// therefore we increase buffer size
			if cap(o.outBuffer) == len(o.outBuffer) {
				newSlice := make(
					[]byte,
					len(o.outBuffer),
					NewBufferSize(len(o.outBuffer)),
				)
				copy(newSlice, o.outBuffer)
				o.outBuffer = newSlice
			}
		}
	}

	// Try to flush the buffer
	err := o.FlushBuffer()

	return writeLen, err
}

func (o *TestOutState) FlushBuffer() error {
	n := len(o.outBuffer)

	i := 0
	j := 0

	for j < n {
		for j < n {
			if rune(o.outBuffer[j]) == rune('\n') {
				break
			}
			j++
		}

		if j < n {
			if i != j {
				event, err := ParseTestEvent(o.outBuffer[i:j])
				if err != nil {
					return err
				}
				o.HandleEvent(&event)
			}
			j += 1
			i = j
		}
	}

	// Make space in buffer
	copy(o.outBuffer, o.outBuffer[i:])
	o.outBuffer = o.outBuffer[:n-i]

	return nil
}

func (o *TestOutState) HandleEvent(event *TestEvent) {
	switch event.Action {
	case string(TestEventActionStart):
		o.PackageOut[event.Package] = make([]byte, 0, DefaultPackageOutBufferSize)
		o.PackageResult[event.Package] = PackageTestResult{}
	case string(TestEventActionPass):
		if event.Test != "" {
			result := o.PackageResult[event.Package]
			result.PassCount++
			o.PackageResult[event.Package] = result
		}
	case string(TestEventActionFail):
		if event.Test != "" {
			result := o.PackageResult[event.Package]
			result.FailCount++
			o.PackageResult[event.Package] = result
		}
	case string(TestEventActionSkip):
		if event.Test != "" {
			result := o.PackageResult[event.Package]
			result.SkipCount++
			o.PackageResult[event.Package] = result
		}
	case string(TestEventActionOutput):
		o.PackageOut[event.Package] = AppendOutput(
			o.PackageOut[event.Package],
			event.Output,
		)
	default:
		// We don't care about other events
	}
}

func AppendOutput(out []byte, testOutput string) []byte {
	switch {
	case strings.HasPrefix(testOutput, "=== RUN"):
		return append(out, color.InfoColor.Sprint(testOutput)...)
	case strings.HasPrefix(testOutput, "--- PASS"):
		return append(out, color.SuccessColor.Sprint(testOutput)...)
	case strings.HasPrefix(testOutput, "--- FAIL"):
		return append(out, color.ErrorColor.Sprint(testOutput)...)
	case strings.HasPrefix(testOutput, "--- SKIP"):
		return append(out, color.WarningColor.Sprint(testOutput)...)
	case strings.HasPrefix(testOutput, "coverage: "):
		return append(out, color.HighLightColor.Sprint(testOutput)...)
	default:
		return append(out, color.MutedColor.Sprint(testOutput)...)
	}
}
