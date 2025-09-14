package ui

import (
	"github.com/getlantern/systray"
	"github.com/pavelpuchok/gomoodist/internal/ui/icons"
)

type Systray struct {
	OpenSettings func()
	TogglePlay   func() bool
}

func (s *Systray) Run() {
	systray.Run(s.onReady, s.onExit)
}

func (s *Systray) onReady() {
	systray.SetIcon(icons.SystrayIcon)
	systray.SetTitle("GoMoodist")
	systray.SetTooltip("Pause")

	togglePlay := systray.AddMenuItem("Play/Pause", "")
	settingsMenuItem := systray.AddMenuItem("Settings", "Open settings")
	quitMenuItem := systray.AddMenuItem("Quit", "Quit")

	go func() {
		for {
			select {
			case <-quitMenuItem.ClickedCh:
				systray.Quit()
			case <-settingsMenuItem.ClickedCh:
				s.OpenSettings()
			case <-togglePlay.ClickedCh:
				if s.TogglePlay() {
					systray.SetIcon(icons.SystrayIcon)
				} else {
					systray.SetIcon(icons.SystrayIconOff)
				}
			}
		}
	}()
}

func (s *Systray) onExit() {

}
