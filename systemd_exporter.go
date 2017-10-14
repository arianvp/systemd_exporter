package main

import (
	"flag"
	"github.com/coreos/go-systemd/dbus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strings"
)

type metric struct {
	desc      *prometheus.Desc
	valueType prometheus.ValueType
}
type metricsMap map[string]metric

type unitMetricsMap map[string]metricsMap

type systemdCollector struct {
	conn    *dbus.Conn
	metrics unitMetricsMap
}

func (c *systemdCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range c.metrics {
		for _, v := range v {
			ch <- v.desc
		}
	}
}

func (c *systemdCollector) Collect(ch chan<- prometheus.Metric) {
	unitStatuses, err := c.conn.ListUnits()
	if err != nil {
		log.Fatal(err)
	}
	for _, unitStatus := range unitStatuses {
		splitted := strings.Split(unitStatus.Name, ".")
		unitType := strings.Title(splitted[1])
		for propName, metric := range c.metrics[unitType] {
			prop, err := c.conn.GetUnitTypeProperty(unitStatus.Name, unitType, propName)
			if err == nil {
				value, ok := prop.Value.Value().(uint64)
				if ok {
					if value != ^uint64(0) {
						ch <- prometheus.MustNewConstMetric(metric.desc, metric.valueType, float64(value), unitStatus.Name)
					}
				}
			} else {
				log.Print(propName, ":", err)
			}
		}
	}
}

func newSystemdCollector(conn *dbus.Conn) *systemdCollector {
	return &systemdCollector{
		conn: conn,
		metrics: unitMetricsMap{
			"Service": metricsMap{
				"CPUUsageNSec": metric{
					desc:      prometheus.NewDesc("systemd_service_cpu_usage_nanoseconds_total", "Total CPU seconds of a unit", []string{"unit"}, nil),
					valueType: prometheus.CounterValue,
				},
				"IPIngressBytes": metric{
					desc:      prometheus.NewDesc("systemd_service_ip_ingress_bytes_total", "Ingress bytes total", []string{"unit"}, nil),
					valueType: prometheus.CounterValue,
				},
				"IPIngressPackets": metric{
					desc:      prometheus.NewDesc("systemd_service_ip_ingress_packets_total", "Ingress packets total", []string{"unit"}, nil),
					valueType: prometheus.CounterValue,
				},
				"IPEgressBytes": metric{
					desc:      prometheus.NewDesc("systemd_service_ip_egress_bytes_total", "Egress bytes total", []string{"unit"}, nil),
					valueType: prometheus.CounterValue,
				},
				"IPEgressPackets": metric{
					desc:      prometheus.NewDesc("systemd_service_ip_eggress_packets_total", "Eggress packets total", []string{"unit"}, nil),
					valueType: prometheus.CounterValue,
				},
				"MemoryCurrent": metric{
					desc:      prometheus.NewDesc("systemd_service_memory_current_bytes", "Amount of bytes", []string{"unit"}, nil),
					valueType: prometheus.GaugeValue,
				},
				"TasksCurrent": metric{
					desc:      prometheus.NewDesc("systemd_service_tasks_current", "amount of tasks. Includes both user processes and kernel threads.", []string{"unit"}, nil),
					valueType: prometheus.GaugeValue,
				},
			},
		},
	}
}

var (
	addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
)

func main() {
	flag.Parse()
	conn, err := dbus.New()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	systemdCollector := newSystemdCollector(conn)
	prometheus.MustRegister(systemdCollector)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
