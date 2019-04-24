# SLO-based queues
For hackdays I created a queue simulator and experimented with different implementations 
of queues based on [aging](https://www.geeksforgeeks.org/starvation-aging-operating-systems/). Although
I tried a bunch of different solutions, this document outlines what I think is the most promising one.

## The Problem
Right now we have a lot of different queues with different priorities as Kir explained in his
[JobsDB Project](https://github.com/kirs/jobsdb). Tuning the performance
characteristics of each queue is not trivial since there are multiple types of job workers each working on multiple
job queues. 
 
 What we really want is to just tell the queue system a particular job's SLO of how long it is acceptable
 for it to wait in the queue. These could be similar to our current Jobs SLOs we monitor, for example:
  
  | Job Type | SLO |
  |:-----:|:-----:|
  | payment | 5 sec|
  | default | 30 sec |
  | webhook | 5 min |
  
  ### Algorithm (SLO based queues)
  
  * each SLO value (5 sec, 30 sec, 5 min sec etc.) has it's own queue.
  * when you enqueue a job, also enqueue it's expected time of dequeue. *Expected Dequeue Time = Current Time + SLO*
  * loop through the queues from shortest SLO to longest SLO  
  * if the job at the front of the queue is *on or behind* schedule, dequeue it
  * if the job at the front of the queue is *ahead* of schedule, don't dequeue it, and go on to the next slowest queue
  * all workers call the same dequeue operation (i.e. no special-purpose workers)
  
  Here is what the algorithm looks like in the code: 
  ```
  func dequeueNextJob() *Job{
    	for _,q:= range orderedQueues{          // queues ordered from shortest SLO, to longest
  		if q.Length() <=0{
  			continue
  		}
  		if q.PeekPriority() <=currentTime{     // PeekPriority() returns expected dequeue time
  			return q.Pop()
  		}
  	}
  	
  	// handle case where everyone is ahead of schedule ...
  	// ...
  }
  ```
  ## Properties
  * under normal load every queue will get worked off faster than it's SLO
  * under high load, the low-priority queues (eg. `webhook`) will get sacrificed so that 
  high-priority queues (eg. `payment` can meet their SLOs)
  * the high-priority queues will throttle themselves not to starve the low-priority queues  
  
  Example of WAIT TIMES for a high-load event:  
![alt text](./wait%20times.png)

SLOs: `payment` = 5000ms, `default`= 30000ms, `webhook`=300000ms

  In the above example, payment wait times go to about 5 seconds, `default` go to about 30 seconds, and `webhook`
  is sacrificed until the high load ends.
     
  ## Benefits
  * each job can just specify its acceptable wait time, no more special queues with ambiguous priorities
  * uniform job workers: no more specialized worker deployments (payment, default,billing etc)
  * can be implemented as part of some future *Jobs V2*, or just on top of what we currently have
  * easier to tune: right now we have a complex system of [worker-types and priorities](https://github.com/Shopify/shopify/blob/master/config/k8s/jobs.yml.erb#L9), this will
  get rid of all of that
  * easy to implement autoscaling: if the lowest-priority queue is behind SLO, we know we need more job workers
