package main

import (
	"fmt"
	"time"

	"github.com/jhunters/timewheel"
)

func main() {

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
			fmt.Println("time out:", task.Delay(), task.Data, task.Elasped())
		}}

	// add task and return unique task id
	taskid, _ := tw.AddTask(5*time.Second, *t)
	fmt.Println("add a new task. taskid=", taskid)

	// before time out we remove the task
	tw.RemoveTask(taskid)
	fmt.Println("remove task. taskid=", taskid)

	// add a new task again
	taskid, _ = tw.AddTask(5*time.Second, *t)
	fmt.Println("add a new task.  taskid=", taskid)

	fmt.Println("wait 10 seconds here.")
	time.Sleep(10 * time.Second)

	// 删除定时器, 参数为添加定时器传递的唯一标识
	tw.RemoveTask(taskid)

	// 停止时间轮
	tw.Stop()
}
