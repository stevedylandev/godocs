package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Goal: have a fuzy find tui or cli to find go docs
// We want the ability to search the content, not just for the package name
// So many times I want to find the package for a specific functionality
//
// To do this we probably need to do the following
// - List all docs for all packages
//   - We can use `go list all` to find all top levle packages
// - cat the contents of each doc using go doc -all <package>
// - Save contents to cache directory
// - Use that directory for searching, e.g. ripgrep for a word or phrase
//

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("Usage: godocs command")
		os.Exit(1)
	}

	switch args[0] {
	case "update":
		updateCache()
	case "search":
		search()
	default:
		fmt.Println("Unknown command:", args[0])
		os.Exit(1)
	}
}

func updateCache() {
	// Find home directory and make cache
	cacheBase, err := os.UserCacheDir()
	if err != nil {
		fmt.Printf("Error finding cache directory: %s\n", err)
		os.Exit(1)
	}
	dir := filepath.Join(cacheBase, "godoc")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Printf("Error creating cache directory: %s\n", err)
		os.Exit(1)
	}
	// List all packages
	out, err := exec.Command("go", "list", "all").Output()
	if err != nil {
		fmt.Printf("Error listing all packages: %s", err)
		os.Exit(1)
	}

	// Split list into array using carriage return
	libs := strings.Split(string(out), "\n")
	// Loop over array to save each doc to file
	for _, lib := range libs {
		lib = strings.TrimSpace(lib)
		if lib == "" {
			continue
		}
		out, err := exec.Command("go", "doc", "-all", lib).Output()
		if err != nil {
			fmt.Printf("Error getting doc: %s", err)
			continue
		}
		name := strings.ReplaceAll(lib, "/", "-")
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, out, 0o644); err != nil {
			fmt.Printf("Error writing doc: %s\n", err)
			continue
		}
	}
	fmt.Println("✓ Cache Updated")
}

func search() {
	for _, bin := range []string{"fzf", "rg"} {
		if _, err := exec.LookPath(bin); err != nil {
			fmt.Fprintf(os.Stderr, "%s not found in PATH\n", bin)
			os.Exit(1)
		}
	}

	reload := `test -n {q} && rg --line-number --no-heading --color=always --smart-case -- {q} || true`

	cmd := exec.Command("fzf",
		"--disabled", "--ansi",
		"--delimiter", ":",
		"--bind", "change:reload:"+reload,
		"--preview", "bat --color=always --highlight-line {2} {1} 2>/dev/null || cat {1}",
		"--preview-window", "right:60%,+{2}/2",
		"--prompt", "godocs> ",
	)
	cacheBase, err := os.UserCacheDir()
	if err != nil {
		fmt.Printf("Error finding cache directory: %s\n", err)
		os.Exit(1)
	}
	cmd.Dir = filepath.Join(cacheBase, "godoc")
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	out, err := cmd.Output()
	if err != nil {
		fmt.Println("fzf error: %s\n", err)
		os.Exit(0)
	}

	selection := strings.TrimSpace(string(out))

	parts := strings.SplitN(selection, ":", 3)
	if len(parts) < 3 {
		fmt.Println("error with fzf output:", selection)
		os.Exit(1)
	}

	targetFilePath := filepath.Join(cacheBase, "godoc", parts[0])
	targetLine := parts[1]

	editor := os.Getenv("EDITOR")
	viewer := exec.Command(editor, targetFilePath, "+"+targetLine)
	viewer.Stdin = os.Stdin
	viewer.Stdout = os.Stdout
	viewer.Stderr = os.Stderr
	if err := viewer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "viewer error: %s\n", err)
		os.Exit(1)
	}
}
