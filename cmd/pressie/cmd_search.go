package main

// cmdSearch searches contacts via configured plugins.
// TODO: implement
func cmdSearch(args []string) {
	if wantsHelp(args) {
		helpSearch()
		return
	}
	println("search: not yet implemented")
}