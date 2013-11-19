// Package simpleq provides a super-simple queue backed by Redis
package simpleq

import (
	"github.com/Rafflecopter/golang-simpleq/scripts"
	"github.com/garyburd/redigo/redis"
)

// A super simple redis-backed queue
type SimpleQ struct {
	conn     redis.Conn
	key      string
	listener *Listener
}

// Create a simpleq
func New(conn redis.Conn, key string) *SimpleQ {
	return &SimpleQ{
		conn: conn,
		key:  key,
	}
}

// End this queue
func (q *SimpleQ) Close() error {
	err := q.conn.Close()
	if err != nil {
		return err
	}
	if q.listener != nil {
		return q.listener.Close()
	}
	return nil
}

// Push an element onto the queue
func (q *SimpleQ) Push(el []byte) (length int64, err error) {
	return redis.Int64(q.conn.Do("LPUSH", q.key, el))
}

// Pop an element off the queue
func (q *SimpleQ) Pop() (el []byte, err error) {
	return redis.Bytes(q.conn.Do("RPOP", q.key))
}

// Block and Pop an element off the queue
// Use timeout_secs = 0 to block indefinitely
// On timeout, this DOES return an error because redigo does.
func (q *SimpleQ) BPop(timeout_secs int) (el []byte, err error) {
	res, err := redis.Values(q.conn.Do("BRPOP", q.key, timeout_secs))

	if len(res) == 2 {
		if bres, ok := res[1].([]byte); ok {
			return bres, err
		}
	}
	return nil, err
}

// Pull an element out the queue (oldest if more than one)
func (q *SimpleQ) Pull(el []byte) (nRemoved int64, err error) {
	return redis.Int64(q.conn.Do("LREM", q.key, -1, el))
}

// Pull an element out of the queue and push it onto another atomically
// Note: This will push the element regardless of the return value from pull
func (q *SimpleQ) PullPipe(q2 *SimpleQ, el []byte) (lengthQ2 int64, err error) {
	c := q.conn

	c.Send("MULTI")
	c.Send("LREM", q.key, -1, el)
	c.Send("LPUSH", q2.key, el)
	res, err := redis.Values(c.Do("EXEC"))

	if len(res) == 2 {
		if ires, ok := res[1].(int64); ok {
			return ires, err
		}
	}

	return 0, err
}

// Safely pull an element out of the queue and push it onto another atomically
// Returns 0 for non-existance in first queue, or length of second queue
func (q *SimpleQ) SPullPipe(q2 *SimpleQ, el []byte) (result int64, err error) {
	return redis.Int64(scripts.SafePullPipe.Do(q.conn, q.key, q2.key, el))
}

// Pop an element out of a queue and put it in another queue atomically
func (q *SimpleQ) PopPipe(q2 *SimpleQ) (el []byte, err error) {
	return redis.Bytes(q.conn.Do("RPOPLPUSH", q.key, q2.key))
}

// Block and Pop an element out of a queue and put it in another queue atomically
// On timeout, this doesn't return an error because redigo doesn't.
func (q *SimpleQ) BPopPipe(q2 *SimpleQ, timeout_secs int) (el []byte, err error) {
	return redis.Bytes(q.conn.Do("BRPOPLPUSH", q.key, q2.key, timeout_secs))
}

// Clear the queue of elements
func (q *SimpleQ) Clear() (nRemoved int64, err error) {
	return redis.Int64(q.conn.Do("DEL", q.key))
}

// List the elements in the queue
func (q *SimpleQ) List() (elements [][]byte, err error) {
	res, err := redis.Values(q.conn.Do("LRANGE", q.key, 0, -1))
	if err != nil {
		return nil, err
	}

	elements = make([][]byte, len(res))
	for i, el := range res {
		elements[i], _ = el.([]byte)
	}
	return elements, nil
}

// Create a listener that calls Pop
func (q *SimpleQ) PopListen(conn redis.Conn) *Listener {
	return q.PopPipeListen(conn, nil)
}

// Create a listener that calls PopPipe
func (q *SimpleQ) PopPipeListen(conn redis.Conn, q2 *SimpleQ) *Listener {
	if q.listener != nil {
		panic("SimpleQ can only have one listener")
	}

	q.listener = NewListener(conn, q, q2)

	go func() {
		<-q.listener.ended
		q.listener = nil
	}()

	return q.listener
}
