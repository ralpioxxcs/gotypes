package app

// Run main app
func Run() {
	app := NewApp()

	// Run tui main
	if err := app.Run(); err != nil {
		panic(err)
	}
}
