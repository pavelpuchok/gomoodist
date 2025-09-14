package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path"

	"github.com/pavelpuchok/gomoodist/internal/app"
	"github.com/pavelpuchok/gomoodist/internal/app/assets"
	"github.com/pavelpuchok/gomoodist/internal/config"
	"github.com/pavelpuchok/gomoodist/internal/server"
	"github.com/pavelpuchok/gomoodist/internal/ui"
)

func main() {
	sounds, err := assets.ReadSounds(assets.BuiltInSounds)
	if err != nil {
		panic(err)
	}

	configPath := getConfigFilePath()
	conf, err := getConfig(configPath, sounds)
	if err != nil {
		panic(err)
	}

	app, err := app.New(*conf, configPath, sounds)
	if err != nil {
		panic(err)
	}

	app.Run(context.Background())

	srv := server.HTTP{
		App:  app,
		Addr: ":8080",
	}
	go srv.ListenAndServe()

	systray := ui.Systray{
		OpenSettings: func() {
			// using xdg browser open link to localhost :8080
			cmd := exec.Command("xdg-open", "http://localhost:8080")
			err := cmd.Start()
			if err != nil {
				slog.Error("Cannot open settings in browser", slog.String("error", err.Error()))
			}

			go func() {
				if err := cmd.Wait(); err != nil {
					slog.Error("Settings URL opening failed", slog.String("error", err.Error()))
				}
			}()
		},

		TogglePlay: func() bool {
			playing, err := app.TogglePlay()
			if err != nil {
				slog.Error("Toggling play failed", slog.String("error", err.Error()))
				return playing
			}
			return playing
		},
	}
	systray.Run()
}

func getConfig(configPath string, sounds map[assets.SoundID]assets.Sound) (*config.Configuration, error) {
	_, err := os.Stat(configPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("unexpected error while stat config path %s. %w", configPath, err)
		}

		err = os.MkdirAll(path.Dir(configPath), os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("unable to make parent dirs for default config. %w", err)
		}

		cfg := config.New(sounds)
		if err := config.WriteConfig(configPath, cfg); err != nil {
			return nil, err
		}
	}

	cfg, err := config.NewFromFile(configPath, sounds)
	if err != nil {
		return nil, fmt.Errorf("unable create config from file. %w", err)
	}
	return &cfg, nil
}

func getConfigFilePath() string {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configPath != "" {
		return *configPath
	}

	configHome := os.Getenv("XDG_CONFIG_HOME")
	path := path.Join(configHome, "gomoodist", "config.json")

	return path
}
