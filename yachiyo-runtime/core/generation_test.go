package core

import (
	"fmt"
	"testing"
)

func TestParseOutput(t *testing.T) {
	cases := []struct {
		intensity string
		value     float64
		result    bool
	}{
		{"high", 0.5, false},
		{"slight", 0, true},
		{"moderate", 0.5, true},
		{"high", 1.0, true},
		{"unknown", 0.5, false},
	}

	for _, c := range cases {
		raw := fmt.Sprintf(
			`{
    "reply": true,
    "answer": [
        "hi"
    ],
    "change": {
        "affect_press": [
            {
                "emotion": "%s",
                "intensity": "%s",
                "value": %v
            }
        ]
    },
    "note": "-1"
}`,
			"Joy", c.intensity, c.value)
		if _, err := parseOutput(raw); (err == nil) != c.result {
			t.Errorf("%s/%v: err = %v, want ok = %v", c.intensity, c.value, err, c.result)
		}
	}
}
