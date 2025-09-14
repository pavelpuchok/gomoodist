package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pavelpuchok/gomoodist/internal/app/assets"
)

type Configuration struct {
	Sounds map[assets.SoundID]Sound `json:"sounds"`
}

type Sound struct {
	Enabled bool    `json:"enabled"`
	Volume  float64 `json:"volume"`
}

func New(sounds map[assets.SoundID]assets.Sound) Configuration {
	var config Configuration
	config.Sounds = make(map[assets.SoundID]Sound, len(sounds))

	for id := range sounds {
		_, has := config.Sounds[id]
		if has {
			continue
		}

		config.Sounds[id] = Sound{
			Volume: 0.7,
		}
	}

	return config
}

func NewFromFile(filePath string, sounds map[assets.SoundID]assets.Sound) (Configuration, error) {
	config := New(sounds)

	raw, err := os.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(raw, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (c Configuration) Marshal() ([]byte, error) {
	return json.Marshal(&c)
}

func WriteConfig(filePath string, config Configuration) error {
	raw, err := config.Marshal()
	if err != nil {
		return fmt.Errorf("unable to marshal default config. %w", err)
	}

	err = os.WriteFile(filePath, raw, os.ModePerm)
	if err != nil {
		return fmt.Errorf("unable to save default config. %w", err)
	}

	return nil
}
