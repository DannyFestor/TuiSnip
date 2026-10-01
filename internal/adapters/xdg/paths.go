package xdg

import (
	"errors"
	"path/filepath"
)

const (
	homeVariable       = "HOME"
	configHomeVariable = "XDG_CONFIG_HOME"
	dataHomeVariable   = "XDG_DATA_HOME"
	stateHomeVariable  = "XDG_STATE_HOME"

	configHomeFallback = ".config"
	dataHomeFallback   = ".local/share"
	stateHomeFallback  = ".local/state"

	appDirName       = "tuisnip"
	configFileName   = "config.toml"
	databaseFileName = "tuisnip.db"
	stateFileName    = "state.toml"
	logFileName      = "tuisnip.log"
)

var ErrNoHome = errors.New("xdg: HOME is not set to an absolute path")

type Paths struct {
	ConfigFile   string
	DatabaseFile string
	StateFile    string
	LogFile      string
}

func Resolve(getenv func(key string) string) (Paths, error) {
	home := getenv(homeVariable)
	if !filepath.IsAbs(home) {
		return Paths{}, ErrNoHome
	}

	dirs := baseDirs{getenv: getenv, home: home}
	stateDir := dirs.appDir(stateHomeVariable, stateHomeFallback)

	return Paths{
		ConfigFile:   filepath.Join(dirs.appDir(configHomeVariable, configHomeFallback), configFileName),
		DatabaseFile: filepath.Join(dirs.appDir(dataHomeVariable, dataHomeFallback), databaseFileName),
		StateFile:    filepath.Join(stateDir, stateFileName),
		LogFile:      filepath.Join(stateDir, logFileName),
	}, nil
}

type baseDirs struct {
	getenv func(key string) string
	home   string
}

func (d baseDirs) appDir(variable, homeFallback string) string {
	base := d.getenv(variable)
	if !filepath.IsAbs(base) {
		base = filepath.Join(d.home, homeFallback)
	}

	return filepath.Join(base, appDirName)
}
