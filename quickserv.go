package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

/******************************************************************************
 * Global Variables and Constants
 *****************************************************************************/

var logger *log.Logger

/******************************************************************************
 * Helper Functions
 *****************************************************************************/

// GetLocalIP finds the IP address of the computer on the local area network so
// anyone on the same network can connect to the server. Code inspired by:
// https://stackoverflow.com/a/37382208/1376127
func GetLocalIP() string {
	conn, err := net.Dial("udp", "example.com:80")
	if err != nil {
		logger.Println(err)
		logger.Println("Could not get local IP address.")
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}

// DecodeForm performs URL query unescaping on encoded form data to make parsing
// easier. Remaining encoded strings are:
// 		% -> %25
// 		& -> %26
// 		= -> %3D
//
// If "%" is not encoded first in the pre-encoding step, then it will encode the
// percent signs from the encoding of & and = in addition to real percent signs,
// which will give incorrect results.
func DecodeForm(form url.Values) ([]byte, error) {
	// Pre-encoding step where special characters are encoded before the entire
	// form is encoded and then decoded
	new_form := make(url.Values, len(form))
	for k, vs := range form {
		// Replace equals, percent, and ampersands in form variable names
		// NOTE: "%" must be encoded first -- see above
		new_k := strings.ReplaceAll(k, "%", "%26")
		new_k = strings.ReplaceAll(new_k, "&", "%25")
		new_k = strings.ReplaceAll(new_k, "=", "%3D")
		new_form[new_k] = make([]string, len(form[k]))

		// Replace equals, percent, and ampersands in form variable values
		// NOTE: "%" must be encoded first -- see above
		for i, v := range vs {
			v = strings.ReplaceAll(v, "%", "%26")
			v = strings.ReplaceAll(v, "&", "%25")
			v = strings.ReplaceAll(v, "=", "%3D")
			new_form[new_k][i] = v
		}
	}

	// Encode the form as a string and decode as almost entirely the plain text
	raw_form_data := []byte(new_form.Encode())
	form_data, err := url.QueryUnescape(string(raw_form_data))
	if err != nil {
		return nil, err
	}

	return []byte(form_data), nil
}

// NewExecutableHandler returns a handler for an executable path that, when
// accessed: executes the file at the path, passes the request body via
// standard input, gets the response via standard output and returns that as
// the response body. The returned function is a closure over the path.
func NewExecutableHandler(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Println("Executing:", path)

		wd, err := os.Getwd()
		if err != nil {
			logger.Println(err)
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
			logger.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		go func() {
			defer stdin.Close()

			if r.Method != "POST" || (len(r.Header["Content-Type"]) >= 1 && r.Header["Content-Type"][0] == "application/x-www-form-urlencoded") {
				// If the submission is a non-POST request, or is a form
				// submission according to content type, treat it like a form
				err := r.ParseForm()
				if err != nil {
					logger.Println(err)
					http.Error(w, http.StatusText(500), 500)
					return
				}

				form_data, err := DecodeForm(r.Form)
				if err != nil {
					logger.Println(err)
					http.Error(w, http.StatusText(500), 500)
					return
				}
				_, err = io.Copy(stdin, bytes.NewReader(form_data))
				if err != nil {
					logger.Println(err)
					http.Error(w, http.StatusText(500), 500)
					return
				}
			} else {
				// This POST data is not form data (may be a JSON API request,
				// for example), so don't encode it as a form. If it is a
				// multipart or other form submission, it will be properly
				// encoded already.
				_, err := io.Copy(stdin, r.Body)
				if err != nil {
					logger.Println(err)
					http.Error(w, http.StatusText(500), 500)
					return
				}
			}
		}()

		// Print out stderror messages for debugging
		stderr, err := cmd.StderrPipe()
		if err != nil {
			logger.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		go func() {
			defer stderr.Close()
			data, _ := io.ReadAll(stderr)
			if len(data) > 0 {
				logger.Println(string(data))
			}
		}()

		// Execute the command and write the output as the HTTP response
		out, err := cmd.Output()
		if err != nil {
			logger.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		w.Write(out)
	}
}

/******************************************************************************
 * Main Function
 *****************************************************************************/

func main() {
	// Parse command line arguments
	var logfilename, wd string
	// flag.StringVar(&logfilename, "l", "-", "Log file path. Stdout if unspecified.")
	flag.StringVar(&logfilename, "logfile", "-", "Log file path. Stdout if unspecified.")
	flag.StringVar(&wd, "working-directory", ".", "Folder to serve files from.")
	flag.Parse()

	// Initialize logger with logfile relative to the initial working directory
	var logfile *os.File
	if logfilename == "-" {
		logfile = os.Stdout
	} else {
		mode := os.O_WRONLY | os.O_APPEND | os.O_CREATE
		var err error
		logfile, err = os.OpenFile(logfilename, mode, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		if abspath, err := filepath.Abs(logfilename); err == nil {
			fmt.Printf("Logging to folder:\n%v\n", abspath)
		} else {
			log.Fatal(err)
		}
	}
	logger = log.New(logfile, "", log.LstdFlags)

	// Switch directories and print the current working directory
	if err := os.Chdir(wd); err != nil {
		logger.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Printf("Running in folder:\n%v\n\n", wd)

	mux := http.NewServeMux()

	// Walk the working directory looking for executable files and register
	// handlers to execute them
	fmt.Println("Files that will be executed if accessed: ")
	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Ignore the executable for quickserv itself if it's in the directory
		_, filename := filepath.Split(path)
		switch filename {
		case "quickserv", "quickserv.exe", logfilename:
			return nil
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
			// Check the permission bits, assume consistency across Linux/OS X/others
			fileinfo, err := d.Info()
			if err != nil {
				return err
			}
			// TODO: Does it make sense to look for files executable by any user?
			filemode := fileinfo.Mode()
			if !filemode.IsDir() && filemode.Perm()&0111 != 0 {
				fmt.Println(path)
				mux.HandleFunc("/"+filepath.ToSlash(path), NewExecutableHandler(path))
			}
		}

		return nil
	})
	fmt.Println("")
	if err != nil {
		logger.Println("Failed while trying to find executables in the working directory!")
		logger.Fatal(err)
	}

	// Statically serve non-executable files that don't already have a handler
	mux.Handle("/", http.FileServer(http.Dir(".")))

	localIP := GetLocalIP()
	logger.Println("Starting a server...")
	fmt.Printf("Visit http://%s:42069 to access the server from the local network.\n", localIP)
	fmt.Println("Press Control + C to stop the server.")
	fmt.Println()

	logger.Fatal(http.ListenAndServe(":42069", mux))
}
