package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

// NewExecutableHandler returns a handler for an executable path that when
// accessed: executes the file at the path, passes the request body via
// standard input, gets the response via standard output and returns that as
// the response body
func NewExecutableHandler(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO
		fmt.Fprintln(w, path)
	}
}

func main() {
	mux := http.NewServeMux()

	// Walk the working directory looking for executable files and register
	// handlers to execute them
	fmt.Println("Files that will be executed if accessed: ")
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Executables are represented differently on different operating systems
		switch runtime.GOOS {
		case "windows":
			// If ends in .exe, register a file handler
			if filepath.Ext(path) == ".exe" {
				fmt.Println(path)
				mux.HandleFunc(string(filepath.Separator)+path, NewExecutableHandler(path))
			}

		default:
			// Check the permission bits, assume consistency across Linux/OS X
			fileinfo, err := d.Info()
			if err != nil {
				return err
			}
			// TODO: Does it make sense to look for executable by any user?
			filemode := fileinfo.Mode()
			if !filemode.IsDir() && filemode.Perm()&0111 != 0 {
				fmt.Println(path)
				mux.HandleFunc(string(filepath.Separator)+path, NewExecutableHandler(path))
			}
		}

		return nil
	})

	fmt.Println("")
	if err != nil {
		log.Fatal("Failed while trying to find executables in the working directory!")
	}

	// Statically serve non-executable files that don't already have a handler
	mux.Handle("/", http.FileServer(http.Dir(".")))

	// TODO: Display local IP address instead of localhost
	fmt.Println("Staring a server...")
	fmt.Println("Visit http://localhost:42069 to access the server from the local network.")
	log.Fatal(http.ListenAndServe(":42069", mux))
}
