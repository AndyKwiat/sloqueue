
package main

import (
	"container/heap"
)


// An item is something we manage in a priority queue.
type item struct {
	value    interface{} // The value of the item; arbitrary.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A priorityQueueBase implements heap.Interface and holds Items.
type priorityQueueBase []*item

func (pq priorityQueueBase) Len() int { return len(pq) }

func (pq priorityQueueBase) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority < pq[j].priority
}

func (pq priorityQueueBase) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *priorityQueueBase) Push(x interface{}) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)

}
func ( pq *priorityQueueBase) Peek() *item{

	old := *pq
	return old[0]

}
func (pq *priorityQueueBase) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an item in the queue.
func (pq *priorityQueueBase) update(item *item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}