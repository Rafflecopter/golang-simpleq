package simpleq

import (
	"github.com/garyburd/redigo/redis"
	"github.com/yanatan16/errorcaller"
)

// A listener on a queue, repeatedly calling BPop or BPopPipe
type Listener struct {
	conn       redis.Conn
	qkey       string
	pipeto     *Queue
	end, ended chan bool
	Elements   chan []byte
	Errors     chan error
}

// Create a new listener. Use pipeto as nil to call BPop.
func NewListener(q, pipeto *Queue) *Listener {
	l := &Listener{
		conn:     q.pool.Get(),
		qkey:     q.key,
		pipeto:   pipeto,
		end:      make(chan bool),
		ended:    make(chan bool),
		Elements: make(chan []byte),
		Errors:   make(chan error),
	}

	go l.listen()

	return l
}

func (l *Listener) Close() error {
	select {
	case <-l.end:
		// Already ending
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
			l.Errors <- errorcaller.Err(err)
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
		res, err := l.conn.Do("BRPOPLPUSH", l.qkey, l.pipeto.key, 1)

		if bres, ok := res.([]byte); ok {
			return bres, err
		}
		return nil, err
	} else {
		res, err := l.conn.Do("BRPOP", l.qkey, 1)

		if ares, ok := res.([]interface{}); ok && len(ares) == 2 {
			if bres, ok := ares[1].([]byte); ok {
				return bres, err
			}
		}
		return nil, err
	}
}
