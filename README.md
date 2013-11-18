# simpleq [![Build Status][1]][2]

A super simple Redis-backed queue. Based on [node-simpleq](https://github.com/Rafflecopter/node-simpleq). Documentation at [godoc](http://godoc.org/github.com/Rafflecopter/golang-simpleq/simpleq)

## Operation

Install:

```
go get github.com/Rafflecopter/golang-simpleq/simpleq
```

Creation:

```golang
import (
  "github.com/Rafflecopter/golang-simpleq/simpleq"
  "github.com/garyburd/redigo/redis"
)

func DoSimpleQ() {
  conn, err := redis.Dial("tcp", ":6379")
  if err != nil { panic(err); }

  q := simpleq.New(conn, "my-simpleq")
}
```

Operations:

- `q.Push(el, cb)` Returns number of elements in queue.
- `q.Pop(cb)` and `q.Bpop(cb)` (blocking) Returns element popped or null
- `q.Pull(el, cb)` Pull out a specific el (the highest/oldest el in the queue to be specific if elements are repeated) from the queue. Returns number of elements removed (0 or 1).
- `q1.Pullpipe(q2, el, cb)` Pull and push into another queue atomicly. Returns elements in second queue. (Note, that if el does not exist in q1, it will still be put into q2)
    - `q1.Spullpipe(q2, el, cb)` is a safe version which will not insert el into q2 unless it has been successfully pulled out of q1. This is done atomically using a lua script.
- `q1.Poppipe(q2, cb)` and `q1.Bpoppipe(q2, cb)` (blocking): Pop and push to another queue; returns popped element (also atomic).
- `q.Clear(cb)` Clear out the queue
- `q.List(cb)` List all elements in the queue

Listeners:

simpleq's can start listeners that run `Bpoppipe` or `Bpop` continuously and make callbacks when they occur. These two functions are `.Poplisten` and `.Poppipelisten`. The both accept the option:

- `max_out` (default: 0 ~ infinity) Maximum number of callbacks allowed to be out at one time.

The callbacks on a message from a listener pass a done function that must be called when processing is complete. If not in the same closure, `listener.done()` and `listener.emit('done');` will suffice.

_Note_: Calling listen will clone the redis connection, allowing `.push` to still work because another connection is being blocked. Calling a listen function more than once will result in a panic, as this is not prudent.

Examples below:

```golang
listener := q.Poplisten(10)

// or
listener := q.Poppipelisten(otherq, 10);

// listen for errors
go func () {
  for msg := range listener.Errors {
    // do stuff
  }
}()

// then, with messages
go func () {
  for msg := range listener.Messages {
    // do stuff
  }
}()

// eventually
ended_chan := listener.End()
<- ended_chan
```

## Tests

```
go test
```

## License

See LICENSE file.

[1]: https://travis-ci.org/Rafflecopter/golang-simpleq.png?branch=master
[2]: http://travis-ci.org/Rafflecopter/golang-simpleq