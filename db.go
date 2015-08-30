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
	Id                  string `xorm:"varchar(128) not null unique"`
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
	opts := docker.ListImagesOptions{}
	dockerImages, err := client.ListImages(opts)
	err = db.engine.Find(&containerImages, &CodetainerImage{Enabled: true})
	if err != nil {
		return nil, err
	}

	// filter codetainer images by stuff in docker.
	for _, img := range containerImages {
		if findDockerImageInList(img.Id, dockerImages) != nil {
			doneImages = append(doneImages, img)
		}
	}

	return &doneImages, nil
}

func findDockerImageInList(id string, dockerImages []docker.APIImages) *docker.APIImages {
	for _, img := range dockerImages {
		if img.ID == id {
			return &img
		}
		for _, tag := range img.RepoTags {
			if tag == id {
				return &img
			}
		}
	}
	return nil
}

func lookupImageInDocker(id string) *docker.APIImages {
	client, err := GlobalConfig.GetDockerClient()
	if err != nil {
		Log.Error("Unable to fetch docker client", err)
		return nil
	}

	opts := docker.ListImagesOptions{}
	imgs, err := client.ListImages(opts)
	if err != nil {
		Log.Error("Unable to fetch image", err)
		return nil
	}
	return findDockerImageInList(id, imgs)
}

//
// Register a new codetainer image.
//
func (db *Database) RegisterCodetainerImage(id string, command string) error {
	// check if image is in docker
	image := lookupImageInDocker(id)
	if image != nil {
		if command == "" {
			command = DefaultExecCommand
		}
		image := CodetainerImage{Id: image.ID, DefaultStartCommand: command, Enabled: true}
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
func (db *Database) LookupCodetainerImage(id string) (*CodetainerImage, error) {
	img := CodetainerImage{Id: id}
	has, err := db.engine.Get(&img)

	if has && err == nil {
		return &img, nil
	} else {
		return nil, err
	}
}

func (db *Database) SaveCodetainer(id string, imageId string) (*Codetainer, error) {
	c := Codetainer{Id: id, ImageId: imageId, Defunct: false}
	_, err := db.engine.Insert(&c)
	if err != nil {
		return nil, err
	}
	return &c, err
}

//
// List all running codetainers
//
func (db *Database) ListCodetainers() {

}
