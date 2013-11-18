package simpleq

import (
	"github.com/garyburd/redigo/redis"
)

// A listener on a queue, repeatedly calling BPop or BPopPipe
type Listener struct {
	conn       redis.Conn
	q, pipeto  *SimpleQ
	end, ended chan bool
	Elements   chan []byte
	Errors     chan error
}

// Create a new listener. Use pipeto as nil to call BPop.
func NewListener(conn redis.Conn, q, pipeto *SimpleQ) *Listener {
	l := &Listener{
		conn:     conn,
		q:        q,
		pipeto:   pipeto,
		end:      make(chan bool),
		ended:    make(chan bool),
		Elements: make(chan []byte),
		Errors:   make(chan error),
	}

	go l.listen()

	return l
}

func (l *Listener) End() error {
	select {
	case <-l.ended:
		// Already ended
	default:
		close(l.end)
		<-l.ended
	}

	return nil
}

func (l *Listener) listen() {

ListenLoop:
	for {
		select {
		case <-l.end:
			break ListenLoop // On end, we should stop listening
		default:
		}

		if el, err := l.call(); err != nil {
			l.Errors <- err
		} else if el != nil {
			l.Elements <- el
		}
	}
	l.close()
}

func (l *Listener) close() {
	close(l.ended)
	l.conn.Close()
	close(l.Errors)
	close(l.Elements)
}

func (l *Listener) call() ([]byte, error) {
	if l.pipeto != nil {
		res, err := l.conn.Do("BRPOPLPUSH", l.q.key, l.pipeto.key, 1)

		if bres, ok := res.([]byte); ok {
			return bres, err
		}
		return nil, err
	} else {
		res, err := l.conn.Do("BRPOP", l.q.key, 1)

		if ares, ok := res.([]interface{}); ok && len(ares) == 2 {
			if bres, ok := ares[1].([]byte); ok {
				return bres, err
			}
		}
		return nil, err
	}
}
