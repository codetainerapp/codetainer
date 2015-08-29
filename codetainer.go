package codetainer

import (
	"os"

	"github.com/codetainerapp/codetainer/mlog"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
)

const (
	// Name application name
	Name = "Codetainer"

	// Description
	Description = "--"

	// Version application version number
	Version = "0.1.0"
)

var (
	// TimeFormat global time format string
	TimeFormat = "15:04:05"

	app        = kingpin.New(Name, Description)
	debug      = app.Flag("debug", "Enable debug logging.").Short('v').Bool()
	dev        = app.Flag("dev", "Enable dev mode.").Bool()
	quiet      = app.Flag("quiet", "Remove all output logging.").Short('q').Bool()
	appSSL     = app.Flag("ssl", "Enable SSL (useful outside nginx/apache).").Short('s').Bool()
	configPath = app.Flag("config", "Config path (default is config.toml)").Short('c').String()

	server = app.Command("server", "Start the Codetainer control server.")

	image           = app.Command("image", "Image commands")
	register        = image.Command("register", "Register an image for use with codetainer")
	registerImageId = register.Arg("image-id", "Docker image id").Required().String()
	registerCommand = register.Arg("command", "Default command to use to start container, e.g. /bin/bash").String()

	// Log Global logger
	Log *mlog.Logger

	// DevMode Development mode switch. If true
	// debug logging and serving assets from disk
	// is enabled.
	DevMode bool

	// TestMode
	TestMode bool
)

func initLogger() {
	Log = mlog.New()

	Log.Prefix = Name

	if *debug {
		Log.SetLevel(mlog.DebugLevel)
	} else {
		Log.SetLevel(mlog.InfoLevel)
	}

	if *dev {
		DevMode = true
		Log.SetLevel(mlog.DebugLevel)
		Log.Info("DEBUG MODE ENABLED.")
	} else {
		DevMode = false
	}

	if *quiet {
		Log.SetLevel(mlog.FatalLevel)
	}

}

func Start() {
	app.Version(Version)
	args, perr := app.Parse(os.Args[1:])

	initLogger()

	config, err := NewConfig(*configPath)

	if err != nil {
		Log.Fatal(err)
	} else {
		GlobalConfig = *config
	}

	if !config.TestConfig() {
		Log.Fatal("Invalid configuration detected.")
	}

	switch kingpin.MustParse(args, perr) {

	case server.FullCommand():
		StartServer()

	case register.FullCommand():
		RegisterCodetainerImage(*registerImageId, *registerCommand)

	default:
		app.Usage(os.Stdout)
	}

}
