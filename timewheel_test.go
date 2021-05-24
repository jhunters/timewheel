package timewheel

import (
	"testing"
	"time"

	convey "github.com/smartystreets/goconvey/convey"
)

// TestTimeWheelCreateFaile test on create a new TimeWheel failed case
func TestTimeWheelCreateFaile(t *testing.T) {
	// to test with the case of New time wheel with error input parameters
	convey.Convey("New timewheel with error parameter", t, func() {
		convey.Convey("invalid slot number", func() {
			timewheel, err := New(100*time.Millisecond, 0)
			convey.So(timewheel, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("invalid interval", func() {
			timewheel, err := New(time.Duration(-1), 10)
			convey.So(timewheel, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

	})

}

// TestTimeWheelCreate test to create a new TimeWheel instance
func TestTimeWheelCreate(t *testing.T) {
	convey.Convey("New time wheel success", t, func() {
		var slotNum uint16 = 16
		var interval time.Duration = 100 * time.Millisecond
		timewheel, err := New(interval, slotNum)
		convey.So(timewheel, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)

		convey.So(timewheel.slotNum, convey.ShouldEqual, slotNum)
		convey.So(timewheel.interval, convey.ShouldEqual, interval)
		convey.So(timewheel.ticker, convey.ShouldBeNil)
	})

}

// TestTimeWheelStartAndTasks test add tasks after start
func TestTimeWheelStartAndTasks(t *testing.T) {

	convey.Convey("New time wheel success and start", t, func() {
		var slotNum uint16 = 16
		var interval time.Duration = 100 * time.Millisecond
		timewheel, err := New(interval, slotNum)
		timewheel.Start()
		defer timewheel.Stop()
		convey.So(timewheel, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)

		convey.So(timewheel.ticker, convey.ShouldNotBeNil)

		var tid taskid = 1
		convey.Convey("add a time task", func() {
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
			convey.So(tid, convey.ShouldEqual, tid)
			convey.So(err, convey.ShouldBeNil)
			convey.So(timewheel.currentTaskID, convey.ShouldEqual, tid+1)

			exsit := timewheel.HasTask(tid)
			convey.So(exsit, convey.ShouldBeTrue)

			time.Sleep(2 * time.Second)

			exsit = timewheel.HasTask(tid)
			convey.So(exsit, convey.ShouldBeFalse)
		})

		convey.Convey("remove a time task", func() {
			// remove time out task
			timewheel.removeTask(tid)
		})

		convey.Convey("add a time task again", func() {
			// test add task and timeout
			delay := 1 * time.Second
			tt := &Task{
				Data: map[string]int{"uid": 105626, "age": 100}, // call back data
				TimeoutCallback: func(task Task) { // call back function on time out
					if task.Elasped() < delay {
						t.Error("time out value is errored it should be large than specified time out.")
					}
				}}

			tid, err := timewheel.AddTask(delay*2*time.Duration(slotNum), *tt)
			convey.So(tid, convey.ShouldEqual, tid)
			convey.So(err, convey.ShouldBeNil)
			convey.So(timewheel.currentTaskID, convey.ShouldEqual, tid+1)

			time.Sleep(2 * time.Second)

			timewheel.RemoveTask(tid) // remove task id which it exsited
		})

	})

}

// TestTimeWheelExceed1CircleCase to test time wheel task exceed 1 circle case
func TestTimeWheelExceed1CircleCase(t *testing.T) {

	convey.Convey("test time wheel exceed circle case", t, func() {
		var slotNum uint16 = 16
		var interval time.Duration = 100 * time.Millisecond
		timewheel, err := New(interval, slotNum)
		timewheel.Start()
		defer timewheel.Stop()
		convey.So(timewheel, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("add a task", func() {
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
			convey.So(err, convey.ShouldBeNil)
			time.Sleep(2 * time.Second)

			exist := timewheel.HasTask(tid)
			if exist {
				t.Errorf("task id '%d' should not exist", tid)
			}

		})
	})

	var slotNum uint16 = 2
	timewheel, err := New(100*time.Millisecond, slotNum)
	if err != nil {
		t.Error(err)
	}

	timewheel.Start()

}
