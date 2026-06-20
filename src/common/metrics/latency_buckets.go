package metrics

// LatencySecondsBuckets are second-scaled histogram bounds (the SDK defaults are ms-scaled, useless for seconds).
var LatencySecondsBuckets = []float64{
	0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10,
}

// LongLatencySecondsBuckets extend the tail for seconds-to-minutes ops (LLM calls, inference loops).
var LongLatencySecondsBuckets = []float64{
	0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20, 30, 60, 120,
}
