package gobhttp

import (
	"bytes"
	"encoding/gob"
	"net/http"
  "log"
  
	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/plugins/outputs"
)

type gobhttp struct {
	Url string
  Jwt string
}

var sampleConfig = `
  ## The full URL of HTTP end point to poste your data.
  url = ["http://localhost:9001/api"] # required

  ## Additional a cookies jwt_token 
  jwt = "=xxx.yyy.zzz"

  ## Write timeout for the http client, formatted as a string.
  ## If not provided, will default to 5s. 0s means no timeout (not recommended).
  timeout = "5s"
`

func (s *gobhttp) Description() string {
	return "gob --> http output"
}

func (s *gobhttp) SampleConfig() string {
	return sampleConfig
}

func (s *gobhttp) Connect() error {
	return nil
}

func (s *gobhttp) Close() error {
	return nil
}

func (s *gobhttp) Write(metrics []telegraf.Metric) error {

	var b bytes.Buffer // post buffer
  
  // build raw metrics slice
  mex := make([]*telegraf.RawMetric, cap(metrics))
	for i, m := range metrics {
    mex[i]=m.Export()
	}
  
  // serialize with gob
  gob.Register(telegraf.RawMetric{})
  err := gob.NewEncoder(&b).Encode(mex)
  if err != nil {
  	log.Printf("W! Cannot gob: %v\n", err)
    return err
  }

	req, err := http.NewRequest("POST", s.Url, &b)
  req.AddCookie(&http.Cookie{Name: "jwt_token", Value: s.Jwt})
  
	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
  	log.Printf("W! Cannot send to %v@%v \n", s.Url, err)
	}
  defer req.Body.Close()
	return nil
}

func init() {
	outputs.Add("gobhttp", func() telegraf.Output { return &gobhttp{} })
}
