package web

import (
	"errors"
	"fmt"
	"net/http"
	"playtime/storage"
	"strconv"

	"github.com/flosch/pongo2/v6"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func (s *Server) settings(c echo.Context) error {
	context := c.(*PlaytimeContext)

	return c.Render(http.StatusOK, "settings", pongo2.Context{
		"_csrf_token": c.Get("csrf"),
		"user":        context.user,
		"settings":    context.settings,
		"platforms":   sortedPlatforms(),
	})
}

func (s *Server) settingsGeneralForm(c echo.Context) error {
	context := c.(*PlaytimeContext)

	return c.Render(http.StatusOK, "settings_general", pongo2.Context{
		"_csrf_token": c.Get("csrf"),
		"user":        context.user,
		"settings":    context.settings,
		"done":        c.QueryParam("done"),
		"languages":   storage.Languages,
	})
}

func (s *Server) settingsGeneralSubmit(c echo.Context) error {
	context := c.(*PlaytimeContext)

	log.Infof("generalSettingsSubmit for %s", context.user.Login)

	settings := context.settings
	settings.Language = c.FormValue("language")

	autoSaveInterval, err := strconv.Atoi(c.FormValue("auto-save-interval"))
	if err != nil {
		return err
	}
	autoSaveCapacity, err := strconv.Atoi(c.FormValue("auto-save-capacity"))
	if err != nil {
		return err
	}

	settings.DefaultGameSettings.NetplayEnabled = c.FormValue("netplay-enabled") == "1"
	settings.DefaultGameSettings.NetplayRequireLogin = c.FormValue("netplay-require-login") == "1"
	settings.DefaultGameSettings.NetplayOpen = c.FormValue("netplay-open") == "1"
	settings.DefaultGameSettings.AutoSaveEnabled = c.FormValue("auto-save-enabled") == "1"
	settings.DefaultGameSettings.AutoSaveInterval = autoSaveInterval * 60
	settings.DefaultGameSettings.AutoSaveCapacity = autoSaveCapacity
	settings.DefaultGameSettings.CueEnabled = c.FormValue("cue-enabled") == "1"

	if _, err := s.storage.SettingsSave(*settings); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/settings/general?done=1")
}

func (s *Server) settingsByPlatformForm(c echo.Context) error {
	context := c.(*PlaytimeContext)
	platform := c.Param("platform")

	if _, ok := storage.Platforms[platform]; !ok {
		return errors.New("platform not found")
	}

	settings := context.settings
	platformSettings, ok := settings.EmulatorSettings[platform]
	if !ok {
		platformSettings = storage.DefaultEmulatorSettings(platform)
	}

	return c.Render(http.StatusOK, "settings_platform", pongo2.Context{
		"_csrf_token":          c.Get("csrf"),
		"user":                 context.user,
		"settings":             platformSettings,
		"shaders":              storage.Shaders,
		"platform":             storage.Platforms[platform],
		"bioses":               storage.Platforms[platform].Bios,
		"cores":                storage.Platforms[platform].Cores,
		"core_options":         storage.CoreOptionsByPlatform(platform),
		"fast_forward_ratios":  storage.FastForwardRatios,
		"slow_motion_ratios":   storage.SlowMotionRatios,
		"rewind_granularities": storage.RewindGranularity,
	})
}

func (s *Server) settingsByPlatformSubmit(c echo.Context) error {
	context := c.(*PlaytimeContext)
	platform := c.Param("platform")

	if _, ok := storage.Platforms[platform]; !ok {
		return errors.New("platform not found")
	}

	log.Infof("settingsByPlatformSubmit %s for %s", platform, context.user.Login)

	settings := context.settings

	settings.EmulatorSettings[platform] = settingsCollectFormData(c)

	if _, err := s.storage.SettingsSave(*settings); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/settings?done=1")
}

func (s *Server) settingsByPlatformRestoreDefaults(c echo.Context) error {
	context := c.(*PlaytimeContext)
	platform := c.Param("platform")

	if _, ok := storage.Platforms[platform]; !ok {
		return errors.New("platform not found")
	}

	return c.Render(http.StatusOK, "settings_platform_restore", pongo2.Context{
		"_csrf_token": c.Get("csrf"),
		"user":        context.user,
		"platform":    storage.Platforms[platform],
	})
}

func (s *Server) settingsByPlatformRestoreDefaultsSave(c echo.Context) error {
	context := c.(*PlaytimeContext)
	platform := c.Param("platform")

	if _, ok := storage.Platforms[platform]; !ok {
		return errors.New("platform not found")
	}

	settings, err := s.storage.SettingsGetByUserId(context.user.Id)
	if err != nil {
		return err
	}

	settings.EmulatorSettings[platform] = storage.DefaultEmulatorSettings(platform)

	if _, err := s.storage.SettingsSave(settings); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/settings/"+platform)
}

///////////////////////////////////////////////////////////////////////////////

func settingsCollectFormData(c echo.Context) storage.EmulatorSettings {
	cacheLimit, err := strconv.ParseInt(c.FormValue("cache-limit"), 10, 64)
	if err != nil {
		log.Warnf("unable to read cache limit: %s", err)
		cacheLimit = storage.DefaultCacheLimit
	}

	volume, err := strconv.ParseFloat(c.FormValue("volume"), 32)
	if err != nil {
		log.Warnf("unable to read volume: %s", err)
		volume = storage.DefaultVolume
	}

	videoRotation, err := strconv.ParseInt(c.FormValue("video-rotation"), 10, 64)
	if err != nil {
		log.Warnf("unable to read video rotation: %s", err)
		videoRotation = 0
	}
	if videoRotation < 0 || videoRotation > 3 {
		log.Warnf("wrong video rotation value: %d", videoRotation)
		videoRotation = 0
	}

	buttons := storage.EmulatorButtons{
		PlayPause:     c.FormValue("button-play-pause") == "1",
		Restart:       c.FormValue("button-restart") == "1",
		Mute:          c.FormValue("button-mute") == "1",
		Settings:      c.FormValue("button-settings") == "1",
		FullScreen:    c.FormValue("button-full-screen") == "1",
		SaveState:     c.FormValue("button-save-state") == "1",
		LoadState:     c.FormValue("button-load-state") == "1",
		ScreenRecord:  c.FormValue("button-screen-record") == "1",
		Gamepad:       c.FormValue("button-gamepad") == "1",
		Cheat:         c.FormValue("button-cheat") == "1",
		Volume:        c.FormValue("button-volume") == "1",
		SaveSavFiles:  c.FormValue("button-save-sav-files") == "1",
		LoadSavFiles:  c.FormValue("button-load-sav-files") == "1",
		QuickSave:     c.FormValue("button-quick-save") == "1",
		QuickLoad:     c.FormValue("button-quick-load") == "1",
		Screenshot:    c.FormValue("button-screenshot") == "1",
		CacheManager:  c.FormValue("button-cache-manager") == "1",
		ContextMenu:   c.FormValue("button-context-menu") == "1",
		RightClick:    c.FormValue("button-right-click") == "1",
		ExitEmulation: c.FormValue("button-exit-emulation") == "1",
	}

	shader := c.FormValue("shader")
	shaderFound := false
	for _, item := range storage.Shaders {
		if item.Value == shader {
			shaderFound = true
		}
	}
	if !shaderFound {
		log.Warnf("wrong shader value: %s", shader)
		shader = storage.Shaders[0].Value
	}

	var core = c.FormValue("core")
	var coreOptions = storage.CoreOptionsByCore(core)
	coreOptionsValues := make(map[string]string)
	for _, option := range coreOptions {
		optionValue := c.FormValue(option.Id)
		if len(optionValue) != 0 {
			coreOptionsValues[option.Id] = optionValue
		}
	}

	settings := storage.EmulatorSettings{
		Core:                   core,
		Bios:                   c.FormValue("bios"),
		ColorScheme:            c.FormValue("color-scheme"),
		ColorBackground:        c.FormValue("color-background"),
		CacheLimit:             cacheLimit,
		Volume:                 volume,
		FastForwardRatio:       c.FormValue("ff-ratio"),
		SlowMotionRatio:        c.FormValue("sm-ratio"),
		RewindGranularity:      c.FormValue("rewind-granularity"),
		Shader:                 shader,
		FPS:                    c.FormValue("fps") == "1",
		VirtualGamepadLeftHand: c.FormValue("virtual-gamepad-left-hand") == "1",
		StartFullScreen:        c.FormValue("start-full-screen") == "1",
		FastForwardMode:        c.FormValue("fast-forward-mode") == "1",
		SlowMotionMode:         c.FormValue("slow-motion-mode") == "1",
		Rewind:                 c.FormValue("rewind-enabled") == "1",
		Threads:                c.FormValue("threads") == "1",
		DisableBrowserDB:       c.FormValue("disable-browser-db") == "1",
		DisableVSync:           c.FormValue("disable-vsync") == "1",
		VideoRotation:          videoRotation,
		Buttons:                buttons,
		Controls:               settingsCollectControls(c),
		CoreOptions:            coreOptionsValues,
	}

	return settings
}

func settingsCollectControls(c echo.Context) [4]storage.EmulatorControls {
	controls := [4]storage.EmulatorControls{}

	for _, player := range []int{0, 1, 2, 3} {
		for _, input := range []string{"keyboard", "gamepad"} {
			mapping := storage.EmulatorControlsMapping{
				B:               settingsReadControlButton(c, input, player, "b"),
				Y:               settingsReadControlButton(c, input, player, "y"),
				Select:          settingsReadControlButton(c, input, player, "select"),
				Start:           settingsReadControlButton(c, input, player, "start"),
				Up:              settingsReadControlButton(c, input, player, "up"),
				Down:            settingsReadControlButton(c, input, player, "down"),
				Left:            settingsReadControlButton(c, input, player, "left"),
				Right:           settingsReadControlButton(c, input, player, "right"),
				A:               settingsReadControlButton(c, input, player, "a"),
				X:               settingsReadControlButton(c, input, player, "x"),
				L:               settingsReadControlButton(c, input, player, "l"),
				R:               settingsReadControlButton(c, input, player, "r"),
				L2:              settingsReadControlButton(c, input, player, "l2"),
				R2:              settingsReadControlButton(c, input, player, "r2"),
				L3:              settingsReadControlButton(c, input, player, "l3"),
				R3:              settingsReadControlButton(c, input, player, "r3"),
				LStickUp:        settingsReadControlButton(c, input, player, "l-stick-up"),
				LStickDown:      settingsReadControlButton(c, input, player, "l-stick-down"),
				LStickLeft:      settingsReadControlButton(c, input, player, "l-stick-left"),
				LStickRight:     settingsReadControlButton(c, input, player, "l-stick-right"),
				RStickUp:        settingsReadControlButton(c, input, player, "r-stick-up"),
				RStickDown:      settingsReadControlButton(c, input, player, "r-stick-down"),
				RStickLeft:      settingsReadControlButton(c, input, player, "r-stick-left"),
				RStickRight:     settingsReadControlButton(c, input, player, "r-stick-right"),
				QuickSaveState:  settingsReadControlButton(c, input, player, "quick-save-state"),
				QuickLoadState:  settingsReadControlButton(c, input, player, "quick-load-state"),
				ChangeStateSlot: settingsReadControlButton(c, input, player, "change-state-slot"),
				FastForward:     settingsReadControlButton(c, input, player, "fast-forward"),
				SlowMotion:      settingsReadControlButton(c, input, player, "slow-motion"),
				Rewind:          settingsReadControlButton(c, input, player, "rewind"),
			}
			switch {
			case input == "keyboard":
				controls[player].Keyboard = mapping
			case input == "gamepad":
				controls[player].Gamepad = mapping
			}
		}
	}

	return controls
}

func settingsReadControlButton(c echo.Context, input string, player int, button string) string {
	key := fmt.Sprintf("control-%s-%d-%s", input, player, button)
	return c.FormValue(key)
}
