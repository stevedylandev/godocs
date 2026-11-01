package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Goal: have a fuzy find tui or cli to find go docs
// We want the ability to search the content, not just for the package name
// So many times I want to find the package for a specific functionality
//
// To do this we probably need to do the following
// - List all docs for all packages
//   - We can use `go list all` to find all top levle packages
//   - Challenge is that there's some general info for each package,
//     then there's methods that you have to call individually.
//     Trying to traverse them could be tricky, but I think doable
//     if we use `func` as the prefix.
//   - Should actually be the following key words
//     - func
//     - type
//     - const
//     - var
// - cat the contents of each doc
// - Save contents to cache directory
// - Use that directory for searching, e.g. ripgrep for a word or phrase
//

func main() {
	// List all packages
	out, err := exec.Command("go", "list", "all").Output()
	if err != nil {
		fmt.Printf("Error: %s", err)
		return
	}

	// Split list into array using carriage return
	libs := strings.Split(string(out), "\n")
	// Loop over array to save each doc to file
	for _, lib := range libs {
		// fmt.Printf("%v: %s\n", i, lib)
		out, err := exec.Command("go", "doc", lib).Output()
		if err != nil {
			fmt.Printf("Error: %s", err)
			return
		}
		filename := fmt.Sprintf("%s.txt", lib)
		if err := os.WriteFile(filename, out, 0644); err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}
	}
}
