package main

import (
	"errors"
	"sync"
)

//Node storage of queue data
type Node struct {
	data interface{}
	prev *Node
	next *Node
}

//QueueBackend Backend storage of the queue, a double linked list
type QueueBackend struct {
	//Pointers to root and end
	head *Node
	tail *Node

	//keep track of current size
	size    uint32
	maxSize uint32
}

func (queue *QueueBackend) createNode(data interface{}) *Node {
	node := Node{}
	node.data = data
	node.next = nil
	node.prev = nil

	return &node
}

func (queue *QueueBackend) put(data interface{}) error {
	if queue.size >= queue.maxSize {
		err := errors.New("Queue full")
		return err
	}

	if queue.size == 0 {
		//new root node
		node := queue.createNode(data)
		queue.head = node
		queue.tail = node

		queue.size++

		return nil
	}

	//queue non-empty append to head
	currentHead := queue.head
	newHead := queue.createNode(data)
	newHead.next = currentHead
	currentHead.prev = newHead

	queue.head = currentHead
	queue.size++
	return nil
}

func (queue *QueueBackend) pop() (interface{}, error) {
	if queue.size == 0 {
		err := errors.New("Queue empty")
		return nil, err
	}

	currentEnd := queue.tail
	newEnd := currentEnd.prev

	if newEnd != nil {
		newEnd.next = nil
	}

	queue.size--
	if queue.size == 0 {
		queue.head = nil
		queue.tail = nil
	}

	return currentEnd.data, nil
}

func (queue *QueueBackend) isEmpty() bool {
	return queue.size == 0
}

func (queue *QueueBackend) isFull() bool {
	return queue.size >= queue.maxSize
}

//ConcurrentQueue concurrent queue
type ConcurrentQueue struct {
	//mutex lock
	lock *sync.Mutex

	//empty and full locks
	notEmpty *sync.Cond
	notFull  *sync.Cond

	//queue storage backend
	backend *QueueBackend
}

func (c *ConcurrentQueue) enqueue(data interface{}) error {
	c.lock.Lock()

	for c.backend.isFull() {
		//wait for empty
		c.notFull.Wait()
	}

	//insert
	err := c.backend.put(data)

	//signal notEmpty
	c.notEmpty.Signal()

	c.lock.Unlock()

	return err
}

func (c *ConcurrentQueue) dequeue() (interface{}, error) {
	c.lock.Lock()

	for c.backend.isEmpty() {
		c.notEmpty.Wait()
	}

	data, err := c.backend.pop()

	//signal notFull
	c.notFull.Signal()

	c.lock.Unlock()

	return data, err
}

func (c *ConcurrentQueue) getSize() uint32 {
	c.lock.Lock()
	size := c.backend.size
	c.lock.Unlock()

	return size
}

func (c *ConcurrentQueue) getMaxSize() uint32 {
	c.lock.Lock()
	maxSize := c.backend.maxSize
	c.lock.Unlock()

	return maxSize
}

//NewConcurrentQueue Creates a new queue
func NewConcurrentQueue(maxSize uint32) *ConcurrentQueue {
	queue := ConcurrentQueue{}

	//init mutexes
	queue.lock = &sync.Mutex{}
	queue.notFull = sync.NewCond(queue.lock)
	queue.notEmpty = sync.NewCond(queue.lock)

	//init backend
	queue.backend = &QueueBackend{}
	queue.backend.size = 0
	queue.backend.head = nil
	queue.backend.tail = nil

	queue.backend.maxSize = maxSize
	return &queue
}
