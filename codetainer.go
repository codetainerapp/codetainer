package main

import (
	"os"
	"runtime"

	"github.com/codetainerapp/control/mlog"
	"gopkg.in/alecthomas/kingpin.v1"
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

	app    = kingpin.New(Name, Description)
	debug  = app.Flag("debug", "Enable debug logging.").Short('v').Bool()
	dev    = app.Flag("dev", "Enable dev mode.").Bool()
	quiet  = app.Flag("quiet", "Remove all output logging.").Short('q').Bool()
	appSSL = app.Flag("ssl", "Enable SSL (usefull outside nginx/apache).").Short('s').Bool()

	server = app.Command("server", "Start the Codetainer control server.")

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

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	app.Version(Version)
	args, perr := app.Parse(os.Args[1:])

	initLogger()

	switch kingpin.MustParse(args, perr) {

	case server.FullCommand():

	default:
		app.Usage(os.Stdout)
	}

}
