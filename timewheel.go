package timewheel

import (
	"container/list"
	"errors"
	"time"
)

// 添加的超时任务key, 唯一
type taskkey interface{}

// 超时发生时，回调的数据
type taskdata interface{}

// 超时任务回调函数
type TimeoutCallbackFn func(Task)

// Task task struct
type Task struct {
	delay           time.Duration
	Key             taskkey
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
}

// TimeWheel 时间轮
type TimeWheel struct {
	interval time.Duration // 指针每隔多久往前移动一格
	ticker   *time.Ticker
	slots    []*list.List // 时间轮槽
	// key: 定时器唯一标识 value: 定时器所在的槽, 主要用于删除定时器, 不会出现并发读写，不加锁直接访问
	timer             map[interface{}]uint16
	currentPos        uint16           // 当前指针指向哪一个槽
	slotNum           uint16           // 槽数量
	addTaskChannel    chan TaskSlot    // 新增任务channel
	removeTaskChannel chan interface{} // 删除任务channel
	stopChannel       chan bool        // 停止定时器channel
}

// New 创建时间轮
func New(interval time.Duration, slotNum uint16) (*TimeWheel, error) {
	if interval <= 0 || slotNum <= 0 {
		return nil, errors.New("invalid parameter 'interval' or 'slotNum' must be large than zero")
	}
	tw := &TimeWheel{
		interval:          interval,
		slots:             make([]*list.List, slotNum),
		timer:             make(map[interface{}]uint16),
		currentPos:        0,
		slotNum:           slotNum,
		addTaskChannel:    make(chan TaskSlot),
		removeTaskChannel: make(chan interface{}),
		stopChannel:       make(chan bool),
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
func (tw *TimeWheel) AddTask(delay time.Duration, task Task) {
	if delay <= 0 {
		return
	}
	task.delay = delay
	tw.addTaskChannel <- TaskSlot{delay: delay, now: time.Now(), task: &task}
}

// 新增任务到链表中
func (tw *TimeWheel) addTask(taskSlot *TaskSlot) {
	pos, circle := tw.getPositionAndCircle(taskSlot.delay)
	taskSlot.circle = circle

	tw.slots[pos].PushBack(taskSlot)

	if taskSlot.task.Key != nil {
		tw.timer[taskSlot.task.Key] = pos
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
func (tw *TimeWheel) RemoveTask(key interface{}) {
	if key == nil {
		return
	}
	tw.removeTaskChannel <- key
}

// 从链表中删除任务
func (tw *TimeWheel) removeTask(key interface{}) {
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
		if taskSlot.task.Key == key {
			l.Remove(e)
		}

		e = e.Next()
	}
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
		if taskSlot.task.Key != nil {
			delete(tw.timer, taskSlot.task.Key)
		}
		e = next
	}
}