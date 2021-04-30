package main

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// NewExecutableHandler returns a handler for an executable path that when
// accessed: executes the file at the path, passes the request body via
// standard input, gets the response via standard output and returns that as
// the response body. The returned function is a closure over the path.
func NewExecutableHandler(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Executing:", path)

		wd, err := os.Getwd()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		// Create the command using all environment variables. Include a
		// REQUEST_METHOD environment variable in imitation of CGI
		cmd := exec.Command(filepath.Join(wd, path))
		cmd.Env = append(os.Environ(), "REQUEST_METHOD="+r.Method)

		// Pass headers as environment variables in imitation of CGI
		for k, v := range r.Header {
			// The same header can have multiple values
			for _, s := range v {
				cmd.Env = append(cmd.Env, "HTTP_"+strings.ReplaceAll(k, "-", "_")+"="+s)
			}
		}

		// Pass request body on standard input
		stdin, err := cmd.StdinPipe()
		if err != nil {
			log.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		go func() {
			defer stdin.Close()

			// TODO: Handle copy failure in both cases
			switch r.Method {
			case "POST":
				// POST data may not necessarily be form data (e.g.  JSON API
				// request), so don't encode it as a form necessarily.  If it is
				// a form submission, it will be properly encoded anyway.
				io.Copy(stdin, r.Body)

			default:
				// Encode non-POST data as a form for consistency
				err := r.ParseForm()
				if err != nil {
					log.Println(err)
					http.Error(w, http.StatusText(500), 500)
					return
				}

				formdata := []byte(r.Form.Encode())
				io.Copy(stdin, bytes.NewReader(formdata))
			}
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
			// Register executable handlers based on file extension
			switch filepath.Ext(path) {
			case ".exe", ".bat":
				fmt.Println(path)
				mux.HandleFunc("/"+filepath.ToSlash(path), NewExecutableHandler(path))
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
				mux.HandleFunc("/"+filepath.ToSlash(path), NewExecutableHandler(path))
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
	log.Println("Staring a server...")
	fmt.Println("Visit http://localhost:42069 to access the server from the local network.")
	fmt.Println("Press Control + C to stop the server.")
	fmt.Println()

	log.Fatal(http.ListenAndServe(":42069", mux))
}
