package formattestjson

import (
	"time"
)

/*
Derived from: github.com/robherley/go-test-action (MIT License)

Code Reference: https://cs.opensource.google/go/go/+/master:/src/cmd/test2json/main.go
*/

type TestEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Elapsed float64   `json:"Elapsed"`
	Output  string    `json:"Output"`
}

type TestEventAction string

//	start  - the test binary is about to be executed
//	run    - the test has started running
//	pause  - the test has been paused
//	cont   - the test has continued running
//	pass   - the test passed
//	bench  - the benchmark printed log output but did not fail
//	fail   - the test or benchmark failed
//	output - the test printed output
//	skip   - the test was skipped or the package contained no tests

const (
	TestEventActionStart  TestEventAction = "start"
	TestEventActionRun    TestEventAction = "run"
	TestEventActionPause  TestEventAction = "pause"
	TestEventActionCont   TestEventAction = "cont"
	TestEventActionPass   TestEventAction = "pass"
	TestEventActionBench  TestEventAction = "bench"
	TestEventActionFail   TestEventAction = "fail"
	TestEventActionOutput TestEventAction = "output"
	TestEventActionSkip   TestEventAction = "skip"
)
