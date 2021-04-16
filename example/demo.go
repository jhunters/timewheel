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

	// 启动时间轮
	tw.Start()

	var key string = "hello"
	t := &timewheel.Task{
		Key:  key,                                       // unqiue key
		Data: map[string]int{"uid": 105626, "age": 100}, // call back data
		TimeoutCallback: func(task timewheel.Task) { // call back function on time out
			fmt.Println("time out:", task.Delay(), task.Key, task.Data, task.Elasped())
		}}
	tw.AddTask(5*time.Second, *t)

	time.Sleep(10 * time.Second)

	// 删除定时器, 参数为添加定时器传递的唯一标识
	tw.RemoveTask(key)

	// 停止时间轮
	tw.Stop()
}
