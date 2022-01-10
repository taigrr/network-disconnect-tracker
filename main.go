package main

import (
	"fmt"

	"github.com/nakabonne/tstorage"
)

func main() {
	storage, _ := tstorage.NewStorage(
		tstorage.WithDataPath("./data"),
		tstorage.WithTimestampPrecision(tstorage.Seconds))
	defer storage.Close()

	_ = storage.InsertRows([]tstorage.Row{
		{
			Metric:    "metric1",
			DataPoint: tstorage.DataPoint{Timestamp: 1600000000, Value: 0.1},
		},
	})
	points, _ := storage.Select("metric1", nil, 1600000000, 1600000001)
	for _, p := range points {
		fmt.Printf("timestamp: %v, value: %v\n", p.Timestamp, p.Value)
		// => timestamp: 1600000000, value: 0.1
	}

}
