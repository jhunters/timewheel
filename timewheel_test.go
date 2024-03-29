// Copyright 2021 The baidu Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
			timewheel, err := New[string](100*time.Millisecond, 0)
			convey.So(timewheel, convey.ShouldBeNil)
			convey.So(err, convey.ShouldNotBeNil)
		})

		convey.Convey("invalid interval", func() {
			timewheel, err := New[string](time.Duration(-1), 10)
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
		timewheel, err := New[string](interval, slotNum)
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
		timewheel, err := New[map[string]int](interval, slotNum)
		timewheel.Start()
		defer timewheel.Stop()
		convey.So(timewheel, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)

		convey.So(timewheel.ticker, convey.ShouldNotBeNil)

		convey.Convey("add time task with error parameter", func() {
			delay := time.Duration(0) // case with delay = 0
			tt := newTask(delay, t)
			tid, err := timewheel.AddTask(delay, *tt)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(tid, convey.ShouldEqual, 0)

			delay = interval / 2 // case with delay smaller than interval
			tid, err = timewheel.AddTask(delay, *tt)
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(tid, convey.ShouldEqual, 0)
		})

		var tid taskid = 1
		convey.Convey("add a time task", func() {
			// test add task and timeout
			delay := 1 * time.Second
			tt := newTask(delay, t)
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
			tt := &Task[map[string]int]{
				Data: map[string]int{"uid": 105626, "age": 100}, // call back data
				TimeoutCallback: func(task Task[map[string]int]) { // call back function on time out
					if task.Elasped() < delay {
						t.Error("time out value is errored it should be large than specified time out.")
					}
				}}

			tid2, err := timewheel.AddTask(delay*2*time.Duration(slotNum), *tt)
			convey.So(tid, convey.ShouldEqual, tid2)
			convey.So(err, convey.ShouldBeNil)
			convey.So(timewheel.currentTaskID, convey.ShouldEqual, tid+1)

			time.Sleep(2 * time.Second)

			timewheel.RemoveTask(tid) // remove task id which it exsited
		})

	})

}

func newTask(delay time.Duration, t *testing.T) *Task[map[string]int] {
	tt := &Task[map[string]int]{
		Data: map[string]int{"uid": 105626, "age": 100}, // call back data
		TimeoutCallback: func(task Task[map[string]int]) { // call back function on time out
			if task.Elasped() < delay { // check elasped time should small than delay
				t.Error("time out value is errored it should be large than specified time out.")
			}
		}}
	return tt
}

// TestTimeWheelExceed1CircleCase to test time wheel task exceed 1 circle case
func TestTimeWheelExceed1CircleCase(t *testing.T) {

	convey.Convey("test time wheel exceed circle case", t, func() {
		var slotNum uint16 = 16
		var interval time.Duration = 100 * time.Millisecond
		timewheel, err := New[map[string]int](interval, slotNum)
		timewheel.Start()
		defer timewheel.Stop()
		convey.So(timewheel, convey.ShouldNotBeNil)
		convey.So(err, convey.ShouldBeNil)

		convey.Convey("add a task", func() {
			delay := 1 * time.Second
			tt := &Task[map[string]int]{
				Data: map[string]int{"uid": 105626, "age": 100}, // call back data
				TimeoutCallback: func(task Task[map[string]int]) { // call back function on time out
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
	timewheel, err := New[string](100*time.Millisecond, slotNum)
	if err != nil {
		t.Error(err)
	}

	timewheel.Start()

}
