// Package ema provides an exponential moving average. It can hold floating
// point values up to 3 decimal points in precision and provides a convenience
// interface for keeping EMAs of time.Durations.
package ema

import (
	"math"
	"sync/atomic"
	"time"
)

const (
	// floating point values are stored to this scale (3 digits behind decimal
	// point).
	scale = 1000

	unset = math.MinInt64
)

// EMA holds the Exponential Moving Average of a float64 with a the given
// default α value and a fixed scale of 3 digits. Safe to access concurrently.
// https://en.wikipedia.org/wiki/Moving_average#Exponential_moving_average.
type EMA struct {
	defaultAlpha float64
	defaultValue float64
	v            int64
}

// New creates an EMA with the given default value and alpha
func New(defaultValue float64, defaultAlpha float64) *EMA {
	return &EMA{defaultAlpha: defaultAlpha, defaultValue: defaultValue, v: unset}
}

// NewDuration is like New but using time.Duration
func NewDuration(defaultValue time.Duration, defaultAlpha float64) *EMA {
	return New(float64(defaultValue), defaultAlpha)
}

// UpdateAlpha calculates and stores new EMA based on the duration and α
// value passed in.
func (e *EMA) UpdateAlpha(v float64, α float64) float64 {
	oldInt := atomic.LoadInt64(&e.v)
	var newInt int64
	var newEMA float64
	if oldInt == unset {
		newInt = scaleToInt(v)
		newEMA = v
	} else {
		oldEMA := scaleFromInt(oldInt)
		newEMA = (1-α)*oldEMA + α*v
		newInt = scaleToInt(newEMA)
	}
	atomic.StoreInt64(&e.v, newInt)
	return newEMA
}

// Update is like UpdateAlpha but using the default alpha
func (e *EMA) Update(v float64) float64 {
	return e.UpdateAlpha(v, e.defaultAlpha)
}

// UpdateDuration is like Update but using time.Duration
func (e *EMA) UpdateDuration(v time.Duration) time.Duration {
	return time.Duration(e.Update(float64(v)))
}

// Set sets the EMA directly.
func (e *EMA) Set(v float64) {
	atomic.StoreInt64(&e.v, scaleToInt(v))
}

// SetDuration is like Set but using time.Duration
func (e *EMA) SetDuration(v time.Duration) {
	e.Set(float64(v))
}

// Clear clears the EMA
func (e *EMA) Clear() {
	atomic.StoreInt64(&e.v, unset)
}

// Get gets the EMA
func (e *EMA) Get() float64 {
	oldInt := atomic.LoadInt64(&e.v)
	if oldInt == unset {
		return e.defaultValue
	}
	return scaleFromInt(oldInt)
}

// GetDuration is like Get but using time.Duration
func (e *EMA) GetDuration() time.Duration {
	return time.Duration(e.Get())
}

func scaleToInt(f float64) int64 {
	return int64(f * scale)
}

func scaleFromInt(i int64) float64 {
	return float64(i) / scale
}
