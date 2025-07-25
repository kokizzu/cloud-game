package libretro

import (
	"github.com/giongto35/cloud-game/v3/pkg/config"
	"github.com/giongto35/cloud-game/v3/pkg/games"
	"github.com/giongto35/cloud-game/v3/pkg/logger"
	"github.com/giongto35/cloud-game/v3/pkg/worker/caged/libretro/manager"
	"github.com/giongto35/cloud-game/v3/pkg/worker/cloud"
)

type Caged struct {
	Emulator

	base *Frontend // maintains the root for mad embedding
	conf CagedConf
	log  *logger.Logger
}

type CagedConf struct {
	Emulator  config.Emulator
	Recording config.Recording
}

func (c *Caged) Name() string { return "libretro" }

func Cage(conf CagedConf, log *logger.Logger) Caged {
	return Caged{conf: conf, log: log}
}

func (c *Caged) Init() error {
	if err := manager.CheckCores(c.conf.Emulator, c.log); err != nil {
		c.log.Warn().Err(err).Msgf("a Libretro cores sync fail")
	}

	if c.conf.Emulator.FailFast {
		if err := c.IsSupported(); err != nil {
			return err
		}
	}

	return nil
}

func (c *Caged) ReloadFrontend() {
	frontend, err := NewFrontend(c.conf.Emulator, c.log)
	if err != nil {
		c.log.Fatal().Err(err).Send()
		return
	}
	c.Emulator = frontend
	c.base = frontend
}

// VideoChangeCb adds a callback when video params are changed by the app.
func (c *Caged) VideoChangeCb(fn func()) { c.base.SetVideoChangeCb(fn) }

func (c *Caged) Load(game games.GameMetadata, path string) error {
	if c.Emulator == nil {
		return nil
	}
	c.Emulator.LoadCore(game.System)
	if err := c.Emulator.LoadGame(game.FullPath(path)); err != nil {
		return err
	}
	c.ViewportRecalculate()
	return nil
}

func (c *Caged) EnableRecording(nowait bool, user string, game string) {
	if c.conf.Recording.Enabled {
		// !to fix races with canvas pool when recording
		c.base.DisableCanvasPool = true
		c.Emulator = WithRecording(c.Emulator, nowait, user, game, c.conf.Recording, c.log)
	}
}

func (c *Caged) EnableCloudStorage(uid string, storage cloud.Storage) {
	if storage == nil {
		return
	}
	if wc, err := WithCloud(c.Emulator, uid, storage); err == nil {
		c.Emulator = wc
		c.log.Info().Msgf("cloud storage has been initialized")
	} else {
		c.log.Error().Err(err).Msgf("couldn't init cloud storage")
	}
}

func (c *Caged) AspectEnabled() bool              { return c.base.nano.Aspect }
func (c *Caged) AspectRatio() float32             { return c.base.AspectRatio() }
func (c *Caged) PixFormat() uint32                { return c.Emulator.PixFormat() }
func (c *Caged) Rotation() uint                   { return c.Emulator.Rotation() }
func (c *Caged) AudioSampleRate() int             { return c.Emulator.AudioSampleRate() }
func (c *Caged) ViewportSize() (int, int)         { return c.base.ViewportSize() }
func (c *Caged) Scale() float64                   { return c.Emulator.Scale() }
func (c *Caged) Input(p int, d byte, data []byte) { c.base.Input(p, d, data) }
func (c *Caged) KbMouseSupport() bool             { return c.base.KbMouseSupport() }
func (c *Caged) Start()                           { go c.Emulator.Start() }
func (c *Caged) SetSaveOnClose(v bool)            { c.base.SaveOnClose = v }
func (c *Caged) SetSessionId(name string)         { c.base.SetSessionId(name) }
func (c *Caged) Close()                           { c.Emulator.Close() }
func (c *Caged) IsSupported() error               { return c.base.IsSupported() }
