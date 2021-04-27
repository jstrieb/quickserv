package main

import (
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// NewExecutableHandler returns a handler for an executable path that when
// accessed: executes the file at the path, passes the request body via
// standard input, gets the response via standard output and returns that as
// the response body. The returned function is a closure over the path.
func NewExecutableHandler(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Handle GET requests
		// TODO: Handle form data

		wd, err := os.Getwd()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		cmd := exec.Command(filepath.Join(wd, path))

		// Pass request body on standard input
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		go func() {
			defer stdin.Close()
			// TODO: Handle copy failure
			io.Copy(stdin, r.Body)
		}()

		// Print out stderror messages for debugging
		stderr, err := cmd.StderrPipe()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		go func() {
			defer stderr.Close()
			data, _ := io.ReadAll(stderr)
			if len(data) > 0 {
				log.Println(string(data))
			}
		}()

		// Execute the command and write the output as the HTTP response
		out, err := cmd.Output()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		w.Write(out)
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
		log.Println("Failed while trying to find executables in the working directory!")
		log.Fatal(err)
	}

	// Statically serve non-executable files that don't already have a handler
	mux.Handle("/", http.FileServer(http.Dir(".")))

	// TODO: Display local IP address instead of localhost
	fmt.Println("Staring a server...")
	fmt.Println("Visit http://localhost:42069 to access the server from the local network.")
	fmt.Println("Press Control + C to stop the server.")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":42069", mux))
}
