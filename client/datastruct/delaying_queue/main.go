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
	"time"

	"k8s.io/apimachinery/pkg/util/clock"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/workqueue"
)

func main() {

	TestSimpleQueue()
}


func TestSimpleQueue() {
	fakeClock := clock.NewFakeClock(time.Now())
	q := workqueue.NewDelayingQueueWithCustomClock(fakeClock, "")

	first := "foo"

	q.AddAfter(first, 50*time.Millisecond)


	if q.Len() != 0 {
		fmt.Printf("should not have added")
	}

    //这时queue里还是空，调用Get会阻塞
	//item1, _ := q.Get()


	fakeClock.Step(60 * time.Millisecond)

	if err := waitForAdded(q, 1); err != nil {
		fmt.Printf("should have added")
	}
	item, _ := q.Get()
	fmt.Printf("%v",item)

	q.Done(item)

	// step past the next heartbeat
	fakeClock.Step(10 * time.Second)

	err := wait.Poll(1*time.Millisecond, 30*time.Millisecond, func() (done bool, err error) {
		if q.Len() > 0 {
			return false, fmt.Errorf("added to queue")
		}

		return false, nil
	})
	if err != wait.ErrWaitTimeout {
		fmt.Printf("expected timeout, got: %v", err)
	}

	if q.Len() != 0 {
		fmt.Printf("should not have added")
	}
}


func waitForAdded(q workqueue.DelayingInterface, depth int) error {
	return wait.Poll(1*time.Millisecond, 10*time.Second, func() (done bool, err error) {
		if q.Len() == depth {
			return true, nil
		}

		return false, nil
	})
}


