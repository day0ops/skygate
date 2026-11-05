package domain

import "testing"

func TestIsValidFlightStatus(t *testing.T) {
	for _, status := range ValidFlightStatuses {
		if !IsValidFlightStatus(string(status)) {
			t.Errorf("expected %q to be valid", status)
		}
	}
	if IsValidFlightStatus("orbiting") {
		t.Error("expected an unknown status to be invalid")
	}
	if IsValidFlightStatus("") {
		t.Error("expected an empty status to be invalid")
	}
}
