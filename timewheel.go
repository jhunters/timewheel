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
package timewheel

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// 超时发生时，回调的数据
type taskdata interface{}

// 超时任务回调函数
type TimeoutCallbackFn func(Task)

// task id
type taskid uint64

// Task task struct
type Task struct {
	delay           time.Duration
	Data            taskdata
	TimeoutCallback TimeoutCallbackFn
	elasped         time.Duration
}

// Delay return delay time
func (task *Task) Delay() time.Duration {
	return task.delay
}

// Elasped to get task
func (t *Task) Elasped() time.Duration {
	return t.elasped
}

// TaskSlot a task with target slot info
type TaskSlot struct {
	delay  time.Duration // 延迟时间
	circle uint16        // 时间轮需要转动几圈，每一圈，circle减一。 只有circle为0时，才是当前槽要触发的超时任务
	task   *Task
	now    time.Time
	end    time.Time
	taskid taskid
}

// TimeWheel 时间轮
type TimeWheel struct {
	interval time.Duration // 指针每隔多久往前移动一格
	ticker   *time.Ticker
	slots    []*list.List // 时间轮槽
	// key: 定时器唯一标识 value: 定时器所在的槽, 主要用于删除定时器, 不会出现并发读写，不加锁直接访问
	timer             map[taskid]uint16
	currentPos        uint16        // 当前指针指向哪一个槽
	slotNum           uint16        // 槽数量
	addTaskChannel    chan TaskSlot // 新增任务channel
	removeTaskChannel chan taskid   // 删除任务channel
	stopChannel       chan bool     // 停止定时器channel
	currentTaskID     taskid        // 最新任务ID
	locker            sync.Mutex    // task id locker
}

// New 创建时间轮
func New(interval time.Duration, slotNum uint16) (*TimeWheel, error) {
	if interval <= 0 || slotNum <= 0 {
		return nil, errors.New("invalid parameter 'interval' or 'slotNum' must be large than zero")
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		timer:             make(map[taskid]uint16),
		currentPos:        0,
		slotNum:           slotNum,
		addTaskChannel:    make(chan TaskSlot),
		removeTaskChannel: make(chan taskid),
		stopChannel:       make(chan bool),
		currentTaskID:     1,
	}

	tw.initSlots()

	return tw, nil
}

// 初始化槽，每个槽指向一个双向链表
func (tw *TimeWheel) initSlots() {
	for i := uint16(0); i < tw.slotNum; i++ {
		tw.slots[i] = list.New()
	}
}

// Start 启动时间轮
func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	go tw.start()
}

// start time wheel. to handle all chan listener in the loop
func (tw *TimeWheel) start() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tickHandler()
		case task := <-tw.addTaskChannel:
			tw.addTask(&task)
		case key := <-tw.removeTaskChannel:
			tw.removeTask(key)
		case <-tw.stopChannel:
			tw.ticker.Stop()
			return
		}
	}
}

// Stop 停止时间轮
func (tw *TimeWheel) Stop() {
	tw.stopChannel <- true
}

// AddTimer 添加定时器 key为定时器唯一标识
func (tw *TimeWheel) AddTask(delay time.Duration, task Task) (taskid, error) {
	if delay <= 0 {
		return 0, errors.New("parameter 'delay' must be large than zero")
	}
	if delay <= tw.interval { // 延迟触发的时间不能小于等于 interval 间隔
		return 0, fmt.Errorf("parameter 'delay'=%d  should not less than interval = %d ", delay, tw.interval)
	}

	task.delay = delay
	tw.locker.Lock()
	tid := tw.currentTaskID
	tw.currentTaskID = (taskid)(atomic.AddUint64((*uint64)(&tw.currentTaskID), uint64(1)))
	tw.locker.Unlock()
	tw.addTaskChannel <- TaskSlot{delay: delay, now: time.Now(), taskid: tid, task: &task}
	return tid, nil
}

// 新增任务到链表中
func (tw *TimeWheel) addTask(taskSlot *TaskSlot) {
	pos, circle := tw.getPositionAndCircle(taskSlot.delay)
	taskSlot.circle = circle

	tw.slots[pos].PushBack(taskSlot)

	if taskSlot.taskid > 0 {
		tw.timer[taskSlot.taskid] = pos
	}
}

// 获取定时器在槽中的位置, 时间轮需要转动的圈数
func (tw *TimeWheel) getPositionAndCircle(d time.Duration) (pos uint16, circle uint16) {
	delaySeconds := int64(d.Milliseconds())
	intervalSeconds := int64(tw.interval.Milliseconds())
	circle = uint16(delaySeconds / intervalSeconds / int64(tw.slotNum))
	pos = uint16(int64(tw.currentPos)+delaySeconds/intervalSeconds) % tw.slotNum

	return
}

// RemoveTimer 删除定时器 key为添加定时器时传递的定时器唯一标识
func (tw *TimeWheel) RemoveTask(key taskid) {
	if key > 0 { // taskid must large than zero
		tw.removeTaskChannel <- key
	}
}

// 从链表中删除任务
func (tw *TimeWheel) removeTask(key taskid) {
	// 获取定时器所在的槽
	position, ok := tw.timer[key]
	if !ok { // key not exist
		return
	}
	delete(tw.timer, key) // remove time and pos map key
	// 获取槽指向的链表
	l := tw.slots[position]
	for e := l.Front(); e != nil; {
		taskSlot := e.Value.(*TaskSlot)
		if taskSlot.taskid == key {
			l.Remove(e)
		}

		e = e.Next()
	}
}

// HasTask to check task id exist
func (tw *TimeWheel) HasTask(key taskid) bool {
	// 获取定时器所在的槽
	position, ok := tw.timer[key]
	if !ok { // key not exist
		return false
	}

	// 获取槽指向的链表
	l := tw.slots[position]
	for e := l.Front(); e != nil; {
		taskSlot := e.Value.(*TaskSlot)
		if taskSlot.taskid == key {
			return true
		}

		e = e.Next()
	}

	return false
}

// 时间轮走动到slot位置时，触发处理
func (tw *TimeWheel) tickHandler() {
	l := tw.slots[tw.currentPos]
	tw.scanAndRunTask(l)
	if tw.currentPos == tw.slotNum-1 {
		tw.currentPos = 0
	} else {
		tw.currentPos++
	}
}

// 扫描链表中过期定时器, 并执行回调函数
func (tw *TimeWheel) scanAndRunTask(l *list.List) {
	for e := l.Front(); e != nil; {
		taskSlot := e.Value.(*TaskSlot)
		if taskSlot.circle > 0 {
			taskSlot.circle--
			e = e.Next()
			continue
		}

		taskSlot.end = time.Now()
		taskSlot.task.elasped = taskSlot.end.Sub(taskSlot.now)
		go taskSlot.task.TimeoutCallback(*taskSlot.task)
		next := e.Next()
		l.Remove(e)
		if taskSlot.taskid > 0 {
			delete(tw.timer, taskSlot.taskid)
		}
		e = next // 往后遍历
	}
}
