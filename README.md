# timewheel

Pure golang implementation for timewheel.

[![Go Report Card](https://goreportcard.com/badge/github.com/jhunters/timewheel?style=flat-square)](https://goreportcard.com/report/github.com/jhunters/timewheel)
[![Build Status](https://travis-ci.com/jhunters/timewheel.svg?branch=main&status=started)](https://travis-ci.org/jhunters/timewheel)
[![codecov](https://codecov.io/gh/jhunters/timewheel/branch/main/graph/badge.svg?token=dhBirUo4qL)](https://codecov.io/gh/jhunters/timewheel)
[![Releases](https://img.shields.io/github/release/jhunters/timewheel/all.svg?style=flat-square)](https://github.com/jhunters/timewheel/releases)
[![Godoc](http://img.shields.io/badge/go-documentation-blue.svg?style=flat-square)](https://godoc.org/github.com/jhunters/timewheel)
[![LICENSE](https://img.shields.io/github/license/jhunters/timewheel.svg?style=flat-square)](https://github.com/jhunters/timewheel/blob/master/LICENSE)

![pic/timewheel.png](pic/timewheel.png)

## Usage

### base method

create timewheel

```
// 初始化时间轮
// 第一个参数为tick刻度, 即时间轮多久转动一次
// 第二个参数为时间轮槽slot数量
tw, err := timewheel.New(100*time.Millisecond, 300)
if err != nil {
    panic(err)
}

tw.Start()

```


add delay task

```
// create a task bind with key, data and  time out call back function.
t := &timewheel.Task{
    Data: map[string]int{"uid": 105626, "age": 100}, // business data
    TimeoutCallback: func(task timewheel.Task) { // call back function on time out
        // process someting after time out happened. 
        fmt.Println("time out:", task.Delay(), task.Key, task.Data, task.Elasped())
    }}

// add task and return unique task id
taskid := tw.AddTask(5*time.Second, *t) // add delay task

```

remove delay task

```
tw.Remove(taskid)
```

check task

```
tw.HasTask(taskid)
```

close time wheel

```
tw.Stop()
```
## example

[example/demo.go](example/demo.go)
