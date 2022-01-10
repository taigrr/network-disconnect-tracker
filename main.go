package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-ping/ping"
	"github.com/nakabonne/tstorage"
	"github.com/taigrr/go-wireless"
)

var exit chan bool

func init() {
	uid := os.Getuid()
	if uid != 0 {
		fmt.Println("This program must be run as root.")
		os.Exit(1)
	}
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	exit = make(chan bool, 1)
	ifaces := wireless.Interfaces()
	if len(ifaces) < 1 {
		fmt.Println("Error: no wireless interfaces detected!")
		os.Exit(1)
	}
	storage, _ := tstorage.NewStorage(
		tstorage.WithDataPath("./data"),
		tstorage.WithTimestampPrecision(tstorage.Seconds),
	)
	defer storage.Close()

	wifiCard := ifaces[0]
	if *debug {
		fmt.Printf("using card %s\n", wifiCard)
	}
	var err error
	connection, err := wireless.Dial(wifiCard)
	if err != nil {
		log.Panicf("Error dialing connection: %v\n", err)
	}
	wc := wireless.NewClientFromConn(connection)
	go cleanup(sigs)
	go checkConnection(&storage)
	go getNetworks(&storage, wc)
	go checkPing(&storage)
	go printData(&storage)
	go sendData(&storage)
	go logEvents(&storage, connection)
	select {
	case <-exit:
	}
}

func logEvents(storage *tstorage.Storage, conn *wireless.Conn) {
	sub := conn.Subscribe(wireless.EventConnected, wireless.EventAuthReject, wireless.EventDisconnected)
	for {
		ev := <-sub.Next()
		switch ev.Name {
		case wireless.EventConnected:
			fmt.Println(ev.Arguments)
		case wireless.EventAuthReject:
			fmt.Println(ev.Arguments)
		case wireless.EventDisconnected:
			fmt.Println(ev.Arguments)
		}
	}
}

func sendData(storage *tstorage.Storage) {
	for {
		time.Sleep(time.Minute * 5)
	}
}
func printData(storage *tstorage.Storage) {
	lastTime := time.Now().Unix()
	points, _ := (*storage).Select("ping", nil, 0, lastTime)
	for _, p := range points {
		fmt.Printf("timestamp: %v, value: %v\n", p.Timestamp, p.Value)
	}
	for {
		time.Sleep(time.Second * 5)
		newTime := time.Now().Unix()
		points, _ := (*storage).Select("ping", nil, lastTime, newTime)
		lastTime = newTime
		for _, p := range points {
			fmt.Printf("timestamp: %v, value: %v\n", p.Timestamp, p.Value)
		}
	}
}
func connected() (ok bool) {
	_, err := http.Get("http://clients3.google.com/generate_204")
	if err != nil {
		return false
	}
	return true
}

func checkConnection(storage *tstorage.Storage) {
	for {
		time.Sleep(time.Millisecond * 1500)
		value := 0.0
		if connected() {
			value = 1.0
		}
		(*storage).InsertRows([]tstorage.Row{
			{
				Metric:    "connectivity",
				DataPoint: tstorage.DataPoint{Timestamp: time.Now().Unix(), Value: value},
			},
		})
	}
}
func checkPing(storage *tstorage.Storage) {
	for {
		if connected() {
			pinger, err := ping.NewPinger("www.google.com")
			if err != nil {
				panic(err)
			}
			pinger.Count = 5
			err = pinger.Run()
			if err != nil {
				log.Printf("Error: %v \n", err)
			}
			stats := pinger.Statistics()
			(*storage).InsertRows([]tstorage.Row{
				{
					Metric:    "ping",
					DataPoint: tstorage.DataPoint{Timestamp: time.Now().Unix(), Value: float64(stats.AvgRtt.Microseconds())},
				},
			})
		}
		time.Sleep(time.Second * 15)
	}
}
func getNetworks(storage *tstorage.Storage, wc *wireless.Client) {
	defer wc.Close()

	for {
		aps, err := wc.Scan()
		if err != nil {
			log.Printf("Error scanning: %v \n", err)
		}
		for _, ap := range aps {
			fmt.Println(ap)
		}
		time.Sleep(time.Minute / 6)
	}
}

// if we receive a signal, shut down cleanly
func cleanup(sigs chan os.Signal) {
	<-sigs
	if *debug {
		fmt.Println("Signal received, shutting down...")
	}
	exit <- true
}
