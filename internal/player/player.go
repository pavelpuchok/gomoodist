package player

import (
	"context"
	"log/slog"
	"time"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/effects"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/pavelpuchok/gomoodist/internal/app/assets"
	"github.com/pavelpuchok/gomoodist/internal/config"
	"github.com/pavelpuchok/gomoodist/internal/easing"
)

var volumeMultiplicator float64 = 10

var (
	MinVolume float64 = volumeMultiplicator * -1
	MaxVolume float64 = 0
)

type Player struct {
	sound            *assets.Sound
	easing           *easing.Controller
	initiallyEnabled bool

	volume *effects.Volume
	loop   beep.Streamer
	ctrl   *beep.Ctrl

	cmdCh chan command
}

type command int

const (
	cmdEnable command = iota
	cmdDisable
)

func NewPlayer(sound *assets.Sound, volume float64, enabled bool) (*Player, error) {
	loop, err := beep.Loop2(sound.Streamer)
	if err != nil {
		return nil, err
	}

	vol := &effects.Volume{
		Base:     2,
		Volume:   calculateVolume(volume),
		Streamer: loop,
	}

	ctrl := &beep.Ctrl{
		Streamer: vol,
		Paused:   !enabled,
	}

	return &Player{
		sound:  sound,
		easing: easing.New(25 * time.Millisecond),
		volume: vol,
		loop:   loop,
		ctrl:   ctrl,
		cmdCh:  make(chan command),
	}, nil
}

func (p *Player) Streamer() beep.Streamer {
	return p.ctrl
}

func (p *Player) Run(ctx context.Context) {
	go p.run(ctx)
}

func (p *Player) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			// TODO: free resources
			return

		case msg := <-p.cmdCh:
			switch msg {
			case cmdEnable:
				speaker.Lock()
				originalVolume := p.volume.Volume
				if !p.ctrl.Paused {
					continue
				}
				p.ctrl.Paused = false
				p.volume.Volume = MinVolume
				speaker.Unlock()

				slog.Info("Sound volume increasing started", slog.String("sound", string(p.sound.ID)), slog.Float64("originalVolume", originalVolume))
				<-p.easing.EaseInCubicFromTo(MinVolume, originalVolume, 3*time.Second, func(val float64, end bool) {
					speaker.Lock()
					defer speaker.Unlock()
					p.volume.Volume = val
					if end {
						p.volume.Volume = originalVolume
						slog.Info("Sound volume set", slog.String("sound", string(p.sound.ID)), slog.Float64("volume", p.volume.Volume))
					}
				})

			case cmdDisable:
				speaker.Lock()
				originalVolume := p.volume.Volume
				if p.ctrl.Paused {
					continue
				}
				speaker.Unlock()

				slog.Info("Sound volume decreasing started", slog.String("sound", string(p.sound.ID)))
				<-p.easing.EaseOutCubicFromTo(originalVolume, MinVolume, 3*time.Second, func(val float64, end bool) {
					speaker.Lock()
					defer speaker.Unlock()
					p.volume.Volume = val
					if end {
						p.ctrl.Paused = true
						p.volume.Volume = originalVolume
						slog.Info("Sound volume set", slog.String("sound", string(p.sound.ID)), slog.Float64("volume", p.volume.Volume))
					}
				})
			}
		}
	}
}

func (p *Player) Enable() {
	p.cmdCh <- cmdEnable
}

func (p *Player) Disable() {
	p.cmdCh <- cmdDisable
}

func (p *Player) SetVolume(v float64) {
	speaker.Lock()
	defer speaker.Unlock()
	p.volume.Volume = calculateVolume(v)
}

func (p *Player) Config() config.Sound {
	return config.Sound{
		Volume:  (p.volume.Volume + volumeMultiplicator) / volumeMultiplicator,
		Enabled: !p.ctrl.Paused,
	}
}

func calculateVolume(v float64) float64 {
	if v > 1 {
		v = 1
	}

	if v < 0 {
		v = 0
	}

	return -volumeMultiplicator - (v * -volumeMultiplicator)
}
