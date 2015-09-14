package codetainer

import (
	"fmt"
	"io/ioutil"
	"log"
)

func RegisterCodetainerProfile(pathToProfile string, name string) {

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		Log.Fatal(err)
	}

	var c CodetainerConfig
	prof, err := ioutil.ReadFile(pathToProfile)

	if err != nil {
		Log.Error("Unable to read file " + pathToProfile)
		Log.Fatal(err)
	}
	c.Profile = string(prof)
	c.Name = name

	err = c.Validate()
	if err != nil {
		Log.Error("Unable to parse " + pathToProfile)
		Log.Fatal(err)
	}
	err = c.Save(db)
	if err != nil {
		Log.Fatal(err)
	}
	log.Printf("Created profile with id=%s: \n", c.Id)
	log.Println("--")
	log.Printf(c.Profile)
}

func ListCodetainerProfiles() {

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		Log.Fatal(err)
	}
	var cl []CodetainerConfig = make([]CodetainerConfig, 0)

	err = db.engine.Find(&cl, &CodetainerConfig{})

	if err != nil {
		Log.Fatal("Unable to list profiles: ", err)
	}

	if len(cl) > 0 {
		fmt.Printf("Found %d profiles:\n", len(cl))
	} else {
		fmt.Println("No profiles found.")
	}

	for _, c := range cl {
		fmt.Printf("-- [%s] %s\n", c.Id, c.Name)
	}
}

func RegisterCodetainerImage(id string, command string) {

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		Log.Fatal(err)
	}

	image := CodetainerImage{Id: id, DefaultStartCommand: command}
	err = image.Register(db)

	if err != nil {
		Log.Fatal("Unable to register container image: ", err)
	}
	Log.Info("Registration succeeded.")
}

func CodetainerList() {
	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		Log.Fatal(err)
	}
	cl, err := db.ListCodetainers()

	if err != nil {
		Log.Fatal(err)
	}

	if len(*cl) > 0 {
		fmt.Printf("Found %d codetainers.\n", len(*cl))
	} else {
		fmt.Println("No codetainers found.")
	}

	for _, c := range *cl {
		fmt.Printf("-- [%s] %s", c.Id, c.Name)
		if c.Running {
			fmt.Printf(" (Running)")
		}
		fmt.Println()
	}
}

func CreateCodetainer(imageId string, name string) {

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		Log.Fatal(err)
	}

	c := Codetainer{ImageId: imageId, Name: name}
	err = c.Create(db)

	if err != nil {
		Log.Fatal("Unable to create codetainer: ", err)
	}
	fmt.Printf("Codetainer %s creation succeeded!\n", c.Name)
	fmt.Printf("You can interact with it here: "+GlobalConfig.Url()+"/api/v1/codetainer/%s/view\n", c.Id)
}
