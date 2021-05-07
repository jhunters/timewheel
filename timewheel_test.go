package timewheel

import (
	"testing"
	"time"
)

// TestTimeWheelCreateFaile test on create a new TimeWheel failed case
func TestTimeWheelCreateFaile(t *testing.T) {
	timewheel, err := New(100*time.Millisecond, 0)
	if err == nil {
		t.Error("New timewheel should return fail while slot num is zero.")
	}

	if timewheel != nil {
		t.Error("New timewheel should return nil while slot num is zero.")
	}
}

// TestTimeWheelCreate test to create a new TimeWheel instance
func TestTimeWheelCreate(t *testing.T) {
	var slotNum uint16 = 16
	timewheel, err := New(100*time.Millisecond, slotNum)
	if err != nil {
		t.Error(err)
	}

	if timewheel == nil {
		t.Error("timewheel.New should not return nil without any error.")
	}

	if timewheel.slotNum != slotNum {
		t.Error("slot Num should be ", slotNum, " but actually is ", timewheel.slotNum)
	}

	if timewheel.interval != 100*time.Millisecond {
		t.Error("timewheel interval should be ", 1, " but actually is ", timewheel.interval)
	}

	if timewheel.ticker != nil {
		t.Error("timewheel.timer should be nil after created.")
	}
}

// TestTimeWheelStartAndTasks test add tasks after start
func TestTimeWheelStartAndTasks(t *testing.T) {
	var slotNum uint16 = 16
	timewheel, err := New(100*time.Millisecond, slotNum)
	if err != nil {
		t.Error(err)
	}

	timewheel.Start()
	if timewheel.ticker == nil {
		t.Error("timewheel.ticker should not be nil after started.")
	}

	// test add task and timeout
	delay := 1 * time.Second
	tt := &Task{
		Data: map[string]int{"uid": 105626, "age": 100}, // call back data
		TimeoutCallback: func(task Task) { // call back function on time out
			if task.Elasped() < delay {
				t.Error("time out value is errored it should be large than specified time out.")
			}
		}}

	tid, err := timewheel.AddTask(delay, *tt)
	if tid != 1 || err != nil {
		t.Error("the first task id should be 1")
	}

	if timewheel.currentTaskID != 2 {
		t.Error("the currentTaskID should be 1 after one task added.", timewheel.currentTaskID)
	}

	time.Sleep(2 * time.Second)
	// remove time out task
	timewheel.removeTask(tid)

	// add second task
	tid, err = timewheel.AddTask(delay*2*time.Duration(slotNum), *tt)
	if err != nil {
		t.Error(err)
	}
	time.Sleep(100 * time.Millisecond)
	timewheel.RemoveTask(tid)

	// add failed task with delay smaller than interval
	_, err = timewheel.AddTask(100*time.Millisecond, *tt)
	if err == nil {
		t.Error("err should not be nil due to add a delay value smaller than interval task")
	}

	time.Sleep(2 * time.Second)

	timewheel.Stop()
}

// TestTimeWheelExceed1CircleCase to test time wheel task exceed 1 circle case
func TestTimeWheelExceed1CircleCase(t *testing.T) {

	var slotNum uint16 = 2
	timewheel, err := New(100*time.Millisecond, slotNum)
	if err != nil {
		t.Error(err)
	}

	timewheel.Start()

	// test add task and timeout
	delay := 1 * time.Second
	tt := &Task{
		Data: map[string]int{"uid": 105626, "age": 100}, // call back data
		TimeoutCallback: func(task Task) { // call back function on time out
			if task.Delay() != delay {
				t.Error("time delay value should be ", delay, " but actually is ", task.Delay())
			}
			if task.Elasped() < delay {
				t.Error("time out value is errored it should be large than specified time out.")
			}
		}}
	tid, err := timewheel.AddTask(delay, *tt)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(2 * time.Second)

	exist := timewheel.HasTask(tid)
	if exist {
		t.Errorf("task id '%d' should not exist", tid)
	}

	timewheel.Stop()
}
