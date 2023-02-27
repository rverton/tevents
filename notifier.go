package tevents

import (
	"fmt"
	"sync"
)

type Notifier struct {
	InputChan chan string
	count     int64

	mu       sync.Locker
	listener map[int64]Listener
}

func NewNotifier() *Notifier {
	return &Notifier{
		listener: make(map[int64]Listener, 0),
		mu:       &sync.Mutex{},
	}
}

type Listener struct {
	id int64
	c  chan Event
}

func (b *Notifier) AddListener() Listener {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.count
	b.count++

	l := Listener{id: id, c: make(chan Event)}
	b.listener[id] = l
	fmt.Printf("%+v\n", b.listener)
	return l
}

func (b *Notifier) RemoveListener(l Listener) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.listener, l.id)
}

func (b *Notifier) Send(e Event) {
	fmt.Printf("sending event to %v listeners\n", len(b.listener))
	for _, l := range b.listener {
		select {
		case l.c <- e:
		default:
			// ignore failed send
			fmt.Println("failed to send event")
		}
	}
}
