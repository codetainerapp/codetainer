package codetainer

import "fmt"

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
	Log.Info("Create codetainer succeeded:", c)
}
