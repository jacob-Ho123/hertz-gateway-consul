package metric

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var (
	reg = prometheus.NewRegistry()
	pid int

	SignCheckCnt = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "",
		Subsystem: "",
		Name:      "test_custom_cnt",
		Help:      "",
	}, []string{"code", "reason", "host", "url", "region"})
)

func init() {
	pid = os.Getpid()
}

func MustRegister(collector ...prometheus.Collector) {
	reg.MustRegister(collector...)
}

func InitPrometheus() {

	reg.MustRegister(collectors.NewBuildInfoCollector())
	reg.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{
		PidFn: func() (int, error) {
			return pid, nil
		},
		Namespace:    "",
		ReportErrors: false,
	}))
	reg.MustRegister(SignCheckCnt)
}

func GetRegistry() *prometheus.Registry {
	return reg
}
