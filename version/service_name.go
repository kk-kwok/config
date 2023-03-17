package version

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

// ServiceName should define in build time
var ServiceName = "undefined"

func PromSubsystem() string {
	return strings.Replace(ServiceName, "-", "_", -1)
}

func MustRegisterVersionCollector() {
	prometheus.MustRegister(NewCollector(PromSubsystem()))
}
