package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"git.mills.io/prologic/bitcask"
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
	storage, err := tstorage.NewStorage(
		tstorage.WithDataPath("./data/ts"),
		tstorage.WithTimestampPrecision(tstorage.Seconds),
	)
	if err != nil {
		fmt.Printf("Error: could not initialize time series database: %v\n", err)
		os.Exit(1)
	}
	defer storage.Close()
	db, err := bitcask.Open("./data/bc")
	if err != nil {
		storage.Close()
		fmt.Printf("Error: could not initialize keyval database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()
	if !db.Has([]byte("startKey")) {
		db.Put([]byte("startKey"), []byte(strconv.Itoa(int(time.Now().Unix()))))
	}
	wifiCard := ifaces[0]
	if *debug {
		fmt.Printf("using card %s\n", wifiCard)
	}
	connection, err := wireless.Dial(wifiCard)
	if err != nil {
		log.Panicf("Error dialing connection: %v\n", err)
	}
	wc := wireless.NewClientFromConn(connection)
	go cleanup(sigs)
	go checkConnection(&storage)
	go getNetworks(wc, db)
	go checkPing(&storage)
	go printData(&storage, db)
	go sendData(&storage)
	go logEvents(&storage, connection, db)
	select {
	case <-exit:
	}
}

func logEvents(storage *tstorage.Storage, conn *wireless.Conn, db *bitcask.Bitcask) {
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
func printData(storage *tstorage.Storage, bc *bitcask.Bitcask) {
	startTime := time.Now()
	ft, err := bc.Get([]byte("startKey"))
	if err != nil {

	} else {
		ts := string(ft)
		num, _ := strconv.Atoi(ts)
		startTime = time.Unix(int64(num), 0)
	}
	lastTime := time.Now()
	points, _ := (*storage).Select("ping", nil, startTime.Unix(), lastTime.Unix())
	for _, p := range points {
		fmt.Printf("timestamp: %v, value: %v\n", p.Timestamp, p.Value)
	}
	bc.Range([]byte(strconv.Itoa(int(startTime.Unix()))), []byte(strconv.Itoa(int(lastTime.Unix()))),
		func(key []byte) error {
			val, err := bc.Get(key)
			if err != nil {
				fmt.Printf("Error getting key %v: %v\n", key, err)
			}
			networks := decodeNetworks(val)
			for _, n := range networks {
				fmt.Println(n)
			}
			return nil
		})
	if err != nil {
		fmt.Printf("Error range scanning: %v\n", err)
	}
	fmt.Println("starting loop")
	for {
		time.Sleep(time.Second * 5)
		newTime := time.Now()
		points, _ := (*storage).Select("ping", nil, lastTime.Unix(), newTime.Unix())
		for _, p := range points {
			fmt.Printf("timestamp: %v, value: %v\n", p.Timestamp, p.Value)
		}
		err := bc.Range([]byte(strconv.Itoa(int(lastTime.Unix()))), []byte(strconv.Itoa(int(newTime.Unix()))),
			func(key []byte) error {
				val, err := bc.Get(key)
				if err != nil {
					fmt.Printf("Error getting key %v: %v\n", key, err)
				}
				networks := decodeNetworks(val)
				for _, n := range networks {
					fmt.Println(n)
				}
				return nil
			})
		if err != nil {
			panic(err)
		}
		lastTime = newTime
	}
}

func encodeNetworks(n NetworkCollection) (b []byte) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(n)
	if err != nil {
		fmt.Printf("Error encoding bytes: %v \n", err)
	}
	return buf.Bytes()
}
func decodeNetworks(b []byte) (n NetworkCollection) {
	dec := gob.NewDecoder(bytes.NewBuffer(b))
	err := dec.Decode(&n)
	if err != nil {
		fmt.Printf("Error decoding bytes: %v \n", err)
	}
	return n
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
func getNetworks(wc *wireless.Client, db *bitcask.Bitcask) {
	defer wc.Close()

	for {
		aps, err := wc.Scan()
		if err != nil {
			log.Printf("Error scanning: %v \n", err)
		}
		var networks NetworkCollection
		for _, ap := range aps {
			n := Network{SSID: ap.SSID, BSSID: ap.BSSID.String(), Channel: uint(ap.Frequency), RSSI: ap.Signal}
			networks = append(networks, n)
		}
		now := time.Now()
		db.Put([]byte(strconv.Itoa(int(now.Unix()))), encodeNetworks(networks))
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
