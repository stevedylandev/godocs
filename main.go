package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type SearchResp struct {
	Items []SearchItem `json:"items"`
	Total int          `json:"total"`
}

type SearchItem struct {
	PackagePath string `json:"packagePath"`
	ModulePath  string `json:"modulePath"`
	Version     string `json:"version"`
	Synopsis    string `json:"synopsis"`
}

type Package struct {
	ModulePath        string   `json:"modulePath"`
	Version           string   `json:"version"`
	IsLatest          bool     `json:"isLatest"`
	IsStandardLibrary bool     `json:"isStandardLibrary"`
	GOOS              string   `json:"goos"`
	GOARCH            string   `json:"goarch"`
	Docs              string   `json:"docs,omitempty"`
	Imports           []string `json:"imports,omitempty"`
	PackageInfo
}

type PackageInfo struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Synopsis string `json:"synopsis"`
	// IsRedistributable is whether the license allows distribution.
	IsRedistributable bool `json:"isRedistributable"`
}

func main() {
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("Usage: godocs command")
		os.Exit(1)
	}

	switch args[0] {
	case "search":
		search(args[1])
	case "define":
		define(args[1])
	case "symbols":
		symbol(args[1])
	default:
		fmt.Println("Unknown command:", args[0])
		os.Exit(1)
	}
}

func search(query string) {
	url := fmt.Sprintf("http://localhost:5678/v1/search?q=%s&limit=100", query)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %s", err)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing body: %s", err)
		os.Exit(1)
	}

	var docsData SearchResp
	if err := json.Unmarshal(body, &docsData); err != nil {
		fmt.Printf("Error unmarshalling json: %s", err)
		os.Exit(1)
	}

	for _, doc := range docsData.Items {
		fmt.Println(doc.PackagePath)
	}
}

func define(packagePath string) {
	url := fmt.Sprintf("http://localhost:5678/v1beta/package/%s?doc=md&examples=true", packagePath)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %s", err)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing body: %s", err)
		os.Exit(1)
	}

	fmt.Println(string(body))

	// var packageData Package
	// if err := json.Unmarshal(body, &packageData); err != nil {
	// 	fmt.Printf("Error unmarshalling json: %s", err)
	// 	os.Exit(1)
	// }
	//
	// fmt.Println(packageData)
}

func symbol(packagePath string) {
	url := fmt.Sprintf("http://localhost:5678/v1beta/symbols/%s", packagePath)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making request: %s", err)
		os.Exit(1)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing body: %s", err)
		os.Exit(1)
	}

	fmt.Println(string(body))

	// var packageData Package
	// if err := json.Unmarshal(body, &packageData); err != nil {
	// 	fmt.Printf("Error unmarshalling json: %s", err)
	// 	os.Exit(1)
	// }
	//
	// fmt.Println(packageData)
}
