package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
)

func main() {
	mux := http.NewServeMux()

	// Walk the working directory looking for executable files
	// TODO: Register handlers for them
	fmt.Println("Files being served: ")
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

    fmt.Println(path)
		return nil
	})

	fmt.Println("")
	if err != nil {
		log.Fatal("Failed to find executables in the working directory!")
	}

	mux.Handle("/", http.FileServer(http.Dir(".")))

	// TODO: Display local IP address instead of localhost
	fmt.Println("Staring a server...")
	fmt.Println("Visit http://localhost:42069 to access the server from the local network.")
	log.Fatal(http.ListenAndServe(":42069", mux))
}
