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
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	cache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog"

	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	fcache "k8s.io/client-go/tools/cache/testing"
)


func main() {

	//Example()
	//ExampleLW()
	ExampleNewInformer()

}

func Example() {
	// source simulates an apiserver object endpoint.
	source := fcache.NewFakeControllerSource()

	// This will hold the downstream state, as we know it.
	downstream := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)

	// This will hold incoming changes. Note how we pass downstream in as a
	// KeyLister, that way resync operations will result in the correct set
	// of update/delete deltas.
	fifo := cache.NewDeltaFIFO(cache.MetaNamespaceKeyFunc, downstream)

	// Let's do threadsafe output to get predictable test results.
	deletionCounter := make(chan string, 1000)

	cfg := &cache.Config{
		Queue:            fifo,
		ListerWatcher:    source,
		ObjectType:       &v1.Pod{},
		FullResyncPeriod: time.Millisecond * 100,
		RetryOnError:     false,

		// Let's implement a simple controller that just deletes
		// everything that comes in.
		Process: func(obj interface{}) error {
			// Obj is from the Pop method of the Queue we make above.
			newest := obj.(cache.Deltas).Newest()

			if newest.Type != cache.Deleted {
				// Update our downstream store.
				err := downstream.Add(newest.Object)
				if err != nil {
					return err
				}

				// Delete this object.
				source.Delete(newest.Object.(runtime.Object))
			} else {
				// Update our downstream store.
				err := downstream.Delete(newest.Object)
				if err != nil {
					return err
				}

				// fifo's KeyOf is easiest, because it handles
				// DeletedFinalStateUnknown markers.
				key, err := fifo.KeyOf(newest.Object)
				if err != nil {
					return err
				}

				// Report this deletion.
				deletionCounter <- key
			}
			return nil
		},
	}

	// Create the controller and run it until we close stop.
	stop := make(chan struct{})
	defer close(stop)
	go cache.New(cfg).Run(stop)

	// Let's add a few objects to the source.
	testIDs := []string{"a-hello", "b-controller", "c-framework"}
	for _, name := range testIDs {
		// Note that these pods are not valid-- the fake source doesn't
		// call validation or anything.
		source.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}})
	}

	// Let's wait for the controller to process the things we just added.
	outputSet := sets.String{}
	for i := 0; i < len(testIDs); i++ {
		outputSet.Insert(<-deletionCounter)
	}

	for _, key := range outputSet.List() {
		fmt.Println(key)
	}
	// Output:
	// a-hello
	// b-controller
	// c-framework
}

func ExampleLW() {
	// source simulates an apiserver object endpoint.
	source := getLW()

	// This will hold the downstream state, as we know it.
	downstream := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)

	// This will hold incoming changes. Note how we pass downstream in as a
	// KeyLister, that way resync operations will result in the correct set
	// of update/delete deltas.
	fifo := cache.NewDeltaFIFO(cache.MetaNamespaceKeyFunc, downstream)

	// Let's do threadsafe output to get predictable test results.
	deletionCounter := make(chan string, 1000)

	cfg := &cache.Config{
		Queue:            fifo,
		ListerWatcher:    source,
		ObjectType:       &v1.Pod{},
		FullResyncPeriod: time.Millisecond * 100,
		RetryOnError:     false,

		// Let's implement a simple controller that just deletes
		// everything that comes in.
		Process: func(obj interface{}) error {
			// Obj is from the Pop method of the Queue we make above.
			newest := obj.(cache.Deltas).Newest()

			if newest.Type != cache.Deleted {
				// Update our downstream store.
				err := downstream.Add(newest.Object)
				if err != nil {
					return err
				}
				// fifo's KeyOf is easiest, because it handles
				// DeletedFinalStateUnknown markers.
				key, err := fifo.KeyOf(newest.Object)
				if err != nil {
					return err
				}

				klog.Info("The Process of controller is running")

				// Report this Add.
				deletionCounter <- key
				// Delete this object.
			//	source.Delete(newest.Object.(runtime.Object))
			} else {
				// Update our downstream store.
				err := downstream.Delete(newest.Object)
				if err != nil {
					return err
				}

				// fifo's KeyOf is easiest, because it handles
				// DeletedFinalStateUnknown markers.
				key, err := fifo.KeyOf(newest.Object)
				if err != nil {
					return err
				}

				// Report this deletion.
				deletionCounter <- key
			}
			return nil
		},
	}

	// Create the controller and run it until we close stop.
	stop := make(chan struct{})
	defer close(stop)
	go cache.New(cfg).Run(stop)

	// Let's wait for the controller to process the things we just added.
	outputSet := sets.String{}
	for i := 0; i < 3; i++ {
		outputSet.Insert(<-deletionCounter)
	}

	for _, key := range outputSet.List() {
		fmt.Println(key)
	}
	// Output:
	// a-hello
	// b-controller
	// c-framework
}



func ExampleNewInformer() {
	// source simulates an apiserver object endpoint.
	source := fcache.NewFakeControllerSource()

	// Let's do threadsafe output to get predictable test results.
	deletionCounter := make(chan string, 1000)

	// Make a controller that immediately deletes anything added to it, and
	// logs anything deleted.
	_, controller := cache.NewInformer(
		source,
		&v1.Pod{},
		time.Millisecond*100,
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				source.Delete(obj.(runtime.Object))
				klog.Info("The controller of Informer is running")
			},
			DeleteFunc: func(obj interface{}) {
				key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
				if err != nil {
					key = "oops something went wrong with the key"
				}

				// Report this deletion.
				deletionCounter <- key
			},
		},
	)

	// Run the controller and run it until we close stop.
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)

	// Let's add a few objects to the source.
	testIDs := []string{"a-hello", "b-controller", "c-framework"}
	for _, name := range testIDs {
		// Note that these pods are not valid-- the fake source doesn't
		// call validation or anything.
		source.Add(&v1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name}})
	}

	// Let's wait for the controller to process the things we just added.
	outputSet := sets.String{}
	for i := 0; i < len(testIDs); i++ {
		outputSet.Insert(<-deletionCounter)
	}

	for _, key := range outputSet.List() {
		fmt.Println(key)
	}
	// Output:
	// a-hello
	// b-controller
	// c-framework
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


