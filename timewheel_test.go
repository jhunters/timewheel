package timewheel

import (
	"testing"
	"time"
)

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

	tid := timewheel.AddTask(delay, *tt)
	if tid != 1 {
		t.Error("the first task id should be 1")
	}

	if timewheel.currentTaskID != 2 {
		t.Error("the currentTaskID should be 1 after one task added.", timewheel.currentTaskID)
	}

	time.Sleep(2 * time.Second)
	// remove time out task
	timewheel.removeTask(tid)

	// add second task
	tid = timewheel.AddTask(delay, *tt)
	time.Sleep(100 * time.Millisecond)
	timewheel.removeTask(tid)

	time.Sleep(2 * time.Second)
}
