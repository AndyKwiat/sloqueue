package main

import "container/heap"

type EventQueue struct{
	queue priorityQueueBase
}

func NewEventQueue() EventQueue{
	q:= EventQueue{}
	heap.Init(&q.queue)
	return q
}

func (jq *EventQueue)Push( functionToRun func(), inHowManyTicks int){
	heap.Push(&jq.queue,
	&item{
		value:functionToRun,
		priority:currentTime + inHowManyTicks,
	})
}
func (jq *EventQueue)PeekPriority() int{
	return jq.queue.Peek().priority
}
func ( jq *EventQueue )Pop() func(){
	return heap.Pop(&jq.queue).(*item).value.(func())
}


