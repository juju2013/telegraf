package metric

import (
  "github.com/influxdata/telegraf"
)

func (m *metric) Export() *telegraf.RawMetric {
  return &telegraf.RawMetric {
    Name: m.name,
    Tags: m.tags,
    Fields : m.fields,
    T: m.t,
    MType: m.mType,
    Aggregate: m.aggregate,
  }
}

func Import(rm *telegraf.RawMetric) telegraf.Metric {
	out := metric{
    name: rm.Name,
    tags: rm.Tags,
    fields : rm.Fields,
    t: rm.T,
    mType: rm.MType,
    aggregate: rm.Aggregate,
	}
	return &out
}
