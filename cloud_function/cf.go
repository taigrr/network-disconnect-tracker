package cloud_function

import (
	"github.com/taigrr/network_disconnect_tracker/cloud_function/bq"
	"github.com/taigrr/network_disconnect_tracker/types"

	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	Package = "save_metrics"
)

var (
	auth       string
	v          bool
	version    bool
	localCount int
)

func init() {
	auth = os.Getenv("METRICS_KEY")
	flag.BoolVar(&version, "version", false, "Get detailed version string")
	flag.BoolVar(&v, "v", false, "Get detailed version string")
	flag.Parse()
}

func Ingest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("key") != auth {
		w.WriteHeader(401)
		return
	}
	loop := r.URL.Query().Get("loop")
	if loop == "" {
		w.WriteHeader(400)
		return
	}
	switch r.Method {
	case "GET":
		fallthrough
	case "HEAD":
		fallthrough
	case "DELETE":
		fallthrough
	case "PUT":
		fallthrough
	case "PATCH":
		fmt.Fprintf(w, "Only POST supported.")
		return
	case "POST":
		break
	default:
		// Should be impossible to reach, Methods are listened on explicitly
		log.Printf("<%d> [0] Unknown verb: %s", localCount, r.Method)
		return
	}

	var buffer []byte
	var err error
	buffer, err = ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(400)
		log.Panic(err)
	}
	var data types.MetricSet
	json.Unmarshal(buffer, &data)
	bq.Insert(data, loop)

	w.WriteHeader(200)
	localCount++
	return
}
