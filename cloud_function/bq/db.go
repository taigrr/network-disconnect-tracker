package bq

import (
	"context"
	"fmt"
	"log"
	"os"

	. "github.com/taigrr/network_disconnect_tracker/types"

	"cloud.google.com/go/bigquery"
)

var (
	connectivityDataSet   string
	connectivityTableName string
	pingsDataSet          string
	pingsTableName        string
	projectID             string
	ssidsDataSet          string
	ssidsTableName        string
)

func init() {
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	connectivityDataSet = os.Getenv("BIGQUERY_DATASET_CONNECTIVITY")
	connectivityTableName = os.Getenv("BIGQUERY_TABLE_CONNECTIVITY")
	pingsDataSet = os.Getenv("BIGQUERY_DATASET_PINGS")
	pingsTableName = os.Getenv("BIGQUERY_TABLE_PINGS")
	ssidsDataSet = os.Getenv("BIGQUERY_DATASET_SSIDS")
	ssidsTableName = os.Getenv("BIGQUERY_TABLE_SSIDS")
	if projectID == "" {
		fmt.Println("GOOGLE_CLOUD_PROJECT environment variable must be set.")
		os.Exit(1)
	}
	if ssidsDataSet == "" {
		fmt.Println("BIGQUERY_DATASET_SSIDS environment variable must be set.")
		os.Exit(1)
	}
	if ssidsTableName == "" {
		fmt.Println("BIGQUERY_TABLE_SSIDS environment variable must be set.")
		os.Exit(1)
	}
	if pingsDataSet == "" {
		fmt.Println("BIGQUERY_DATASET_PINGS environment variable must be set.")
		os.Exit(1)
	}
	if pingsTableName == "" {
		fmt.Println("BIGQUERY_TABLE_PINGS environment variable must be set.")
		os.Exit(1)
	}
	if connectivityDataSet == "" {
		fmt.Println("BIGQUERY_DATASET_CONNECTIVITY environment variable must be set.")
		os.Exit(1)
	}
	if connectivityTableName == "" {
		fmt.Println("BIGQUERY_TABLE_CONNECTIVITY environment variable must be set.")
		os.Exit(1)
	}
}

func Insert(data MetricSet, loop string) {
	ctx := context.Background()
	client, err := bigquery.NewClient(ctx, projectID)
	if err != nil {
		log.Panicf("bigquery.NewClient: %v", err)
	}
	defer client.Close()

	err = queryConnectivity(ctx, client, data.Connectivity, loop)
	if err != nil {
		log.Panicf("bigquery insertion fail: %v", err)
	}

	err = queryPings(ctx, client, data.Pings, loop)
	if err != nil {
		log.Panicf("bigquery insertion fail: %v", err)
	}

	err = queryNetworks(ctx, client, data.Networks, loop)
	if err != nil {
		log.Panicf("bigquery insertion fail: %v", err)
	}
}

func queryNetworks(ctx context.Context, client *bigquery.Client, data NetworkCollection, loop string) error {
	qstring := `INSERT INTO ` + projectID + "." + ssidsDataSet + "." + ssidsTableName + `(ssids, bssid, channel, loop,rssi,timestamp) VALUES`
	qps := []bigquery.QueryParameter{}
	for i, network := range data {
		if i == len(data)-1 {
			qstring += "(?,?,?,?,?,?);"
		} else {
			qstring += "(?,?,?,?,?,?),"
		}
		qps = append(qps, []bigquery.QueryParameter{
			{Value: network.SSID},
			{Value: network.BSSID},
			{Value: network.Channel},
			{Value: loop},
			{Value: network.RSSI},
			{Value: network.Timestamp}}...)
	}
	query := client.Query(qstring)
	job, err := query.Run(ctx)
	if err != nil {
		log.Printf("Error creating ssids query job: %s", err.Error())
		log.Printf("Query: %s", qstring)
		return err
	}
	status, err := job.Wait(ctx)
	if err != nil {
		log.Printf("Error running ssids query: %s", err.Error())
		return err
	}
	err = status.Err()
	return err
}
func queryConnectivity(ctx context.Context, client *bigquery.Client, data ConnectivityCollection, loop string) error {
	qstring := `INSERT INTO ` + projectID + "." + connectivityDataSet + "." + connectivityTableName + `(loop, timestamp, connected) VALUES`
	qps := []bigquery.QueryParameter{}
	for i, connection := range data {
		if i == len(data)-1 {
			qstring += "(?,?,?);"
		} else {
			qstring += "(?,?,?),"
		}
		qps = append(qps,
			[]bigquery.QueryParameter{{Value: loop},
				{Value: connection.Timestamp},
				{Value: connection.Connected}}...)
	}
	query := client.Query(qstring)
	query.Parameters = qps
	job, err := query.Run(ctx)
	if err != nil {
		log.Printf("Error creating connectivity query job: %s", err.Error())
		log.Printf("Query: %s", qstring)
		return err
	}
	stat, err := job.Wait(ctx)
	if err != nil {
		log.Printf("Error running connectivity query: %s", err.Error())
		return err
	}
	if stat.Err() != nil {
		return stat.Err()
	}

	return nil
}

func queryPings(ctx context.Context, client *bigquery.Client, data PingCollection, loop string) error {
	qstring := `INSERT INTO ` + projectID + "." + pingsDataSet + "." + pingsTableName + `(loop, timestamp, rtt) VALUES`
	qps := []bigquery.QueryParameter{}
	for i, ping := range data {
		if i == len(data)-1 {
			qstring += "(?,?,?);"
		} else {
			qstring += "(?,?,?),"
		}
		qps = append(qps,
			[]bigquery.QueryParameter{{Value: loop},
				{Value: ping.Timestamp},
				{Value: ping.RTT}}...)
	}
	query := client.Query(qstring)
	query.Parameters = qps
	job, err := query.Run(ctx)
	if err != nil {
		log.Printf("Error creating ping query job: %s", err.Error())
		log.Printf("Query: %s", qstring)
		return err
	}
	stat, err := job.Wait(ctx)
	if err != nil {
		log.Printf("Error running ping query: %s", err.Error())
		return err
	}
	if stat.Err() != nil {
		return stat.Err()
	}

	return nil
}
