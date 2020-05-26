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
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	st "k8s.io/client-go/tools/cache"
)

func main() {

	TestDeltaFIFO_replaceWithDeleteDeltaIn()

}



func TestDeltaFIFO_basic() {

	a1Obj := mkFifoObj("A", 11)
	a2Obj := mkFifoObj("A", 22)
	a3Obj := mkFifoObj("A", 33)
	b4Obj := mkFifoObj("B", 44)


	f := st.NewDeltaFIFO(testFifoObjectKeyFunc, nil)

	f.Add(a1Obj)
	f.Update(a2Obj)
	f.Delete(a3Obj)

	f.Add(b4Obj)


	actualDeltasA := st.Pop(f)
	fmt.Print("actualDeltas %#v", actualDeltasA)

	actualDeltasB := st.Pop(f)
	fmt.Print("actualDeltas %#v", actualDeltasB)

}


func TestDeltaFIFO_replaceWithDeleteDeltaIn() {
	a1Obj := mkFifoObj("A", 11)
	a2Obj := mkFifoObj("A", 22)

	a3Obj := mkFifoObj("A", 33)
	b4Obj := mkFifoObj("B", 44)
	c5Obj := mkFifoObj("C", 55)

	f := st.NewDeltaFIFO(testFifoObjectKeyFunc, nil)

	f.Add(a1Obj)
	f.Update(a2Obj)
	f.Delete(a3Obj)
	f.Replace([]interface{}{a1Obj,b4Obj,a3Obj,c5Obj}, "")


	f.Add(b4Obj)
	f.Replace([]interface{}{b4Obj}, "")

	actualDeltasA := st.Pop(f)
	fmt.Print("actualDeltas %#v", actualDeltasA)

	actualDeltasB := st.Pop(f)
	fmt.Print("actualDeltas %#v", actualDeltasB)

}


func testIndexFunc(obj interface{}) ([]string, error) {
	pod := obj.(*v1.Pod)
	return []string{pod.Labels["foo"]}, nil
}

func TestGetIndexFuncValues() {
	index := st.NewIndexer(st.MetaNamespaceKeyFunc, st.Indexers{"byLabel": testIndexFunc})

	pod1 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "one", Labels: map[string]string{"foo": "bar"}}}
	pod2 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "two", Labels: map[string]string{"foo": "bar"}}}
	pod3 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "tre", Labels: map[string]string{"foo": "biz"}}}
	//pod4 := &v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "tre", Labels: map[string]string{"coo": "fei"}}}

	index.Add(pod1)
	index.Add(pod2)
	index.Add(pod3)
	//	index.Add(pod4)

	// ByIndex returns the stored objects whose set of indexed values
	// for the named index includes the given indexed value
	objResults, err := index.ByIndex("byLabel", "biz")
	if err != nil {
		fmt.Print("Unexpected error %v", err)
	}
	fmt.Print("Obj is %v", objResults)


	// ListIndexFuncValues returns all the indexed values of the given index
	keys := index.ListIndexFuncValues("byLabel")
	if len(keys) != 2 {
		fmt.Print("Expected 2 keys but got %v", len(keys))
	}

	for _, key := range keys {
		if key != "bar" && key != "biz" {
			fmt.Print("Expected only 'bar' or 'biz' but got %s", key)
		}
	}
	/*
		// Index returns the stored objects whose set of indexed values
		// intersects the set of indexed values of the given object, for
		// the named index
		indexResults, err := index.Index("byLabel", pod1)
		if err != nil {
			fmt.Print("Unexpected error %v", err)
		}
		fmt.Print("Obj is %v", indexResults)
	*/

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
