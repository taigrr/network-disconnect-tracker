package dendra_hummingbird_monitor

import (
	"net/http"

	cloud_function "github.com/taigrr/network_disconnect_tracker/cloud_function"
)

func Ingest(w http.ResponseWriter, r *http.Request) {
	cloud_function.Ingest(w, r)
}
