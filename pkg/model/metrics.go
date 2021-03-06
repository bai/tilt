package model

import "time"

var DefaultReportingPeriod = 5 * time.Minute

// Metrics settings generally map to exporter options
// https://pkg.go.dev/contrib.go.opencensus.io/exporter/ocagent?tab=doc#ExporterOption
type MetricsSettings struct {
	Enabled  bool
	Address  string
	Insecure bool

	// How often Tilt reports its metrics. Useful for testing.
	// https://pkg.go.dev/go.opencensus.io/stats/view?tab=doc#SetReportingPeriod
	ReportingPeriod time.Duration

	// Whether anonymous metrics are allowed.
	// The normal tilt prod metrics processor requires the
	// user to be logged in.
	AllowAnonymous bool
}

func DefaultMetricsSettings() MetricsSettings {
	return MetricsSettings{
		Enabled:         false,
		Address:         "opentelemetry.tilt.dev:443",
		ReportingPeriod: DefaultReportingPeriod,
	}
}

// User metrics preferences
type MetricsMode string

const MetricsDefault = MetricsMode("")
const MetricsDisabled = MetricsMode("disabled")
const MetricsLocal = MetricsMode("local")
const MetricsProd = MetricsMode("prod")
