package main

import "sort"

type timeStats struct{
	count int
	total int
	avg int
	values []int
	times []int
}
func (s*timeStats)orderedBucketsAvg() (xData, yData []int){
	if len(s.times) != len(s.values){
		panic("data problem")
	}
	counts:= make(map[int]int)
	totals:= make(map[int]int)
	for i:=0; i < len(s.times); i++ {
		x := s.times[i]
		counts[x]++
		totals[x]+=s.values[i]
	}
	for k,_ := range counts{
		xData = append(xData, k)
	}
	sort.Ints(xData)
	for i:=0; i < len(xData); i++{
		x := xData[i]
		avg:= totals[x] / counts[x]
		yData = append(yData,avg)
	}
	return
}

func (s*timeStats)orderedBucketsSum() (xData, yData []int){
	if len(s.times) != len(s.values){
		panic("data problem")
	}
	counts:= make(map[int]int)
	totals:= make(map[int]int)
	for i:=0; i < len(s.times); i++ {
		x := s.times[i]
		counts[x]++
		totals[x]+=s.values[i]
	}
	for k,_ := range counts{
		xData = append(xData, k)
	}
	sort.Ints(xData)
	for i:=0; i < len(xData); i++{
		x := xData[i]
		yData = append(yData,totals[x])
	}
	return
}

func (s *timeStats)add(value int, tickTime int ){
	s.count ++
	s.total += value
	s.avg = s.total / s.count

	s.values = append(s.values,value)
	s.times = append(s.times, tickTime)
}


