package constants

import "testing"

func TestGetInfluenceID(t *testing.T) {

	tests := []struct {
		testName string
		input    string
		want     int
	}{
		{"id found", InfluenceWealth, InfluenceWealthID},
		{"id not found", "failed", 0},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetInfluenceID(tt.input)
			if got != tt.want {
				t.Errorf("got %d want %d given, %v", got, tt.want, tt.input)

			}
		})
	}

}

func TestGetJobID(t *testing.T) {

	tests := []struct {
		testName string
		input    string
		want     int
	}{
		{"id found", JobWarrior, JobWarriorID},
		{"id not found", "failed", 0},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetJobID(tt.input)
			if got != tt.want {
				t.Errorf("got %d want %d given, %v", got, tt.want, tt.input)

			}
		})
	}

}
