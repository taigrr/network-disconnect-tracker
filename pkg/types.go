package main

import "fmt"

type NetworkCollection []Network

type Network struct {
	SSID    string
	BSSID   string
	Channel uint
	RSSI    int
}

func (n Network) String() string {
	return fmt.Sprintf("%s %s %d %d", n.SSID, n.BSSID, n.Channel, n.RSSI)
}
