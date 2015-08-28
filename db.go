package codetainer

import (
	"errors"
	"runtime"
	"time"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

var (
	//
	// Default command to start in a container
	//
	DefaultExecCommand string = "/bin/bash"
)

type CodetainerImage struct {
	Id                  string
	Tags                []string
	DefaultStartCommand string
	Description         string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	Enabled             bool
}

type Codetainer struct {
	Id        string
	ImageId   string
	Defunct   bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Database struct {
	engine *xorm.Engine
}

func CloseDb(db *Database) {
	db.engine.Close()
}

func NewDatabase(dbPath string) (*Database, error) {
	db := &Database{}
	engine, err := xorm.NewEngine("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if !fileExists(dbPath) {
		engine.Sync(new(Codetainer), new(CodetainerImage))
	}

	db.engine = engine
	runtime.SetFinalizer(db, CloseDb)
	return db, nil
}

func (db *Database) ListCodetainerImages() (*[]CodetainerImage, error) {
	var containerImages []CodetainerImage = make([]CodetainerImage, 0)
	var doneImages []CodetainerImage = make([]CodetainerImage, 0)

	client, err := GlobalConfig.GetDockerClient()

	if err != nil {
		return nil, err
	}
	opts := docker.ListImagesOptions{All: true}
	dockerImages, err := client.ListImages(opts)
	err = db.engine.Find(&containerImages, &CodetainerImage{Enabled: true})
	if err != nil {
		return nil, err
	}

	// filter codetainer images by stuff in docker.
	for _, img := range containerImages {
		if findDockerImageInList(img.Id, dockerImages) {
			doneImages = append(doneImages, img)
		}
	}

	return &doneImages, nil
}

func findDockerImageInList(id string, dockerImages []docker.APIImages) bool {
	for _, img := range dockerImages {
		if img.ID == id {
			return true
		}
	}
	return false
}

func imageExistsInDocker(id string) bool {
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		Log.Error("Unable to fetch docker client", err)
		return false
	}

	filter := map[string][]string{"Id": []string{id}}
	opts := docker.ListImagesOptions{
		Filters: filter,
	}
	imgs, err := client.ListImages(opts)
	if err != nil {
		Log.Error("Unable to fetch image", err)
		return false
	}
	if imgs == nil || len(imgs) == 0 {
		return false
	}
	return true
}

//
// Register a new codetainer image.
//
func (db *Database) RegisterCodetainerImage(id string, command string) error {
	// check if image is in docker
	if imageExistsInDocker(id) {
		if command == "" {
			command = DefaultExecCommand
		}
		image := CodetainerImage{Id: id, DefaultStartCommand: command}
		_, err := db.engine.Insert(&image)
		return err

	} else {
		return errors.New("No image found in docker.")
	}
	return nil
}

//
// List all running codetainers
//
func (db *Database) ListCodetainers() {

}
