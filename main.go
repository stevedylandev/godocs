package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"slices"
	"flag"
)

// Goal: have a fuzy find tui or cli to find go docs
// We want the ability to search the content, not just for the package name
// So many times I want to find the package for a specific functionality
//
// MVP: using exec.Command to run go list -all with go doc <package>,
// then ripgrep through fzf 
//
// V1: use go/build, go/ast, go/doc. go/parser, go/printer to generate
// raw doc values to build more controled and robust views. 
// - Syntax highlighting
// - Better indexing for search
// - Full TUI for navigating and accessing
// - CLI for printing to stdout

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: gobble command")
        os.Exit(1)
    }

    editor := os.Getenv("EDITOR")

    switch os.Args[1] {
    case "update":
        updateCache()
    case "search":
        fs := flag.NewFlagSet("search", flag.ExitOnError)
        editorFlag := fs.String("e", editor, "Open the docs with provided editor")
        fs.Parse(os.Args[2:])
        search(*editorFlag)
    default:
        fmt.Println("Unknown command:", os.Args[1])
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
	dir := filepath.Join(cacheBase, "gobble")
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

	stdLibsOut, err := exec.Command("go", "list", "std").Output()
	if err != nil {
		fmt.Printf("Error listing all packages: %s", err)
		os.Exit(1)
	}
	stdLibs := strings.Split(string(stdLibsOut), "\n")

	allLibs := slices.Concat(libs, stdLibs)

	for i, lib := range allLibs {
		fmt.Printf("Processed %v of %v\r", i, len(allLibs))
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
	fmt.Println("\n✓ Cache Updated")
}

func search(editor string) {
	for _, bin := range []string{"fzf", "rg"} {
		if _, err := exec.LookPath(bin); err != nil {
			fmt.Fprintf(os.Stderr, "%s not found in PATH\n", bin)
			os.Exit(1)
		}
	}

	cmd := exec.Command("fzf",
		"--disabled", "--ansi",
		"--delimiter", ":",
		"--bind", "change:reload:test -n {q} && rg --line-number --no-heading --color=always --smart-case -- {q} || true",
		"--preview", "bat --color=always --highlight-line {2} {1} 2>/dev/null || cat {1}",
		"--preview-window", "right:60%,+{2}/2",
		"--prompt", "gobble> ",
	)
	cacheBase, err := os.UserCacheDir()
	if err != nil {
		fmt.Printf("Error finding cache directory: %s\n", err)
		os.Exit(1)
	}
	cmd.Dir = filepath.Join(cacheBase, "gobble")
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

	targetFilePath := filepath.Join(cacheBase, "gobble", parts[0])
	targetLine := parts[1]

	viewer := exec.Command(editor, targetFilePath, "+"+targetLine)
	viewer.Stdin = os.Stdin
	viewer.Stdout = os.Stdout
	viewer.Stderr = os.Stderr
	if err := viewer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "viewer error: %s\n", err)
		os.Exit(1)
	}
}
