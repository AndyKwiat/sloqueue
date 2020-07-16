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
var utilizationStats = timeStats{}
var waitStats = map[string]*timeStats{ "admin":{}, "nonadmin":{}}

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
	scheduleQueriesOverTime(0,1000, "nonadmin",1000)

	simulateUntilJobsdone()
	summary()
}


func summary(){
	log("completed jobs", len(completedQueries))


	x,y := utilizationStats.orderedBucketsAvg()
	graph( []chart.Series{ createSeries(x,y,"utilization")},"Utilization" )

	var waitStatSeries []chart.Series
	for key,stats:= range waitStats{
		x,y:= stats.orderedBucketsAvg()
		waitStatSeries = append(waitStatSeries, createSeries(x,y,key))
	}
	graph(waitStatSeries,"queryWaitTimes")


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

const THREAD_COUNT = 50
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
			waitStats[q.QueryType].add(q.CompletionTime - q.EnqueueTime, currentTime)
			log("Finished query", q.ID)
		}else{
			newWorkQueue = append(newWorkQueue, q)
		}
	}
	dbCurrentlyWorkingOn = newWorkQueue

	utilization:= (len(dbWorkQueue) + len(dbCurrentlyWorkingOn))*100 / THREAD_COUNT
	utilizationStats.add(utilization,currentTime)
	currentTime++

}

func log(str... interface{}){
	fmt.Print("[",currentTime,"]  ")
	for _,s:= range str{
		fmt.Print(s," ")
	}
	fmt.Print("\n")
}