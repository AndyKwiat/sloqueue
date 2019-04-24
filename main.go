package main

import (
	"encoding/csv"
	"fmt"
	"github.com/wcharczuk/go-chart"
	"math/rand"
	"os"
	"sort"
	"strconv"
)


type Job struct{
	JobType string
	JobDuration int
	EnqueueTime int
	DequeueTime int
	CompletionTime int
	ID int
}


var currentTime int
var eventQueue EventQueue
var completedJobs JobQueue
var totalJobsInScenario int
var nextJobId int

func makeJob(jobType string, AvgDuration int ) *Job{
	nextJobId++
	return &Job{
		JobType:jobType,
		JobDuration:AvgDuration,
		ID:nextJobId,
	}
}


type jobSLOState struct {
	lastQueueTime int
	slo           int
	priority      int
	lastEnqueuePriority int
	priorityStats timeStats
	lastQTimeStats timeStats
	name string

	queue JobQueue
}

var sloState map[string]*jobSLOState
var orderedQueues []*JobQueue

func enqueueJob( job *Job ) {
	job.EnqueueTime = currentTime
	state := sloState[job.JobType]
	state.queue.Push(job, state.slo+currentTime)
}

func dequeueNextJob() *Job{

	for _,q:= range orderedQueues{
		if q.Length() <=0{
			continue
		}
		if q.PeekPriority() <=currentTime{
			return q.Pop()
		}
	}
	// no jobs to work right now
	var jobQToWork *JobQueue
	for _,q:= range orderedQueues{
		if q.Length() <=0{
			continue
		}
		if jobQToWork == nil {
			jobQToWork = q
			continue
		}
		if jobQToWork.PeekPriority() > q.PeekPriority(){
			jobQToWork = q
		}

		/*if jobQToWork.Length() < q.Length(){
			jobQToWork = q
		}*/
	}
	if jobQToWork != nil{
		return jobQToWork.Pop()
	}
	return nil
}


func pushJobsRate( q *JobQueue, jobType string, avgDuration,  jobRate, duration int ){
	jobCount := jobRate *duration / 1000
	totalJobsInScenario += jobCount
	log("scheduling ",jobCount,"jobs of type",jobType , " over ", duration," ticks" )
	for i:=0; i < jobCount; i++ {
		variance:= rand.Intn(avgDuration) - avgDuration/2
		job := makeJob(jobType,avgDuration + variance)
		timeToSched := i * duration / jobCount // spread out the jobs over the time period
		eventQueue.Push( func(){
			enqueueJob(job)

		},timeToSched )
	}
}

func pushJobsOnSchedule( jobType string ){
	s:= minuteEnqueueSchedule[jobType]
	for i:=0; i < len(s); i++ {
		pushJobsOverTime( i * 1000 * 60, 1000 * 60, jobType, s[i])
	}
}

func pushJobsOverTime( start int, duration int, jobType string, jobCount int ){
	avgDuration:= 1000
	totalJobsInScenario += jobCount
		for i:=0; i < jobCount; i++ {
			job := makeJob(jobType, avgDuration)
			timeToSched := i * duration / jobCount + start// spread out the jobs over the time period
			eventQueue.Push(func() {
				enqueueJob(job)

			}, timeToSched)
		}
}
func pushJobs( q *JobQueue, jobType string , avgDuration, priority, jobCount int){
	totalJobsInScenario += jobCount
	log("creating",jobCount,"jobs of type",jobType , " duration ",avgDuration)
	for i:=0; i < jobCount; i++ {
		job := makeJob(jobType,avgDuration)
		job.EnqueueTime = currentTime
		q.Push(job ,priority)
	}
	//log("new q length",q.Length())
}

type Worker struct{
	jobQueue *JobQueue
}

func ( w *Worker )WorkNewJob(){
	//log("workNewJob")
	j:= dequeueNextJob()

	if j== nil{
		eventQueue.Push(w.WorkNewJob,100)
		return
	}
	j.DequeueTime = currentTime
	qTime:= j.DequeueTime- j.EnqueueTime
	lq := sloState[j.JobType].lastQueueTime
	if lq <=0{
		lq = qTime
	}else{
		t:= 10
		lq = (qTime * t + lq *(1000-t))/1000
	}
	sloState[j.JobType].lastQueueTime= lq
	sloState[j.JobType].lastQTimeStats.add(lq,currentTime)
	eventQueue.Push( w.FinishJob(j) , j.JobDuration )
}
func ( w *Worker )FinishJob(job *Job)func(){
	return func(){
		//log("FinishJob",job.ID )
		job.CompletionTime = currentTime
		completedJobs.Push(job,currentTime)
		eventQueue.Push(w.WorkNewJob,1)
	}
}

func startWorkers( queue *JobQueue, workerCount int ){
	log("starting",workerCount,"workers")
	w:= &Worker{
		jobQueue:queue,
	}
	for i:=0; i < workerCount; i++{
		startWorker(w)
	}

}
func startWorker( worker *Worker ){

	eventQueue.Push( worker.WorkNewJob,1 )


}
var minuteEnqueueSchedule map[string][]int

func main(){
	// parse csv
	f,_ := os.Open("yeezy hour.csv")
	r:=csv.NewReader(f)
	records,err := r.ReadAll()
	if err != nil{
		panic(err)
	}
	minuteEnqueueSchedule = make(map[string][]int)
	headers := records[0]
	for i:=1; i< len(records); i++{
		for j:=1; j < len(records[i]); j++ {
			val,e := strconv.Atoi(records[i][j])
			if e!= nil{
				panic(e)
			}
			minuteEnqueueSchedule[headers[j]] = append( minuteEnqueueSchedule[headers[j]], val )
		}
	}



	// scenario
	eventQueue= NewEventQueue()
	currentTime = 0
	completedJobs = CreateQ()
	nextJobId = 0
	totalJobsInScenario = 0
	sloState=map[string]*jobSLOState	{
		//"realtime":{slo:1000},
		"payment":{slo:5000,name:"payment"},
		"webhook":{slo:31000,name:"webhook"},
		"default":{slo:30000,name:"default"},
		//"low":{slo:900000},
	//	"checkout_completion":{slo:5000},
	}

	// make sure queues are in slo order
	var orderedSloState []*jobSLOState
	for k,_ := range sloState{
		orderedSloState = append(orderedSloState, sloState[k])
	}
	sort.Slice(orderedSloState, func(i,j int)bool{ return orderedSloState[i].slo < orderedSloState[j].slo })

	for _,oss := range orderedSloState{
		orderedQueues = append(orderedQueues,&oss.queue)
	}




	jobQueue := CreateQ()

	startWorkers(&jobQueue,300)


	pushJobsRate( &jobQueue, "payment", 999, 70, 100000 )
	pushJobsRate( &jobQueue, "default", 1020, 150, 100000 )
	pushJobsRate( &jobQueue, "webhook", 1050, 300, 100000 )

	//pushJobsOnSchedule( "payment" )
	//pushJobsOnSchedule( "webhook" )
	//pushJobsOnSchedule( "default")
	//pushJobsOnSchedule( "low")
	//pushJobsOnSchedule( "checkout_completion")
	//pushJobsOnSchedule( "realtime")

	simulateUntilJobsdone()
	summary()
}


func summary(){
	log("completed jobs",completedJobs.Length())
	queueTimes := make(map [string]*timeStats)
	queueSizes := make(map[string]*timeStats)
	// create empty timeStats for each job type
	for k,_ := range sloState{
		queueTimes[k] = &timeStats{}
		queueSizes[k] = &timeStats{}
	}

	job:= completedJobs.Pop()
	for job != nil{
		queueTime := job.DequeueTime - job.EnqueueTime
		queueTimes[job.JobType].add(queueTime, job.DequeueTime)
		queueSizes[job.JobType].add(+1, job.EnqueueTime)
		queueSizes[job.JobType].add(-1,job.DequeueTime)

		job= completedJobs.Pop()
	}

	var qSizeChartData,qPriorities,qTimeSeries []chart.Series

	for k,v := range queueTimes {
		/*if k!= "payment"{
			continue
		}*/
		if v.count <=0{
			continue
		}
		log(k,"count",v.count,"avg",v.avg)

		// graph queuesizes
		x,y := queueSizes[k].orderedBucketsSum()
		// cumulative sum
		for i:=1; i < len(y); i++ {
			y[i] += y[i-1]
		}
		qSizeChartData = append(qSizeChartData,createSeries(x,y,k))
		x,y = sloState[k].priorityStats.orderedBucketsAvg()
		qPriorities = append ( qPriorities, createSeries(x,y,k))

		x,y = v.orderedBucketsAvg()
		qTimeSeries = append( qTimeSeries, createSeries(x,y,k))

	//	x,y = sloState[k].lastQTimeStats.orderedBucketsAvg()
//		qTimeSeries = append( qTimeSeries, createSeries(x,y,k+" avg"))

	}
	graph(qSizeChartData,"queueSizes")
//	graph(qPriorities,"queuePriorities")
	graph(qTimeSeries, "wait times")



}
func intsToFloats(data []int)[]float64{
	f:= make([]float64, len(data))
	for i:= range data{
		f[i] = float64(data[i])
	}
	return f
}
func createSeries(x,y []int, name string ) chart.Series{

	return chart.ContinuousSeries{
		Name:name,
		XValues: intsToFloats(x),
		YValues: intsToFloats(y),
	}
}
func simulateUntilJobsdone(){
	for totalJobsInScenario > completedJobs.Length() {
		simulateTick()
	}
}
func simulateUntil(endTime int ){
	prevTime:=0
	for currentTime < endTime{
		if currentTime - prevTime > 1000 {
			log("currentTime",currentTime)
			prevTime = currentTime
		}
		simulateTick()
	}
}
func simulateTick(){
	nextTick := eventQueue.PeekPriority()
	currentTime = nextTick
	for nextTick == currentTime{
		e:= eventQueue.Pop()
		e()
		nextTick= eventQueue.PeekPriority()

	}

}

func log(str... interface{}){
	fmt.Print("[",currentTime,"]  ")
	for _,s:= range str{
		fmt.Print(s," ")
	}
	fmt.Print("\n")
}