package gobhttp

import (
	"log"
	"net/http"
  "net/url"
	"sync"
  "time"
  "encoding/gob"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/internal"
	"github.com/influxdata/telegraf/plugins/inputs"
  
)

const (
	// DEFAULT_MAX_BODY_SIZE is the default maximum request body size, in bytes.
	// if the request body is over this size, we will return an HTTP 413 error.
	// 500 MB
	DEFAULT_MAX_BODY_SIZE = 500 * 1024 * 1024

	// max_request_size is the maximum size, in bytes, that can be allocated for
	// a single InfluxDB point.
	// 64 KB
	DEFAULT_max_request_size = 64 * 1024
)

type HTTPListener struct {
	Host         string
  Path            string
  Service_url     string
	ReadTimeout    internal.Duration
	WriteTimeout   internal.Duration
	MaxBodySize    int64
	MaxLineSize    int

  httpServer        *http.Server
  
	mu sync.Mutex
	wg sync.WaitGroup

	acc    telegraf.Accumulator
}

const sampleConfig = `
  ## Address and port to host HTTP listener on
  service_url = "http://localhost:9001/api"

  ## maximum duration before timing out read of the request
  read_timeout = "10s"
  ## maximum duration before timing out write of the response
  write_timeout = "10s"

  ## Maximum allowed http request body size in bytes.
  ## 0 means to use the default of 53,687,091 bytes (50 mebibytes)
  max_body_size = 0
`

func (h *HTTPListener) SampleConfig() string {
	return sampleConfig
}

func (h *HTTPListener) Description() string {
	return "gohttp listener"
}

func (h *HTTPListener) Gather(_ telegraf.Accumulator) error {
  // Retrive raw metrics and add to accumulator
	return nil
}

// Start starts the http listener service.
func (h *HTTPListener) Start(acc telegraf.Accumulator) error {
	h.mu.Lock()
	defer h.mu.Unlock()

  u, err := url.Parse(h.Service_url)
  if err != nil {
    return err
  }
  h.Host = u.Host
  h.Path = u.Path

	if h.MaxBodySize == 0 {
		h.MaxBodySize = DEFAULT_MAX_BODY_SIZE
	}

	h.acc = acc


	if h.ReadTimeout.Duration < time.Second {
		h.ReadTimeout.Duration = time.Second *10
	}
	if h.WriteTimeout.Duration < time.Second {
		h.WriteTimeout.Duration = time.Second * 10
	}

  f := func(w http.ResponseWriter, r *http.Request) {
    gobHandler(h, w, r)
  }
  http.HandleFunc(h.Path, f)

  h.httpServer = &http.Server{
    Addr:          h.Host,
    ReadTimeout:    h.ReadTimeout.Duration,
    WriteTimeout:   h.WriteTimeout.Duration,
  }
  
  h.httpServer.SetKeepAlivesEnabled(false)

	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
    go h.httpServer.ListenAndServe()
  }()
	log.Printf("I! Started gobhttp listener service on %s\n", h.Service_url)

	return nil
}

// Stop cleans up all resources
func (h *HTTPListener) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	//h.httpServer.Close()
	h.wg.Wait()

	log.Println("I! Stopped v listener service on ", h.Service_url)
}

func gobHandler(h *HTTPListener, w http.ResponseWriter, r *http.Request) {
  m := []*telegraf.RawMetric{}
  gob.Register(telegraf.RawMetric{})
  
  dec := gob.NewDecoder(http.MaxBytesReader(w, r.Body, h.MaxBodySize))
  err := dec.Decode(&m)
  if err != nil {
    log.Printf("E! %v\n", err)
    // don't stop even error : try to do a patial import 
  }
  
  // import metrics
  for _, rm := range m {
    h.acc.AddRaw(rm)
  }
	log.Printf("D! http-> %v metrics gobed\n", cap(m))
}

func init() {
	inputs.Add("gobhttp", func() telegraf.Input {
		return &HTTPListener{
			Service_url: "http://localhost:9001/api",
		}
	})
}
