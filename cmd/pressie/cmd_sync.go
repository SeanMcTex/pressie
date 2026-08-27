package main

// cmdSync does git pull + push on the gifts directory.
// TODO: implement
func cmdSync(args []string) {
	if wantsHelp(args) {
		helpSync()
		return
	}
	println("sync: not yet implemented")
}