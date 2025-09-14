package easing

import (
	"math"
	"sync"
	"time"
)

type Controller struct {
	mu             *sync.Mutex
	updateInterval time.Duration
}

type UpdateCallback func(float64, bool)

type easingFunc func(float64) float64

func New(updateInterval time.Duration) *Controller {
	return &Controller{
		mu:             new(sync.Mutex),
		updateInterval: updateInterval,
	}
}

func (c *Controller) EaseInCubicFromTo(from, to float64, duration time.Duration, updateCallback UpdateCallback) <-chan struct{} {
	return c.ease(duration, easeInCubic, wrapEasingWithFromTo(from, to, updateCallback))
}

func (c *Controller) EaseInCubic(duration time.Duration, updateCallback UpdateCallback) <-chan struct{} {
	return c.ease(duration, easeInCubic, updateCallback)
}

func (c *Controller) EaseOutCubicFromTo(from, to float64, duration time.Duration, updateCallback UpdateCallback) <-chan struct{} {
	return c.ease(duration, easeOutCubic, wrapEasingWithFromTo(from, to, updateCallback))
}

func (c *Controller) EaseOutCubic(duration time.Duration, updateCallback UpdateCallback) <-chan struct{} {
	return c.ease(duration, easeOutCubic, updateCallback)
}

func (c *Controller) ease(duration time.Duration, easingFunc easingFunc, updateCallback UpdateCallback) <-chan struct{} {
	ch := make(chan struct{})
	c.mu.Lock()

	start := time.Now()
	end := start.Add(duration)
	delta := end.Sub(start)

	go func() {
		t := time.NewTicker(c.updateInterval)
		defer c.mu.Unlock()
		defer close(ch)
		defer t.Stop()

		for tt := range t.C {
			isFinish := tt.After(end) || tt.Equal(end)
			if isFinish {
				updateCallback(1.0, true)
				return
			}
			elapsed := tt.Sub(start)
			d := elapsed.Seconds() / delta.Seconds()
			updateCallback(easingFunc(d), false)
		}
	}()

	return ch
}

func wrapEasingWithFromTo(from, to float64, cb UpdateCallback) UpdateCallback {
	delta := to - from
	return func(v float64, isFinish bool) {
		cb(from+(delta*v), isFinish)
	}
}

func easeInCubic(x float64) float64 {
	return x * x * x
}

func easeOutCubic(x float64) float64 {
	return 1 - math.Pow(1-x, 3)
}
