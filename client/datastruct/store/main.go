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
	"k8s.io/apimachinery/pkg/util/sets"
	st "k8s.io/client-go/tools/cache"
	"k8s.io/klog"
)

func main() {

	TestCache()

}

// Test public interface
func doTestStore(store  st.Store) {
	mkObj := func(id string, val string) testStoreObject {
		return testStoreObject{id: id, val: val}
	}

	store.Add(mkObj("foo", "bar"))
	if item, ok, _ := store.Get(mkObj("foo", "")); !ok {
		klog.Fatal("didn't find inserted item")
	} else {
		if e, a := "bar", item.(testStoreObject).val; e != a {
			klog.Fatal("expected %v, got %v", e, a)
		}
	}
	store.Update(mkObj("foo", "baz"))
	if item, ok, _ := store.Get(mkObj("foo", "")); !ok {
		klog.Fatal("didn't find inserted item")
	} else {
		if e, a := "baz", item.(testStoreObject).val; e != a {
			klog.Fatal("expected %v, got %v", e, a)
		}
	}
	store.Delete(mkObj("foo", ""))
	if _, ok, _ := store.Get(mkObj("foo", "")); ok {
		klog.Fatal("found deleted item??")
	}

	// Test List.
	store.Add(mkObj("a", "b"))
	store.Add(mkObj("c", "d"))
	store.Add(mkObj("e", "e"))
	{
		found := sets.String{}
		for _, item := range store.List() {
			found.Insert(item.(testStoreObject).val)
		}
		if !found.HasAll("b", "d", "e") {
			klog.Fatal("missing items, found: %v", found)
		}
		if len(found) != 3 {
			klog.Fatal("extra items")
		}
	}

	// Test Replace.
	store.Replace([]interface{}{
		mkObj("foo", "foo"),
		mkObj("bar", "bar"),
	}, "0")

	{
		found := sets.String{}
		for _, item := range store.List() {
			found.Insert(item.(testStoreObject).val)
		}
		if !found.HasAll("foo", "bar") {
			klog.Fatal("missing items")
		}
		if len(found) != 2 {
			klog.Fatal("extra items")
		}
	}
}


type testStoreObject struct {
	id  string
	val string
}

func testStoreKeyFunc(obj interface{}) (string, error) {
	return obj.(testStoreObject).id, nil
}

func TestCache() {
	doTestStore(st.NewStore(testStoreKeyFunc))
}



