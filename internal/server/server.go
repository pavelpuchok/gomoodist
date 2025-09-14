package server

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/pavelpuchok/gomoodist/internal/app"
	"github.com/pavelpuchok/gomoodist/internal/app/assets"
	"github.com/pavelpuchok/gomoodist/internal/config"

	"encoding/json"
)

type App interface {
	Config() config.Configuration
	EnableSound(assets.SoundID) error
	DisableSound(assets.SoundID) error
	SetSoundVolume(assets.SoundID, float64) error
}

type HTTP struct {
	App  App
	Addr string
}

func (s *HTTP) ListenAndServe() error {
	return http.ListenAndServe(s.Addr, s.buildHandler())
}

func (s *HTTP) buildHandler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /", http.FileServer(http.Dir("./internal/server/static/")))

	mux.HandleFunc("GET /api/v1/config", func(w http.ResponseWriter, r *http.Request) {
		conf := s.App.Config()

		raw, err := json.Marshal(&conf)

		if err != nil {
			slog.Error("Failed to marshal config", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(raw)
	})

	mux.HandleFunc("POST /api/v1/sound/{soundID}/enable", func(w http.ResponseWriter, r *http.Request) {
		var sid assets.SoundID = assets.SoundID(r.PathValue("soundID"))
		err := s.App.EnableSound(sid)
		if err != nil {
			var uerr *app.UnknownSoundErr
			if errors.As(err, &uerr) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			slog.Error("Failed to enable sound", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Write([]byte{})
	})
	mux.HandleFunc("POST /api/v1/sound/{soundID}/disable", func(w http.ResponseWriter, r *http.Request) {
		var sid assets.SoundID = assets.SoundID(r.PathValue("soundID"))
		err := s.App.DisableSound(sid)
		if err != nil {
			var uerr *app.UnknownSoundErr
			if errors.As(err, &uerr) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			slog.Error("Failed to disable sound", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Write([]byte{})
	})
	mux.HandleFunc("POST /api/v1/sound/{soundID}/volume/{volume}", func(w http.ResponseWriter, r *http.Request) {
		var sid assets.SoundID = assets.SoundID(r.PathValue("soundID"))

		v, err := strconv.ParseFloat(r.PathValue("volume"), 64)
		if err != nil {
			slog.Warn("Failed to parse volume", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		err = s.App.SetSoundVolume(sid, v)
		if err != nil {
			var uerr *app.UnknownSoundErr
			if errors.As(err, &uerr) {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}
			slog.Error("Failed to disable sound", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		w.Write([]byte{})
	})

	return mux
}
