package main

import (
	"bytes"
	"fmt"
	"github.com/wcharczuk/go-chart"
	"io/ioutil"
)

func graph(data []chart.Series,fileName string) {
	graph := chart.Chart{
		XAxis: chart.XAxis{
			Name:"ticks",
			NameStyle: chart.StyleShow(),
			Style: chart.StyleShow(), //enables / displays the x-axis
			ValueFormatter: func(v interface{}) string {
				if vf, isFloat := v.(float64); isFloat {
					return fmt.Sprintf(" %0.0f", vf / 1000.0)
				}
				return ""
			},
		},
		YAxis: chart.YAxis{
			Name:"avg",
			NameStyle: chart.StyleShow(),
			Style: chart.StyleShow(), //enables / displays the y-axis
		},
		Series: data,
	}
	//note we have to do this as a separate step because we need a reference to graph
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	buffer := bytes.NewBuffer([]byte{})
	err := graph.Render(chart.PNG, buffer)

	if err != nil {
		fmt.Print(err)
		panic("shoot graphing")
	}
	ioutil.WriteFile(fileName +".png", buffer.Bytes(), 0644)
}