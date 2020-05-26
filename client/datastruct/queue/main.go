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
	//	"reflect"

	//	"time"
	st "k8s.io/client-go/tools/cache"
)

func main() {

//	TestFIFO_basic()
	TestFIFO_small()
}


func TestFIFO_small() {
	f := st.NewFIFO(testFifoObjectKeyFunc)

	f.Add(mkFifoObj("A", 11))
	f.Add(mkFifoObj("B", 22))
	f.Add(mkFifoObj("C", 33))
	f.Add(mkFifoObj("D", 44))
	f.Add(mkFifoObj("E", 55))

	if e, a := 11, st.Pop(f).(testFifoObject).val; a != e {
		fmt.Print("expected %d, got %d", e, a)
	}

	f.Add(mkFifoObj("foo", 14))

	if e, a := 22, st.Pop(f).(testFifoObject).val; a != e {
		fmt.Print("expected %d, got %d", e, a)
	}

	if e, a := 33, st.Pop(f).(testFifoObject).val; a != e {
		fmt.Print("expected %d, got %d", e, a)
	}

	if e, a := 44, st.Pop(f).(testFifoObject).val; a != e {
		fmt.Print("expected %d, got %d", e, a)
	}
}


func testFifoObjectKeyFunc(obj interface{}) (string, error) {
	return obj.(testFifoObject).name, nil
}

type testFifoObject struct {
	name string
	val  interface{}
}

func mkFifoObj(name string, val interface{}) testFifoObject {
	return testFifoObject{name: name, val: val}
}

func TestFIFO_basic() {
	f := st.NewFIFO(testFifoObjectKeyFunc)
	const amount = 500
	go func() {
		for i := 0; i < amount; i++ {
			f.Add(mkFifoObj(string([]rune{'a', rune(i)}), i+1))
		}
	}()
	go func() {
		for u := uint64(0); u < amount; u++ {
			f.Add(mkFifoObj(string([]rune{'b', rune(u)}), u+1))
		}
	}()

	lastInt := int(0)
	lastUint := uint64(0)
	for i := 0; i < amount*2; i++ {
		switch obj := st.Pop(f).(testFifoObject).val.(type) {
		case int:
			if obj <= lastInt {
				fmt.Print("got %v (int) out of order, last was %v", obj, lastInt)
			}
			lastInt = obj
		case uint64:
			if obj <= lastUint {
				fmt.Print("got %v (uint) out of order, last was %v", obj, lastUint)
			} else {
				lastUint = obj
			}
		default:
			fmt.Print("unexpected type %#v", obj)
		}
	}
}

