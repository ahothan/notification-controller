//go:build gofuzz
// +build gofuzz

/*
Copyright 2022 The Flux authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package notifier

import (
	"io"
	"net/http"
	"net/http/httptest"

	fuzz "github.com/AdaLogics/go-fuzz-headers"
	"github.com/fluxcd/pkg/runtime/events"
)

// FuzzWebex implements a fuzzer that targets Webex.Post().
func FuzzWebex(data []byte) int {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		r.Body.Close()
	}))
	defer ts.Close()

	webex, err := NewWebex(ts.URL, "", nil, "", "")
	if err != nil {
		return 0
	}

	f := fuzz.NewConsumer(data)
	event := events.Event{}

	if err := f.GenerateStruct(&event); err != nil {
		return 0
	}

	_ = webex.Post(event)

	return 1
}
