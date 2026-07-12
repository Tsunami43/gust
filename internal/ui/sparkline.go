package ui

import "strings"

// sparkRunes are the eight block heights used to draw sparklines.
var sparkRunes = []rune("▁▂▃▄▅▆▇█")

// Sparkline renders a slice of values as a compact single-line bar graph.
// An empty input yields an empty string.
func Sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}
	min, max := values[0], values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}
	span := max - min

	var b strings.Builder
	for _, v := range values {
		idx := 0
		if span > 0 {
			idx = int((v-min)/span*float64(len(sparkRunes)-1) + 0.5)
		}
		b.WriteRune(sparkRunes[idx])
	}
	return b.String()
}
