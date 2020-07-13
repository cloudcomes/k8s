/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"k8s.io/client-go/util/workqueue"

)

func main() {

	TestReinsert()
}

func TestReinsert() {
	q := workqueue.New()
	q.Add("one")
	q.Add("two")
	q.Add("three")
	q.Add("one")

	// Start processing
	i, _ := q.Get()
	if i != "foo" {
		fmt.Print("Expected %v, got %v", "foo", i)
	}

	// Add it back while processing
	q.Add(i)

	// Finish it up
	q.Done(i)

	// It should be back on the queue
	i, _ = q.Get()
	if i != "foo" {
		fmt.Print("Expected %v, got %v", "foo", i)
	}

	// Finish that one up
	q.Done(i)

	if a := q.Len(); a != 0 {
		fmt.Print("Expected queue to be empty. Has %v items", a)
	}
}
