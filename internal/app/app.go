package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/pavelpuchok/gomoodist/internal/app/assets"
	"github.com/pavelpuchok/gomoodist/internal/config"
	"github.com/pavelpuchok/gomoodist/internal/player"
)

type UnknownSoundErr struct {
	SoundID assets.SoundID
}

func (s UnknownSoundErr) Error() string {
	return fmt.Sprintf("unknown sound %s", s.SoundID)
}

type App struct {
	mu         *sync.Mutex
	players    map[assets.SoundID]*player.Player
	playing    bool
	configPath string
	dirtyCh    chan struct{}
}

func New(conf config.Configuration, configPath string, sounds map[assets.SoundID]assets.Sound) (*App, error) {
	players, err := initializePlayers(conf, sounds)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize players. %w", err)
	}

	return &App{
		mu:         new(sync.Mutex),
		players:    players,
		configPath: configPath,
		dirtyCh:    make(chan struct{}),
	}, nil
}

func (a *App) Run(ctx context.Context) {
	speaker.Init(beep.SampleRate(44100), 1024*10)

	plrs := make([]beep.Streamer, 0, len(a.players))

	for _, plr := range a.players {
		plrs = append(plrs, plr.Streamer())
		plr.Run(ctx)
	}

	speaker.Play(plrs...)

	a.mu.Lock()
	defer a.mu.Unlock()
	a.playing = true
}

func initializePlayers(conf config.Configuration, sounds map[assets.SoundID]assets.Sound) (map[assets.SoundID]*player.Player, error) {
	players := make(map[assets.SoundID]*player.Player, len(sounds))
	for id, soundCfg := range conf.Sounds {
		asset, has := sounds[id]
		if !has {
			return nil, fmt.Errorf("unknown sound '%s'", id)
		}

		p, err := player.NewPlayer(&asset, soundCfg.Volume, soundCfg.Enabled)
		if err != nil {
			return nil, fmt.Errorf("failed to create new player. Sound %s. %w", id, err)
		}
		players[id] = p
	}

	return players, nil
}

func (a *App) Config() config.Configuration {
	cfg := config.Configuration{
		Sounds: make(map[assets.SoundID]config.Sound, len(a.players)),
	}
	speaker.Lock()
	defer speaker.Unlock()
	for id, plr := range a.players {
		cfg.Sounds[id] = plr.Config()
	}
	return cfg
}

func (a *App) IsPlaying() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.playing
}

func (a *App) TogglePlay() (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.playing {
		a.playing = false
		return a.playing, speaker.Suspend()
	}

	a.playing = true
	return a.playing, speaker.Resume()
}

func (a *App) EnableSound(sid assets.SoundID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	plr, has := a.players[sid]
	if !has {
		return UnknownSoundErr{sid}
	}

	plr.Enable()

	a.saveConfig()
	return nil
}

func (a *App) DisableSound(sid assets.SoundID) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	plr, has := a.players[sid]
	if !has {
		return UnknownSoundErr{sid}
	}

	plr.Disable()

	a.saveConfig()
	return nil
}

func (a *App) SetSoundVolume(sid assets.SoundID, vol float64) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	plr, has := a.players[sid]
	if !has {
		return UnknownSoundErr{sid}
	}

	plr.SetVolume(vol)

	a.saveConfig()
	return nil
}

func (a *App) saveConfig() {
	err := config.WriteConfig(a.configPath, a.Config())
	if err != nil {
		slog.Error("Failed to save config", slog.String("error", err.Error()))
	}
}
