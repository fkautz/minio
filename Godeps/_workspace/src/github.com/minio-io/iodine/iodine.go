/*
 * Iodine, (C) 2014 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package iodine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
)

// Error is the iodine error which contains a pointer to the original error
// and stack traces.
type Error struct {
	EmbeddedError error `json:"-"`
	ErrorMessage  string

	Stack []StackEntry
}

// StackEntry contains the entry in the stack trace
type StackEntry struct {
	Host string
	File string
	Line int
	Data map[string]string
}

var globalState = struct {
	sync.RWMutex
	m map[string]string
}{m: make(map[string]string)}

func SetGlobalState(key, value string) {
	globalState.Lock()
	globalState.m[key] = value
	globalState.Unlock()
}

func ClearGlobalState() {
	globalState.Lock()
	for k, _ := range globalState.m {
		delete(globalState.m, k)
	}
	globalState.Unlock()
}

func GetGlobalState() map[string]string {
	result := make(map[string]string)
	globalState.RLock()
	for k, v := range globalState.m {
		result[k] = v
	}
	globalState.RUnlock()
	return result
}

// Wrap an error, turning it into an iodine error.
// Adds an initial stack trace.
func New(err error, data map[string]string) *Error {
	entry := createStackEntry()
	for k, v := range data {
		entry.Data[k] = v
	}
	return &Error{
		EmbeddedError: err,
		ErrorMessage:  err.Error(),
		Stack:         []StackEntry{entry},
	}
}

func createStackEntry() StackEntry {
	host, _ := os.Hostname()
	_, file, line, _ := runtime.Caller(2)
	entry := StackEntry{
		Host: host,
		File: file,
		Line: line,
		Data: GetGlobalState(),
	}
	return entry
}

// Annotate an error with a stack entry and returns itself
func (err *Error) Annotate(info map[string]string) *Error {
	entry := createStackEntry()
	for k, v := range info {
		entry.Data[k] = v
	}
	err.Stack = append(err.Stack, entry)
	return err
}

// EmitJSON writes JSON output for the error
func (err Error) EmitJSON() ([]byte, error) {
	return json.Marshal(err)
}

// EmitHumanReadable returns a human readable error message
func (err Error) EmitHumanReadable() string {
	var errorBuffer bytes.Buffer
	fmt.Fprintln(&errorBuffer, err.Error())
	for i, entry := range err.Stack {
		fmt.Fprintln(&errorBuffer, "-", i, entry.Host+":"+entry.File+":"+strconv.Itoa(entry.Line), entry.Data)
	}
	return string(errorBuffer.Bytes())
}

// Emits the original error message
func (err Error) Error() string {
	return err.EmbeddedError.Error()
}
