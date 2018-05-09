//go:generate go-bindata -pkg=codetainer web/...
package codetainer

import (
	"fmt"
	"os"

	"github.com/recruit2class/codetainer/mlog"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const (
	// Name application name
	Name = "Codetainer"

	// Description
	Description = ""

	// Version application version number
	Version = "0.1.1"
)

var (
	// Build SHA
	Build string

	// TimeFormat global time format string
	TimeFormat = "15:04:05"

	app           = kingpin.New(Name, Description)
	debug         = app.Flag("debug", "Enable debug logging.").Short('v').Bool()
	dev           = app.Flag("dev", "Enable dev mode.").Bool()
	quiet         = app.Flag("quiet", "Remove all output logging.").Short('q').Bool()
	appSSL        = app.Flag("ssl", "Enable SSL (useful outside nginx/apache).").Short('s').Bool()
	appConfigPath = app.Flag("config", "Config path (default is ~/.codetainer/config.toml or /etc/codetainer/config.toml)").Short('c').String()

	server = app.Command("server", "Start the Codetainer API server.")

	profileCommand         = app.Command("profile", "Profile commands")
	profileListCommand     = profileCommand.Command("list", "List profiles")
	profileRegisterCommand = profileCommand.Command("register", "Register a profile")
	profileRegisterPath    = profileRegisterCommand.Arg("path", "Path to load of JSON profile").Required().String()
	profileRegisterName    = profileRegisterCommand.Arg("name", "name of profile").Required().String()

	imageCommand       = app.Command("image", "Image commands")
	registerCommand    = imageCommand.Command("register", "Register an image for use with codetainer")
	registerImageId    = registerCommand.Arg("image-id", "Docker image id").Required().String()
	registerCommandArg = registerCommand.Arg("command", "Default command to use to start container, e.g. /bin/bash").String()
	listImagesCommand  = imageCommand.Command("list", "List Images")

	codetainerCreate        = app.Command("create", "Launch a new codetainer")
	codetainerCreateImageId = codetainerCreate.Arg("image-id", "Docker image id").Required().String()
	codetainerCreateName    = codetainerCreate.Arg("name", "Name of container").String()
	codetainerRemove        = app.Command("remove", "Remove a codetainer")
	codetainerRemoveId      = codetainerRemove.Arg("id", "id to remove").Required().String()

	codetainerList = app.Command("list", "List all codetainers")

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
	app.Version(fmt.Sprintf("Version: %s Build: %s", Version, Build))
	args, perr := app.Parse(os.Args[1:])

	initLogger()

	config, err := NewConfig(*appConfigPath)

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

	case registerCommand.FullCommand():
		RegisterCodetainerImage(*registerImageId, *registerCommandArg)

	case codetainerCreate.FullCommand():
		CreateCodetainer(*codetainerCreateImageId, *codetainerCreateName)

	case codetainerList.FullCommand():
		CodetainerList()

	case codetainerRemove.FullCommand():
		CodetainerRemove(*codetainerRemoveId)

	case listImagesCommand.FullCommand():
		ListCodetainerImages()

	case profileListCommand.FullCommand():
		ListCodetainerProfiles()

	case profileRegisterCommand.FullCommand():
		RegisterCodetainerProfile(*profileRegisterPath, *profileRegisterName)

	default:
		app.Usage([]string{})
	}

}
