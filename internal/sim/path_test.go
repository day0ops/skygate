package sim

import (
	"fmt"
	"math"
	"testing"

	"skygate/internal/domain"
)

func TestFlightCurveIsDeterministicPerFlightID(t *testing.T) {
	origin := domain.Airport{Code: "A", X: 0, Y: 0}
	dest := domain.Airport{Code: "B", X: 100, Y: 0}

	c1 := newFlightCurve("flt_fixed", origin, dest)
	c2 := newFlightCurve("flt_fixed", origin, dest)

	if c1.wavy != c2.wavy || c1.controlPts[0] != c2.controlPts[0] {
		t.Fatalf("expected the same flight ID to always produce the same curve, got %+v and %+v", c1, c2)
	}
}

func TestFlightCurveProducesAMixOfStyles(t *testing.T) {
	origin := domain.Airport{Code: "A", X: 0, Y: 0}
	dest := domain.Airport{Code: "B", X: 100, Y: 0}

	seenArc, seenWavy := false, false
	for i := 0; i < 50; i++ {
		c := newFlightCurve(fmt.Sprintf("flt_%d", i), origin, dest)
		if c.wavy {
			seenWavy = true
		} else {
			seenArc = true
		}
	}
	if !seenArc || !seenWavy {
		t.Fatalf("expected a mix of arc and wavy curves across many flights, got seenArc=%v seenWavy=%v", seenArc, seenWavy)
	}
}

func TestQuadraticArcMidpointTangentMatchesChord(t *testing.T) {
	origin := domain.Airport{Code: "A", X: 0, Y: 0}
	dest := domain.Airport{Code: "B", X: 100, Y: 0}
	curve := flightCurve{wavy: false, controlPts: [][2]float64{{50, 40}}}

	_, _, heading := curve.pointAt(origin, dest, 0.5)

	// A quadratic Bezier's tangent at t=0.5 always equals the chord
	// direction, regardless of where the control point is - a property of
	// the curve, not something that depends on the specific bulge chosen.
	want := math.Atan2(dest.Y-origin.Y, dest.X-origin.X) * 180 / math.Pi
	if diff := math.Abs(heading - want); diff > 0.001 {
		t.Fatalf("heading = %v, want %v (straight bearing)", heading, want)
	}
}

func TestWavyCurveBulgesAwayFromStraightLine(t *testing.T) {
	origin := domain.Airport{Code: "A", X: 0, Y: 0}
	dest := domain.Airport{Code: "B", X: 100, Y: 0}
	curve := flightCurve{wavy: true, controlPts: [][2]float64{{33, 30}, {66, -20}}}

	_, y, _ := curve.pointAt(origin, dest, 0.5)

	// On the straight line, y would be 0 at every point - a wavy curve
	// should visibly deviate from it partway through.
	if math.Abs(y) < 1 {
		t.Fatalf("expected the wavy curve to bulge away from the straight line, got y=%v", y)
	}
}
