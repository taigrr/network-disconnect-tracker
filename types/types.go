package types

import (
	"fmt"
	"time"
)

type NetworkCollection []Network

type ConnectivityCollection []Connectivity

type PingCollection []Ping

type Network struct {
	SSID      string    `json:"ssid"`
	BSSID     string    `json:"bssid"`
	Channel   int      `json:"channel"`
	RSSI      int       `json:"rssi"`
	Timestamp time.Time `json:"timestamp"`
}
type Connectivity struct {
	Timestamp time.Time `json:"timestamp"`
	Connected bool      `json:"connected"`
}
type MetricSet struct {
	Networks     NetworkCollection      `json:"networks"`
	Connectivity ConnectivityCollection `json:"connectivity"`
	Pings        PingCollection         `json:"pings"`
}

type Ping struct {
	Timestamp time.Time `json:"timestamp"`
	RTT       int64`json:"rtt"`
}

func (n Network) String() string {
	return fmt.Sprintf("%s %s %d %d", n.SSID, n.BSSID, n.Channel, n.RSSI)
}
