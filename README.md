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
  pool := redis.NewPool(func () (redis.Conn, error) {
    return Dial("tcp", ":6379")
  }, 2)

  q := simpleq.New(pool, "my-simpleq")
}
```

Operations:

- `q.Push(el)` Returns number of elements in queue.
- `q.Pop()` and `q.BPop()` (blocking) Returns element popped or null
- `q.Pull(el)` Pull out a specific el (the highest/oldest el in the queue to be specific if elements are repeated) from the queue. Returns number of elements removed (0 or 1).
- `q1.PullPipe(q2, el)` Pull and push into another queue atomicly. Returns elements in second queue. (Note, that if el does not exist in q1, it will still be put into q2)
    - `q1.SPullPipe(q2, el)` is a safe version which will not insert el into q2 unless it has been successfully pulled out of q1. This is done atomically using a lua script.
- `q1.PopPipe(q2)` and `q1.BPopPipe(q2)` (blocking): Pop and push to another queue; returns popped element (also atomic).
- `q.Clear()` Clear out the queue
- `q.List()` List all elements in the queue

Listeners:

simpleq's can start listeners that run `BPopPipe` or `BPop` continuously and make callbacks when they occur. These two functions are `.PopListen` and `.PopPipeListen`.

_Note_: Calling listen will get another connection from the pool, allowing `.push` to still work because another connection is being blocked. Calling a listen function more than once will result in a panic, as this is not prudent.

_Note_: Unlike `node-simpleq`, the listener will not control the flow of messages. That is because the channel metaphor available in Go is a much better way to control flow using backpressure. `listener.Messages` and `listener.Errors` are buffer-less channels so by not pulling the next element off

Examples below:

```golang
listener := q.PopListen()

// or
listener := q.PopPipeListen(otherq);

// listen for messages and errors
go func () {
  for {
    select {
    case err, ok := <- listener.Errors:
      if ok {
        panic(err)
      }
    case msg, ok := <- listener.Messages:
      if ok {
        // Do something downstream with the message
        continue
      }
    }

    // Only get here if ok == false which mean's we've closed
    break
  }
  // Will get here on listener.End()
}()

// eventually
if err := listener.End(); err != nil {
  panic(err)
}
```

## Tests

```
go test github.com/Rafflecopter/golang-simpleq/simpleq
```

## License

See LICENSE file.

[1]: https://travis-ci.org/Rafflecopter/golang-simpleq.png?branch=master
[2]: http://travis-ci.org/Rafflecopter/golang-simpleq