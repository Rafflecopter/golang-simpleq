package simpleq

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

var _ = fmt.Println

func onerror(t *testing.T, l *Listener) {
	go func() {
		for err := range l.Errors {
			t.Error(err)
		}
	}()
}

func extraElements(t *testing.T, l *Listener) {
	go func() {
		for el := range l.Elements {
			t.Error("Extra element:", string(el))
		}
	}()
}

func TestBasicPopListen(t *testing.T) {
	q := begin()
	defer end(t, q)
	l := q.PopListen()
	onerror(t, l)

	go func() {
		epush(t, q, "hello")
	}()

	select {
	case el := <-l.Elements:
		if !reflect.DeepEqual(el, b("hello")) {
			t.Error("el isn't hello?", string(el))
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Timeout!")
	}

	extraElements(t, l)
	if err := l.End(); err != nil {
		t.Error(err)
	}

	select {
	case _, ok := <-l.Errors:
		if ok {
			t.Error("Errors not closed?")
		}
	default:
		t.Error("Errors not closed?")
	}
	select {
	case _, ok := <-l.Elements:
		if ok {
			t.Error("Elements not closed?")
		}
	default:
		t.Error("Elements not closed?")
	}
}

func TestBasicPopPipe(t *testing.T) {
  q, q2 := begin2()
  defer end(t, q, q2)
  l := q.PopPipeListen(q2)
  onerror(t, l)

  go func() {
    epush(t, q, "world")
  }()

  select {
  case el := <-l.Elements:
    if !reflect.DeepEqual(el, b("world")) {
      t.Error("el isn't world?", string(el))
    }
  case <-time.After(50 * time.Millisecond):
    t.Error("Timeout!")
  }

  extraElements(t, l)
  checkList(t, q2, b("world"))
  checkList(t, q)
}

func TestListenerEndTwice(t *testing.T) {
  q := begin()
  defer end(t, q)
  l := q.PopListen()
  onerror(t,l)
  extraElements(t, l)

  if err := l.End(); err != nil {
    t.Error(err)
  }
  if q.listener != nil {
    t.Error("q.listener isn't nil yet")
  }
  if err := l.End(); err != nil {
    t.Error(err)
  }
}