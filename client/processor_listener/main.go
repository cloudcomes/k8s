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
	"k8s.io/apimachinery/pkg/util/wait"
	buffer "k8s.io/utils/buffer"
	"math/rand"
	"time"
)


type ResourceEventHandler struct {
	receivedItemNames []string
	name              string

}

func (r *ResourceEventHandler) onUpdate (newObj, oldObj interface{}) {
	fmt.Println("Update Handler newObj is %v,oldObj is %v", newObj, oldObj)
}

func (r *ResourceEventHandler) onAdd (newObj interface{}) {
	fmt.Println("Add Handler %v", newObj)
}

func (r *ResourceEventHandler) onDel (oldObj interface{}) {
	fmt.Println("Del Handler%v", oldObj)
}

func newResourceHandler(name string,received ...string) *ResourceEventHandler {
	l := &ResourceEventHandler{
		receivedItemNames: received,
		name:              name,
	}
	return l
}


type updateNotification struct {
	newObj  interface{}
	oldObj  interface{}
}

type addNotification struct {
	newObj  interface{}
}

type deleteNotification struct {
	oldObj  interface{}
}


type processorListener struct {
	nextCh chan interface{}
	addCh  chan interface{}

	handler *ResourceEventHandler
	pendingNotifications *buffer.RingGrowing
}


func newProcessListener(handler ResourceEventHandler, requestedResyncPeriod, resyncPeriod time.Duration, now time.Time, bufferSize int) *processorListener {
	ret := &processorListener{
		nextCh:                make(chan interface{}),
		addCh:                 make(chan interface{}),
		handler:               &handler,
		pendingNotifications:  buffer.NewRingGrowing(bufferSize),
	}

	return ret
}


func (p *processorListener) add(notification interface{}) {
	p.addCh <- notification
}


func (p *processorListener) pop() {
	defer close(p.nextCh)
	var nextCh chan <- interface{}
	var notification interface{}

	for {
		select {

		//notification写入nextCh。然后判断buffer中是否还有notification，如果还有，则不停的传递给 nextCh
		case nextCh <- notification:
			var ok bool

			notification, ok = p.pendingNotifications.ReadOne()
			fmt.Println("Pop:Get notification from Ring Buffer%v", notification, ok, nextCh, p.nextCh)
			// 如果没有notification，那就把nextCh再次设置为nil
			if !ok {
				nextCh = nil
			}

		// 从p.addCh读notification
		// 不停的从 addCh中获得最新notificationAdd，然后判断是否存在于buffer，
		// 如果存在则把事件添加到 buffer 中，如果不存在则尝试推给 nextCh
		case notificationAdd, ok :=  <- p.addCh:
			fmt.Println("Pop:Get notification from addCh%v", notificationAdd, ok, notification == nil)
			//读取失败退出
			if !ok {
				return
			}
			// 判断buffer里是否还有有数据，如果没有则推给nextCh
			if notification == nil {
				notification = notificationAdd
				nextCh = p.nextCh
			//如果buffer里还有数据添，则把notificationAdd缓存到buffer里
			} else {
				p.pendingNotifications.WriteOne(notificationAdd)
			}
		}
	}
}


func (p *processorListener) run () {

	stopCh := make(chan struct{})

	wait.Until(func () {
		for next := range p.nextCh {
			switch notification := next.(type) {
			case updateNotification:
				p.handler.onUpdate(notification.oldObj, notification.newObj)
			case addNotification:
				p.handler.onAdd(notification.newObj)
			case deleteNotification:
				p.handler.onDel(notification.oldObj)
			}
		}
		close(stopCh)
	}, 1* time.Second, stopCh)
}



func main() {

	var notification interface{}
	// listener 1, never resync
	listener1 := newResourceHandler("listener1","pod2")
	// Preallocate enough space so that benchmark does not run out of it
	pl := newProcessListener(*listener1,0, 0, time.Now(), 1024*1024)


	var wg wait.Group
	defer wg.Wait()       // Wait for .run and .pop to stop
	defer close(pl.addCh) // Tell .run and .pop to stop
	wg.Start(pl.run)
	wg.Start(pl.pop)

	stopCh := make(chan struct{})

	wait.Until(func() {
		number := rand.Intn(100)
		println(number)

		if number < 20 {
			notification = addNotification{newObj: number}
		} else if number < 50 {
			notification = updateNotification{newObj:number, oldObj:number}
		} else {
			notification = deleteNotification{oldObj: number}
		}
		fmt.Println("notification is %+v", notification)

		pl.add(notification)
	}, 1*time.Second, stopCh)

}


