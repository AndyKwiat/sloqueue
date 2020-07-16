package main

import (
	"fmt"
	"github.com/wcharczuk/go-chart"
)


type Query struct{
	QueryType      string
	QueryDuration  int

	WaitingDuration int
	EnqueueTime    int
	DequeueTime    int
	CompletionTime int

	ID             int
}

var currentTime int
var eventQueue []*Query
var dbWorkQueue []*Query
var dbCurrentlyWorkingOn []*Query
var completedQueries []*Query
var totalQueriesInScenario int
var nextQueryId int

func makeQuery(queryType string, AvgDuration int ) *Query {
	nextQueryId++
	return &Query{
		QueryType:     queryType,
		QueryDuration: AvgDuration,
		ID:            nextQueryId,
	}
}

func enqueueQuery( query *Query) {
	query.EnqueueTime = currentTime
	dbWorkQueue = append(dbWorkQueue, query )
}

func dequeueNextQuery() *Query {
	var val *Query
	val, dbWorkQueue = dbWorkQueue[0], dbWorkQueue[1:]
	return val

}


func scheduleQueriesOverTime( start int, duration int, jobType string, jobCount int ){
	avgDuration:= 100
	totalQueriesInScenario += jobCount
		for i:=0; i < jobCount; i++ {
			job := makeQuery(jobType, avgDuration)
			timeToSched := i * duration / jobCount + start// spread out the jobs over the time period
			job.EnqueueTime = timeToSched
			eventQueue = append(eventQueue, job )

		}
}


func main(){
	// scenario
	currentTime = 0
	nextQueryId = 0
	totalQueriesInScenario = 0

//	jobQueue := CreateQ()
	scheduleQueriesOverTime(0,1, "checkout",10)

	simulateUntilJobsdone()
	summary()
}


func summary(){
	log("completed jobs", len(completedQueries))
	/*queueTimes := make(map [string]*timeStats)
	queueSizes := make(map[string]*timeStats)
	jobsWorkedCount := make(map[string]*timeStats)
	// create empty timeStats for each job type
	k:= "default"
		queueTimes[k] = &timeStats{}
		queueSizes[k] = &timeStats{}
		jobsWorkedCount[k] = &timeStats{}


	job:= completedQueries.Pop()
	for job != nil{
		queueTime := job.DequeueTime - job.EnqueueTime
		queueTimes[job.QueryType].add(queueTime, job.DequeueTime)
		queueSizes[job.QueryType].add(+1, job.EnqueueTime)
		queueSizes[job.QueryType].add(-1,job.DequeueTime)
		jobsWorkedCount[job.QueryType].add(1, job.DequeueTime)

		job= completedQueries.Pop()
	}

	var qSizeChartData,qCounts,qTimeSeries []chart.Series

		v:= queueTimes[k]


		log(k,"count",v.count,"avg",v.avg)

		// graph queuesizes
		x,y := queueSizes[k].orderedBucketsSum()
		// cumulative sum
		for i:=1; i < len(y); i++ {
			y[i] += y[i-1]
		}
		qSizeChartData = append(qSizeChartData,createSeries(x,y,k))


		x,y = jobsWorkedCount[k].orderedBucketsSum()
		// cumulative sum
		for i:=1; i < len(y); i++ {
			y[i] += y[i-1]
		}


		qCounts = append( qCounts, createSeries(x,y,k))

		x,y = v.orderedBucketsAvg()
		qTimeSeries = append( qTimeSeries, createSeries(x,y,k))

	//	x,y = sloState[k].queueTimeStats.orderedBucketsAvg()
//		qTimeSeries = append( qTimeSeries, createSeries(x,y,k+" avg"))


	graph(qSizeChartData,"queueSizes")
	graph(qCounts,"jobsWorked")
	graph(qTimeSeries, "wait times")

*/

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
	for totalQueriesInScenario > len(completedQueries) {
		simulateTick()
	}
}

const THREAD_COUNT = 3
func simulateTick(){
	// add scheduled queries
	var lastIndex int
	for i,q := range eventQueue{
		if q.EnqueueTime > currentTime {
			// this query is not scheduled to be added yet
			break
		}

		q.EnqueueTime = currentTime
		dbWorkQueue = append(dbWorkQueue, q )
		lastIndex = i+1
		log("querySentToDB", q.ID , "duration",q.QueryDuration)
	}
	eventQueue = eventQueue[lastIndex:]

	// start working on queued queries
	lastIndex = 0
	for i,q := range dbWorkQueue{
		if len(dbCurrentlyWorkingOn) >= THREAD_COUNT {
			break // too busy right now
		}

		q.DequeueTime = currentTime
		q.WaitingDuration = q.DequeueTime - q.EnqueueTime
		dbCurrentlyWorkingOn = append(dbCurrentlyWorkingOn,q)
		lastIndex = i+1
		log("addingQueryToWorkOn", q.ID , "duration",q.QueryDuration)
	}
	dbWorkQueue = dbWorkQueue[lastIndex:]

	// work off existing queries
	var newWorkQueue []*Query
	for _,q := range dbCurrentlyWorkingOn{
		if q.DequeueTime + q.QueryDuration <= currentTime {
			completedQueries = append(completedQueries, q)
			q.CompletionTime = currentTime
			log("Finished query", q.ID)
		}else{
			newWorkQueue = append(newWorkQueue, q)
		}
	}
	dbCurrentlyWorkingOn = newWorkQueue


	currentTime++
}

func log(str... interface{}){
	fmt.Print("[",currentTime,"]  ")
	for _,s:= range str{
		fmt.Print(s," ")
	}
	fmt.Print("\n")
}