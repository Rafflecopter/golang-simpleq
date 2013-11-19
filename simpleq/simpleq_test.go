package simpleq

import (
	"github.com/garyburd/redigo/redis"
	"math/rand"
	"reflect"
	"testing"
	"time"
	"io"
)

var pool *redis.Pool

func init() {
	rand.Seed(time.Now().Unix())
	pool = redis.NewPool(func() (redis.Conn, error) {
		return redis.Dial("tcp", ":6379")
	}, 10)
}


// -- Tests --

func TestPush(t *testing.T) {
	q := begin()
	defer end(t, q)

	if n, err := q.Push(b("foo123")); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Error("n != 1", n)
	}
	if n, err := q.Push(b("456bar")); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Error("n != 2", n)
	}

	checkList(t, q, b("456bar"), b("foo123"))
}

func TestPop(t *testing.T) {
	q := begin()
	defer end(t, q)

	if el, err := q.Pop(); err == nil {
		t.Error("No pop error on nil?")
	} else if el != nil {
		t.Error("Pop() returned element when it shouldn't have!")
	}

	epush(t, q, "father")
	epush(t, q, "mother")
	epush(t, q, "baby")

	if el, err := q.Pop(); err != nil {
		t.Error("Error Pop():", err)
	} else if !reflect.DeepEqual(el, b("father")) {
		t.Error("Element isn't father?", el)
	}
}

func TestBPop(t *testing.T) {
	q := begin()
	qc := clone(q)
	defer end(t, q, qc)

	go func() {
		time.Sleep(10 * time.Millisecond)
		epush(t, qc, "urukai")
	}()

	if el, err := q.BPop(1); err != nil {
		t.Error("Error BPop():", err)
	} else if !reflect.DeepEqual(el, b("urukai")) {
		t.Error("Element isn't urukai?", string(el))
	}
}

func TestBPopTimeout(t *testing.T) {
	q := begin()
	defer end(t, q)

	now := time.Now()

	if el, _ := q.BPop(1); el != nil {
		t.Error("Element isn't nil", el)
	}

	if float64(time.Now().Sub(now)) < float64(time.Second) * .9 {
		t.Error("Timeout didn't last a second!", time.Now().Sub(now))
	}
}

func TestPull(t *testing.T) {
	q := begin()
	defer end(t, q)

	epush(t, q, "darth")
	epush(t, q, "vader")
	epush(t, q, "mothe-vader")

	if n, err := q.Pull(b("mothe-vader")); err != nil {
		t.Error("Pull():", err)
	} else if n != 1 {
		t.Error("n != 1", n)
	}
	if n, err := q.Pull(b("darth")); err != nil {
		t.Error("Pull():", err)
	} else if n != 1 {
		t.Error("n != 1", n)
	}
	if n, err := q.Pull(b("kefka")); err != nil {
		t.Error("Pull():", err)
	} else if n != 0 {
		t.Error("n != 0", n)
	}

	checkList(t, q, b("vader"))
}

func TestPopPipe(t *testing.T) {
	q, q2 := begin2()
	defer end(t, q, q2)

	epush(t, q, "luke")
	epush(t, q, "skywalker")
	epush(t, q, "groundcrawler")

	if el, err := q.PopPipe(q2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(el, b("luke")) {
		t.Error("not luke?", string(el))
	}

	if el, err := q.PopPipe(q2); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(el, b("skywalker")) {
		t.Error("not skywalker?", string(el))
	}

	checkList(t, q, b("groundcrawler"))
	checkList(t, q2, b("skywalker"), b("luke"))
}

func TestBPopPipe(t *testing.T) {
	q, q2 := begin2()
	qc := clone(q)
	defer end(t, q, q2, qc)

	go func() {
		time.Sleep(10 * time.Millisecond)
		epush(t, qc, "jack")
	}()

	if el, err := q.BPopPipe(q2, 1); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(el, b("jack")) {
		t.Error("not jack?", string(el))
	}

	epush(t, q, "sparrow")
	epush(t, q, "lives")

	if el, err := q.BPopPipe(q2, 1); err != nil {
		t.Error(err)
	} else if !reflect.DeepEqual(el, b("sparrow")) {
		t.Error("not sparrow?", string(el))
	}

	checkList(t, q, b("lives"))
	checkList(t, q2, b("sparrow"), b("jack"))
}

func TestBPopPipeTimeout(t *testing.T) {
	q, q2 := begin2()
	defer end(t, q, q2)

	now := time.Now()

	if el, _ := q.BPopPipe(q2, 2); el != nil {
		t.Error("not nil?", el)
	}

	if float64(time.Now().Sub(now)) < float64(time.Second) * .9 {
		t.Error("Timeout didn't last a second!", time.Now().Sub(now))
	}
}

func TestPullPipe(t *testing.T) {
	q, q2 := begin2()
	defer end(t, q, q2)

	epush(t, q, "gimli")
	epush(t, q, "legolas")
	epush(t, q, "aragorn")

	if n, err := q.PullPipe(q2, b("gimli")); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Error("n != 1", n)
	}
	if n, err := q.PullPipe(q2, b("aragorn")); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Error("n != 2", n)
	}
	if n, err := q.PullPipe(q2, b("frodo")); err != nil {
		t.Error(err)
	} else if n != 3 {
		t.Error("n != 3", n)
	}

	checkList(t, q, b("legolas"))
	checkList(t, q2, b("frodo"), b("aragorn"), b("gimli"))
}

func TestSPullPipe(t *testing.T) {
	q, q2 := begin2()
	defer end(t, q, q2)

	epush(t, q, "tobias")
	epush(t, q, "gob")
	epush(t, q, "maybe")

	if n, err := q.SPullPipe(q2, b("gob")); err != nil {
		t.Error(err)
	} else if n != 1 {
		t.Error("n != 1", n)
	}
	if n, err := q.SPullPipe(q2, b("george-michael")); err != nil {
		t.Error(err)
	} else if n != 0 {
		t.Error("n != 0", n)
	}
	if n, err := q.SPullPipe(q2, b("maybe")); err != nil {
		t.Error(err)
	} else if n != 2 {
		t.Error("n != 2", n)
	}

	checkList(t, q, b("tobias"))
	checkList(t, q2, b("maybe"), b("gob"))
}

// -- Helpers --

func randKey() string {
	return "go-simpleq-test:" + rstr(8)
}

func rstr(n int) string {
	s := make([]byte, 8)
	for i := 0; i < n; i++ {
		s[i] = byte(rand.Int()%26 + 97)
	}
	return string(s)
}

// Just easier syntax
func b(s string) []byte {
	return []byte(s)
}

func checkList(t *testing.T, q *SimpleQ, els ...[]byte) {
	list, err := q.List()
	if err != nil {
		t.Error("Error List(): " + err.Error())
	}
  if len(list) == 0 && len(els) == 0 {
    return
  }
	if !reflect.DeepEqual(list, els) {
		t.Error("List isn't as it should be:", strlist(list), strlist(els))
	}
}

func strlist(b [][]byte) []string {
	s := make([]string, len(b))
	for i, x := range b {
		s[i] = string(x)
	}
	return s
}

func begin() *SimpleQ {
	q := New(pool.Get(), randKey())
	q.Clear()
	return q
}
func begin2() (*SimpleQ, *SimpleQ) {
	return begin(), begin()
}

func clone(q *SimpleQ) *SimpleQ {
	return New(pool.Get(), q.key)
}

func end(t *testing.T, qs ...io.Closer) {
	for _, q := range qs {
		if err := q.Close(); err != nil {
			t.Error(err)
		}
	}
}

func epush(t *testing.T, q *SimpleQ, el string) {
	if _, err := q.Push(b(el)); err != nil {
		t.Error("Error Push(", el, "): ", err)
	}
}