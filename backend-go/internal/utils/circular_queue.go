package utils



type IntQueue struct {
	data     []int64
	head     int
	tail     int
	size     int
	capacity int
}


func NewIntQueue(capacity int) *IntQueue {
	return &IntQueue{
		data:     make([]int64, capacity),
		capacity: capacity,
	}
}

func (q *IntQueue) Enqueue(val int64) {
	if q.size == q.capacity {
		panic("queue is full")
	}
	q.data[q.tail] = val
	q.tail = (q.tail + 1) % q.capacity
	q.size++
}

func (q *IntQueue) Dequeue() int64 {
	if q.size == 0 {
		panic("queue is empty")
	}
	val := q.data[q.head]
	q.head = (q.head + 1) % q.capacity
	q.size--
	return val
}

func (q *IntQueue) IsEmpty() bool {
	return q.size == 0
}
