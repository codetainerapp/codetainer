package codetainer

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
