package main

import (
	"container/heap"
)

type QueryQueue struct{
	queue priorityQueueBase
}

func CreateQ() QueryQueue {
	jq:= QueryQueue{}
	heap.Init(&jq.queue)
	return jq
}

func (jq *QueryQueue)Push(job *Query,priority int){
	heap.Push(&jq.queue,
		&item{
			value:job,
			priority:priority,
		})
}

func (jq *QueryQueue)Pop()*Query {
	if jq.Length() <=0 {
		return nil
	}
	return heap.Pop(&jq.queue).(*item).value.(*Query)
}

func (jq *QueryQueue)Length()int{
	return jq.queue.Len()
}

func (jq *QueryQueue)PeekPriority() int{
	return jq.queue.Peek().priority
}

