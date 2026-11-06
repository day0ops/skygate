package sim

import (
	"hash/fnv"
	"math"
	"math/rand"
	"time"

	"skygate/internal/domain"
)

const (
	minCurveOffset = 20.0
	maxCurveOffset = 80.0
)

// flightSeed derives a deterministic seed from a flight ID and a purpose
// label, so different random decisions about the same flight (curve shape,
// hold decision) don't correlate with each other just because they share an
// ID, while still being perfectly stable across every tick for that flight.
func flightSeed(flightID, purpose string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(flightID + ":" + purpose))
	return int64(h.Sum64())
}

func flightRand(flightID, purpose string) *rand.Rand {
	return rand.New(rand.NewSource(flightSeed(flightID, purpose)))
}

// flightCurve holds the Bezier control points for one flight's curved path.
// Deterministically derived from the flight's ID (see newFlightCurve), so
// the same flight always computes the same curve shape on every tick
// without storing anything new on domain.Flight.
type flightCurve struct {
	wavy       bool
	controlPts [][2]float64 // 1 point for a single arc, 2 for a wavy S-curve.
}

// newFlightCurve picks a curve shape for flightID: a single gentle arc, or a
// wavier S-curve with two bends, each randomized in which side it bulges to
// and by how much.
func newFlightCurve(flightID string, origin, dest domain.Airport) flightCurve {
	r := flightRand(flightID, "curve")

	dx, dy := dest.X-origin.X, dest.Y-origin.Y
	length := math.Hypot(dx, dy)
	var perpX, perpY float64
	if length > 0 {
		perpX, perpY = -dy/length, dx/length
	}

	randomOffset := func() float64 {
		mag := minCurveOffset + r.Float64()*(maxCurveOffset-minCurveOffset)
		if r.Float64() < 0.5 {
			mag = -mag
		}
		return mag
	}

	if r.Float64() < 0.5 {
		offset := randomOffset()
		mid := [2]float64{origin.X + dx*0.5, origin.Y + dy*0.5}
		return flightCurve{
			wavy:       false,
			controlPts: [][2]float64{{mid[0] + perpX*offset, mid[1] + perpY*offset}},
		}
	}

	offset1 := randomOffset()
	offset2 := -offset1 * (0.5 + r.Float64()*0.5)
	p1 := [2]float64{origin.X + dx/3 + perpX*offset1, origin.Y + dy/3 + perpY*offset1}
	p2 := [2]float64{origin.X + dx*2/3 + perpX*offset2, origin.Y + dy*2/3 + perpY*offset2}
	return flightCurve{wavy: true, controlPts: [][2]float64{p1, p2}}
}

// pointAt returns the (x, y) position and heading in degrees (atan2-from-East
// convention - the caller applies the icon-orientation offset) at progress t
// in [0, 1] along the curve.
func (c flightCurve) pointAt(origin, dest domain.Airport, t float64) (x, y, headingDeg float64) {
	if !c.wavy {
		p1 := c.controlPts[0]
		mt := 1 - t
		x = mt*mt*origin.X + 2*mt*t*p1[0] + t*t*dest.X
		y = mt*mt*origin.Y + 2*mt*t*p1[1] + t*t*dest.Y
		tx := 2*mt*(p1[0]-origin.X) + 2*t*(dest.X-p1[0])
		ty := 2*mt*(p1[1]-origin.Y) + 2*t*(dest.Y-p1[1])
		return x, y, math.Atan2(ty, tx) * 180 / math.Pi
	}

	p1, p2 := c.controlPts[0], c.controlPts[1]
	mt := 1 - t
	x = mt*mt*mt*origin.X + 3*mt*mt*t*p1[0] + 3*mt*t*t*p2[0] + t*t*t*dest.X
	y = mt*mt*mt*origin.Y + 3*mt*mt*t*p1[1] + 3*mt*t*t*p2[1] + t*t*t*dest.Y
	tx := 3*mt*mt*(p1[0]-origin.X) + 6*mt*t*(p2[0]-p1[0]) + 3*t*t*(dest.X-p2[0])
	ty := 3*mt*mt*(p1[1]-origin.Y) + 6*mt*t*(p2[1]-p1[1]) + 3*t*t*(dest.Y-p2[1])
	return x, y, math.Atan2(ty, tx) * 180 / math.Pi
}

const (
	holdChance      = 0.28
	holdRadius      = 18.0
	holdingDuration = 10 * time.Second
)

// holdDecision reports whether flightID holds before landing, and if so how
// many loops - derived deterministically from the flight's ID, exactly like
// newFlightCurve, so it's stable across ticks without storing new state.
func holdDecision(flightID string) (holds bool, loops int) {
	r := flightRand(flightID, "hold")
	if r.Float64() >= holdChance {
		return false, 0
	}
	if r.Float64() < 0.5 {
		return true, 1
	}
	return true, 2
}

// holdingPosition computes a circular loop around dest, completing loops
// total loops as elapsed advances from 0 to holdingDuration.
func holdingPosition(dest domain.Airport, elapsed time.Duration, loops int) (x, y, headingDeg, progress float64) {
	t := float64(elapsed) / float64(holdingDuration)
	theta := 2 * math.Pi * float64(loops) * t
	x = dest.X + holdRadius*math.Cos(theta)
	y = dest.Y + holdRadius*math.Sin(theta)
	tx := -math.Sin(theta)
	ty := math.Cos(theta)
	return x, y, math.Atan2(ty, tx) * 180 / math.Pi, t
}
