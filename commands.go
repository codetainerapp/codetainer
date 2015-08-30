package codetainer

func RegisterCodetainerImage(id string, command string) {

	db, err := GlobalConfig.GetDatabase()
	if err != nil {
		Log.Fatal(err)
	}
	err = db.RegisterCodetainerImage(id, command)

	if err != nil {
		Log.Fatal("Unable to register container image: ", err)
	}
	Log.Info("Registration succeeded.")
}
