package main

import (
	"container/heap"
)

type JobQueue struct{
	queue priorityQueueBase
}

func CreateQ() JobQueue {
	jq:= JobQueue{}
	heap.Init(&jq.queue)
	return jq
}

func (jq *JobQueue)Push(job *Job,priority int){
	heap.Push(&jq.queue,
		&item{
			value:job,
			priority:priority,
		})
}

func (jq *JobQueue)Pop()*Job{
	if jq.Length() <=0 {
		return nil
	}
	return heap.Pop(&jq.queue).(*item).value.(*Job)
}

func (jq *JobQueue)Length()int{
	return jq.queue.Len()
}

func (jq *JobQueue)PeekPriority() int{
	return jq.queue.Peek().priority
}

