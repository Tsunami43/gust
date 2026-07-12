package ui

// Grade is a simple quality rating derived from a connection's throughput and
// latency.
type Grade struct {
	Letter string
	Label  string
}

// GradeResult rates a connection from its download speed (Mbps) and latency
// (milliseconds). Thresholds are deliberately simple and human-friendly.
func GradeResult(downMbps, latencyMs float64) Grade {
	switch {
	case downMbps >= 200 && latencyMs < 30:
		return Grade{"A+", "excellent"}
	case downMbps >= 100 && latencyMs < 50:
		return Grade{"A", "great"}
	case downMbps >= 50:
		return Grade{"B", "good"}
	case downMbps >= 20:
		return Grade{"C", "fair"}
	case downMbps >= 5:
		return Grade{"D", "slow"}
	default:
		return Grade{"E", "poor"}
	}
}
