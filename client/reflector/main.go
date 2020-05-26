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
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"flag"
	"k8s.io/api/core/v1"
	"strconv"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)


func main() {

	//TestReflectorListAndWatch()
	TestReflectorListAndWatch01()
}



type testLW struct {
	ListFunc  func(options metav1.ListOptions) (runtime.Object, error)
	WatchFunc func(options metav1.ListOptions) (watch.Interface, error)
}

func (t *testLW) List(options metav1.ListOptions) (runtime.Object, error) {
	return t.ListFunc(options)
}
func (t *testLW) Watch(options metav1.ListOptions) (watch.Interface, error) {
	return t.WatchFunc(options)
}


func TestReflectorListAndWatch() {
	createdFakes := make(chan *watch.FakeWatcher)

	// The ListFunc says that it's at revision 1. Therefore, we expect our WatchFunc
	// to get called at the beginning of the watch with 1, and again with 3 when we
	// inject an error.
	expectedRVs := []string{"1", "3"}
	lw := &testLW{
		WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
			rv := options.ResourceVersion
			fw := watch.NewFake()
			if e, a := expectedRVs[0], rv; e != a {
				klog.Errorf("Expected rv %v, but got %v", e, a)
			}
			expectedRVs = expectedRVs[1:]
			// channel is not buffered because the for loop below needs to block. But
			// we don't want to block here, so report the new fake via a go routine.
			go func() { createdFakes <- fw }()
			return fw, nil
		},
		ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
			return &v1.PodList{ListMeta: metav1.ListMeta{ResourceVersion: "1"}}, nil
		},
	}
	s := cache.NewFIFO(cache.MetaNamespaceKeyFunc)
	r := cache.NewReflector(lw, &v1.Pod{}, s, 0)
	go r.ListAndWatch(wait.NeverStop)

	ids := []string{"foo", "bar", "baz", "qux", "zoo"}
	var fw *watch.FakeWatcher
	for i, id := range ids {
		if fw == nil {
			fw = <-createdFakes
		}
		sendingRV := strconv.FormatUint(uint64(i+2), 10)
		fw.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: id, ResourceVersion: sendingRV}})
		if sendingRV == "3" {
			// Inject a failure.
			fw.Stop()
			fw = nil
		}
	}

	// Verify we received the right ids with the right resource versions.
	for i, id := range ids {
		pod := cache.Pop(s).(*v1.Pod)
		if e, a := id, pod.Name; e != a {
			klog.Errorf("%v: Expected %v, got %v", i, e, a)
		}
		if e, a := strconv.FormatUint(uint64(i+2), 10), pod.ResourceVersion; e != a {
			klog.Errorf("%v: Expected %v, got %v", i, e, a)
		}
	}

	if len(expectedRVs) != 0 {
		klog.Errorf("called watchStarter an unexpected number of times")
	}
}

func TestReflectorListAndWatch01() {
	//createdFakes := make(chan *watch.FakeWatcher)

	// The ListFunc says that it's at revision 1. Therefore, we expect our WatchFunc
	// to get called at the beginning of the watch with 1, and again with 3 when we
	// inject an error.
	expectedRVs := []string{"1", "3"}

	lw := getLW()
	s := cache.NewFIFO(cache.MetaNamespaceKeyFunc)
	r := cache.NewReflector(lw, &v1.Pod{}, s, 0)
	go r.ListAndWatch(wait.NeverStop)

	ids := []string{"gohttpdemo-5cb6954c5c-42bkc", "gohttpdemo-5cb6954c5c-p4wz6"}


	// Verify we received the right ids with the right resource versions.
	for i, id := range ids {
		pod := cache.Pop(s).(*v1.Pod)
		if e, a := id, pod.Name; e != a {
			klog.Errorf("%v: Expected %v, got %v", i, e, a)
			fmt.Print("%v: Expected %v, got %v", i, e, a)
		}
		if e, a := strconv.FormatUint(uint64(i+2), 10), pod.ResourceVersion; e != a {
			klog.Errorf("%v: Expected %v, got %v", i, e, a)
		}
	}

	if len(expectedRVs) != 0 {
		klog.Errorf("called watchStarter an unexpected number of times")
	}
}


func getLW() cache.ListerWatcher {

   var kubeconfig string
   var master string

   flag.StringVar(&kubeconfig, "kubeconfig", "./config", "absolute path to the kubeconfig file")
   //flag.String("kubeconfig", "./config", "Path to a kube config. Only required if out-of-cluster.")
   flag.StringVar(&master, "master", "", "master url")
   flag.Parse()

   // creates the connection
   config, err := clientcmd.BuildConfigFromFlags(master, kubeconfig)
   if err != nil {
      klog.Fatal(err)
   }

   // creates the clientset
   clientset, err := kubernetes.NewForConfig(config)
   if err != nil {
      klog.Fatal(err)
    }

   // create the pod watcher
   podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Everything())
   //podListWatcher := cache.NewListWatchFromClient(clientset.CoreV1().RESTClient(), "pods", v1.NamespaceDefault, fields.Set{"spec.nodeName": "node3"}.AsSelector())

   podListWatcher.DisableChunking = true

   return podListWatcher

}


