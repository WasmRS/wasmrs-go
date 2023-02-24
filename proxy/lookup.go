package proxy

import (
	"sync"
)

type (
	Lookup struct {
		mu    sync.Mutex
		count int
		head  *node
		tail  *node
	}

	node struct {
		prev, next *node
		stream     Stream
	}
)

func (l *Lookup) Size() int {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.count
}

func (l *Lookup) IsEmpty() bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.count == 0
}

func (l *Lookup) Get(id uint32) (Stream, bool) {
	if n := l.find(id); n != nil {
		return n.stream, true
	}

	return nil, false
}

func (l *Lookup) Add(s Stream) {
	l.mu.Lock()
	defer l.mu.Unlock()

	newNode := &node{
		stream: s,
	}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		currentNode := l.head
		for currentNode.next != nil {
			currentNode = currentNode.next
		}
		newNode.prev = currentNode
		currentNode.next = newNode
		l.tail = newNode
	}
	l.count++
}

func (l *Lookup) Remove(id uint32) {
	l.mu.Lock()
	defer l.mu.Unlock()

	nodeToDelete := l.find(id)
	if nodeToDelete == nil {
		return
	}

	l.count--
	prevNode := nodeToDelete.prev
	nextNode := nodeToDelete.next

	if prevNode != nil {
		prevNode.next = nodeToDelete.next
	}
	if nextNode != nil {
		nextNode.prev = nodeToDelete.prev
	}
}

func (l *Lookup) find(id uint32) *node {
	for n := l.head; n != nil; n = n.next {
		if n.stream.StreamID() == id {
			return n
		}
	}
	return nil
}
