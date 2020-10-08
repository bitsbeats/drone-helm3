package errorhandler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bitsbeats/drone-helm3/internal/helm"
)

type (
	Status string

	Handler interface {
		// Fatalf is used for undefined errors, always exits the program
		Fatalf(message string, v ...interface{})

		// Status is used to propagate a status, always exits the
		// program
		Status(status error, message string, v ...interface{})
	}
)

// Pushgateway is a Handler implementation that reports back to a pushgateway
// server to monitor the outcome
type Pushgateway struct {
	Repo           string
	Namespace      string
	Release        string
	PushGatewayURL string
}

func NewPushgateway(repo, namespace, release, pushGatewayURL string) *Pushgateway {
	return &Pushgateway{
		Repo:           repo,
		Namespace:      namespace,
		Release:        release,
		PushGatewayURL: pushGatewayURL,
	}
}

func (e *Pushgateway) Fatalf(message string, v ...interface{}) {
	(&Log{}).Fatalf(message, v...)
}

func (e *Pushgateway) Status(status error, message string, v ...interface{}) {
	msg := ""
	if status == nil {
		msg = "success"
	} else if wrappedErr, ok := status.(*helm.HelmError); ok {
		msg = wrappedErr.Error()
	} else {
		msg = "undefined"
	}

	buffer := bytes.NewBuffer([]byte("# TYPE drone_helm3_build_status gauge\n"))
	_, _ = fmt.Fprintf(buffer, "drone_helm3_build_status{status=%q} %d\n", msg, time.Now().Unix())
	url := fmt.Sprintf(
		"%s/job/drone_helm3/repo@base64/%s/namespace@base64/%s/release@base64/%s",
		e.PushGatewayURL,
		base64.StdEncoding.EncodeToString([]byte(e.Repo)),
		base64.StdEncoding.EncodeToString([]byte(e.Namespace)),
		base64.StdEncoding.EncodeToString([]byte(e.Release)),
	)
	resp, err := http.Post(url, "text", bytes.NewReader(buffer.Bytes()))
	if err != nil {
		log.Printf("unable to push result to pushgateway host %q: %s", e.PushGatewayURL, err)
	} else if resp.StatusCode >= 400 {
		body, _ := ioutil.ReadAll(resp.Body)
		log.Printf("non [23]xx status code from pushgateway: %d", resp.StatusCode)
		log.Printf("response: \n%s\n\n", string(body))
	}
	if err == nil {
		defer resp.Body.Close()
	}

	// use Log for exitcode handling
	(&Log{}).Status(status, message, v...)
}

// Log is a Handler implementation that just logs
// Note: this is used by other Handlers, more specific code should always go
//       into a more specific implementation
type Log struct{}

func NewLog() *Log {
	return &Log{}
}

func (e *Log) Fatalf(message string, v ...interface{}) {
	log.Fatalf(message, v...)
}

func (e *Log) Status(status error, message string, v ...interface{}) {
	if status == nil {
		log.Printf(message, v...)
		os.Exit(0)
	} else if _, ok := status.(*helm.HelmError); ok {
		log.Fatalf(message, v...)
	} else {
		log.Printf("undefined status reported: %+v", status)
		log.Fatalf(message, v...)
	}
}
