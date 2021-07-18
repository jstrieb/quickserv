package main

import (
	"bytes"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

/******************************************************************************
 * Global Variables and Constants
 *****************************************************************************/

var logger *log.Logger
var routes map[string]string

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
		new_k := strings.ReplaceAll(k, "%", "%25")
		new_k = strings.ReplaceAll(new_k, "&", "%26")
		new_k = strings.ReplaceAll(new_k, "=", "%3D")
		new_form[new_k] = make([]string, len(form[k]))

		// Replace equals, percent, and ampersands in form variable values
		// NOTE: "%" must be encoded first -- see above
		for i, v := range vs {
			v = strings.ReplaceAll(v, "%", "%25")
			v = strings.ReplaceAll(v, "&", "%26")
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

		abspath, err := filepath.Abs(path)
		if err != nil {
			logger.Println(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}
		dir, _ := filepath.Split(abspath)

		// Create the command using all environment variables. Include a
		// REQUEST_METHOD environment variable in imitation of CGI
		cmd := exec.Command(abspath)
		cmd.Env = append(os.Environ(), "REQUEST_METHOD="+r.Method)

		// Execute the route in its own directory so relative paths behave
		// sensibly
		cmd.Dir = dir

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

			if r.Method != "POST" || (len(r.Header["Content-Type"]) > 0 &&
				r.Header["Content-Type"][0] == "application/x-www-form-urlencoded") {
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

// RegisterExecutableHandler associates the given route with a handler to
// execute the file at the given path when accessed. It also modifies the
// global list of routes.
func RegisterExecutableHandler(mux *http.ServeMux, path, route, dir, filename string) {
	handler := NewExecutableHandler(path)

	mux.HandleFunc(route, handler)
	routes[route] = ""

	dirRoute := "/" + filepath.ToSlash(dir)
	if _, in := routes[dirRoute]; strings.TrimSuffix(filename, filepath.Ext(filename)) == "index" && !in {
		// Register handlers with and without the trailing "/" to avoid redirect
		mux.HandleFunc(dirRoute, handler)
		mux.HandleFunc(dirRoute+"/", handler)
		routes[dirRoute+"/"] = route
	}
}

// RegisterPaths walks the current directory and registers handlers to run any
// executable files it finds. The handlers are added as routes to the given
// mux.
func RegisterPaths(mux *http.ServeMux, logfileName string) error {
	return filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Don't re-register any existing routes
		path = filepath.Clean(path)
		route := "/" + filepath.ToSlash(path)
		if _, in := routes[route]; in {
			return nil
		}

		// Ignore the executable for quickserv itself if it's in the directory
		dir, filename := filepath.Split(path)
		dir = filepath.Clean(dir)
		switch filename {
		case "quickserv", "quickserv.exe", logfileName:
			return nil
		}

		// Executables are represented differently on different operating systems
		switch runtime.GOOS {
		case "windows":
			// Register executable handlers based on file extension
			switch filepath.Ext(path) {
			case ".exe", ".bat":
				RegisterExecutableHandler(mux, path, route, dir, filename)
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
				RegisterExecutableHandler(mux, path, route, dir, filename)
			}
		}

		return nil
	})
}

/******************************************************************************
 * Main Function
 *****************************************************************************/

func main() {
	// Parse command line arguments
	// TODO: Handle long and short options
	var logfileName, wd string
	var randomPort bool
	// flag.StringVar(&logfileName, "l", "-", "Log file path. Stdout if unspecified.")
	flag.StringVar(&logfileName, "logfile", "-", "Log file path. Stdout if unspecified.")
	// flag.StringVar(&wd, "d", ".", "Folder to serve files from.")
	flag.StringVar(&wd, "dir", ".", "Folder to serve files from.")
	flag.BoolVar(&randomPort, "random-port", false, "Use a random port instead of 42069.")
	flag.Parse()

	// Initialize logfile relative to the initial working directory
	var logfile *os.File
	if logfileName == "-" {
		logfile = os.Stdout
	} else {
		mode := os.O_WRONLY | os.O_APPEND | os.O_CREATE
		var err error
		logfile, err = os.OpenFile(logfileName, mode, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		if abspath, err := filepath.Abs(logfileName); err == nil {
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
	routes = make(map[string]string)
	err = RegisterPaths(mux, logfileName)
	if err != nil {
		logger.Println("Failed while trying to find executables in the working directory!")
		logger.Fatal(err)
	}

	// Print non-static routes that will be executed (if any)
	if len(routes) > 0 {
		fmt.Println("Files that will be executed if accessed: ")
		for k, v := range routes {
			if v == "" {
				fmt.Println(k)
			} else {
				fmt.Printf("%v -> %v\n", k, v)
			}
		}
	} else {
		fmt.Println("No files will be executed if accessed!")
		fmt.Println("To make a file executable on Mac or Linux, run \"chmod +x filename\" from the Terminal.")
		fmt.Println("On Windows only .bat and .exe files will be executed.")
		// TODO
		// fmt.Println("For more information see the documentation here: TODO")
	}
	fmt.Println("")

	// Statically serve non-executable files that don't already have a handler
	mux.Handle("/", http.FileServer(http.Dir(".")))

	// Pick a random port if the user wants -- for slightly more professional
	// demos where the number 42069 might be undesirable
	var port int64
	if randomPort {
		// Avoid privileged ports (those below 1024). Cryptographic randomness
		// might be a bit much here, but ¯\(°_o)/¯
		rawPort, err := rand.Int(rand.Reader, big.NewInt(65535-1025))
		if err != nil {
			logger.Fatal(err)
		}
		port = rawPort.Int64() + 1025
		fmt.Printf("Using port %v.\n\n", port)
	} else {
		port = 42069
	}

	localIP := GetLocalIP()
	logger.Println("Starting a server...")
	fmt.Printf("Visit http://%v:%v to access the server from the local network.\n", localIP, port)
	fmt.Println("Press Control + C to stop the server.")
	fmt.Println()

	logger.Fatal(http.ListenAndServe(":"+strconv.FormatInt(port, 10), mux))
}
