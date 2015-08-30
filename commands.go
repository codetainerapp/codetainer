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
