package assets

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/vorbis"
)

//go:embed sounds/**/*
var BuiltInSounds embed.FS

type Sound struct {
	FS    fs.FS
	Path  string
	Name  string
	Group string
	ID    SoundID

	Streamer beep.StreamSeekCloser
	Format   beep.Format

	f fs.File
}

type SoundID string

func (s *Sound) InitializeStreamer() error {
	var err error
	s.f, err = s.FS.Open(s.Path)
	if err != nil {
		return err
	}

	s.Streamer, s.Format, err = vorbis.Decode(s.f)
	if err != nil {
		return err
	}

	return nil
}

func ReadSounds(soundsDir fs.FS) (map[SoundID]Sound, error) {
	entries, err := fs.ReadDir(soundsDir, "sounds")
	if err != nil {
		return nil, fmt.Errorf("unable to read sounds dir. %w", err)
	}

	var result map[SoundID]Sound = map[SoundID]Sound{}
	for _, v := range entries {
		if !v.IsDir() {
			continue
		}

		sounds, err := readSoundsGroup(soundsDir, v.Name())
		if err != nil {
			return nil, fmt.Errorf("unable to read sounds group dir %s. %w", v.Name(), err)
		}

		for _, v := range sounds {
			result[v.ID] = v
		}
	}

	return result, nil
}

func readSoundsGroup(dir fs.FS, groupName string) ([]Sound, error) {
	entries, err := fs.ReadDir(dir, path.Join("sounds", groupName))
	if err != nil {
		return nil, err
	}

	var result []Sound = make([]Sound, 0, len(entries))

	for _, v := range entries {
		if v.IsDir() {
			continue
		}

		name := strings.Replace(v.Name(), path.Ext(v.Name()), "", 1)

		id := SoundID(groupName + "/" + name)

		sound := Sound{
			FS:    dir,
			Path:  path.Join("sounds", groupName, v.Name()),
			Name:  name,
			Group: groupName,
			ID:    id,
		}

		if err := sound.InitializeStreamer(); err != nil {
			return nil, fmt.Errorf("Can't initialize sound %s streamer. %w", name, err)
		}

		result = append(result, sound)
	}

	return result, nil
}
