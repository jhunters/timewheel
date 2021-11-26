/*
 * @Author: Malin Xie
 * @Description:
 * @Date: 2021-05-25 16:28:36
 */
// Copyright 2021 The baidu Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package timewheel_test

import (
	"fmt"
	"time"

	"github.com/jhunters/timewheel"
)

// ExampleTimeWheel example code for simple timewheel api usage.
func ExampleTimeWheel() {
	// 初始化时间轮
	// 第一个参数为tick刻度, 即时间轮多久转动一次
	// 第二个参数为时间轮槽slot数量
	tw, err := timewheel.New(100*time.Millisecond, 300)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("time wheel created.")

	// 启动时间轮
	tw.Start()
	fmt.Println("time wheel started.")

	t := &timewheel.Task{
		Data: map[string]int{"uid": 105626, "age": 100}, // call back data
		TimeoutCallback: func(task timewheel.Task) { // call back function on time out

			data, _ := task.Data.(map[string]int)
			fmt.Println("time out:", task.Delay(), data["uid"], data["age"] /*, task.Elasped()*/)
		}}

	// add task and return unique task id
	taskid, err := tw.AddTask(5*time.Second, *t)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("add a new task. taskid=", taskid)

	// before time out we remove the task
	tw.RemoveTask(taskid)
	fmt.Println("remove task. taskid=", taskid)

	// add a new task again
	taskid, err = tw.AddTask(5*time.Second, *t)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("add a new task.  taskid=", taskid)

	fmt.Println("wait 10 seconds here.")
	time.Sleep(10 * time.Second)

	// 删除定时器, 参数为添加定时器传递的唯一标识
	tw.RemoveTask(taskid)

	// 停止时间轮
	tw.Stop()

	// Output: time wheel created.
	// time wheel started.
	// add a new task. taskid= 1
	// remove task. taskid= 1
	// add a new task.  taskid= 2
	// wait 10 seconds here.
	// time out: 5s 105626 100
}
