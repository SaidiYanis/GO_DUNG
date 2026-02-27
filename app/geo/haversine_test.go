package geo

import "testing"

func TestHaversineMeters(t *testing.T) {
	// Approx distance between Eiffel Tower and Louvre (~3160m)
	d := HaversineMeters(48.85837, 2.294481, 48.860611, 2.337644)
	if d < 3000 || d > 3400 {
		t.Fatalf("unexpected distance %.2f", d)
	}
}
