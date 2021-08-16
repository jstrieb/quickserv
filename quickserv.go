package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/jstrieb/killfam"
)

/******************************************************************************
 * Global Variables and Constants
 *****************************************************************************/

var logger *log.Logger

/******************************************************************************
 * Helper Functions
 *****************************************************************************/

// NewLogFile initializes the logfile relative to the initial working directory.
func NewLogFile(logfileName string) *log.Logger {
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
		defer logfile.Close()
		if abspath, err := filepath.Abs(logfileName); err == nil {
			fmt.Printf("Logging to folder:\n%v\n", abspath)
		} else {
			log.Fatal(err)
		}
	}
	return log.New(logfile, "", log.LstdFlags)
}

// PickPort returns the port that the server should run on. It either returns
// port 42069 or a random port depending on the value of the argument.
func PickPort(randomPort bool) int64 {
	if randomPort {
		// Avoid privileged ports (those below 1024). Cryptographic randomness
		// might be a bit much here, but ¯\(°_o)/¯
		rawPort, err := rand.Int(rand.Reader, big.NewInt(65535-1025))
		if err != nil {
			logger.Fatal(err)
		}
		port := rawPort.Int64() + 1025
		fmt.Printf("Using port %v.\n\n", port)
		return port
	} else {
		return 42069
	}
}

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
	newForm := make(url.Values, len(form))
	for k, vs := range form {
		// Replace equals, percent, and ampersands in form variable names
		// NOTE: "%" must be encoded first -- see above
		newK := strings.ReplaceAll(k, "%", "%25")
		newK = strings.ReplaceAll(newK, "&", "%26")
		newK = strings.ReplaceAll(newK, "=", "%3D")
		newForm[newK] = make([]string, len(form[k]))

		// Replace equals, percent, and ampersands in form variable values
		// NOTE: "%" must be encoded first -- see above
		for i, v := range vs {
			v = strings.ReplaceAll(v, "%", "%25")
			v = strings.ReplaceAll(v, "&", "%26")
			v = strings.ReplaceAll(v, "=", "%3D")
			newForm[newK][i] = v
		}
	}

	// Encode the form as a string and decode as almost entirely plain text
	rawFormData := []byte(newForm.Encode())
	formData, err := url.QueryUnescape(string(rawFormData))
	if err != nil {
		return nil, err
	}

	return []byte(formData), nil
}

// IsWSL returns whether or not the binary is running inside Windows Subsystem
// for Linux (WSL). It guesses at this based on some heuristics. For more, see:
// https://github.com/microsoft/WSL/issues/423#issuecomment-221627364
func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	filesToCheck := []string{"/proc/version", "/proc/sys/kernel/osrelease"}

	r, err := regexp.Compile("(?i)(wsl|microsoft|windows)")
	if err != nil {
		logger.Println("Error compiling regular expression.")
		logger.Fatal(err)
		return false
	}

	for _, filename := range filesToCheck {
		raw, err := ioutil.ReadFile(filename)
		if err == nil && r.Match(raw) {
			return true
		}
	}
	return false
}

// IsPathExecutable returns whether or not a given file is executable based on
// its path and/or its permission bits (depending on the operating system).
//
// On Windows, a file is executable if and only if it has a file extension of
// "exe" or "bat." On other operating systems, any file with the execute bit set
// for at least one user is deemed executable.
//
// NOTE: This function may not return accurate results if given a directory as
// input. This has not been extensively tested.
func IsPathExecutable(path string, fileinfo fs.FileInfo) bool {
	goos := runtime.GOOS

	// Check if we are in Windows Subsystem for Linux. If so, behave differently
	// since it's Linux but we can run .exe files, and since the permission bits
	// are all messed up such that everything will be viewed as executable.
	if IsWSL() {
		goos = "wsl"
	}

	switch strings.ToLower(goos) {
	case "windows":
		// Register executable handlers based on file extension
		switch strings.ToLower(filepath.Ext(path)) {
		case ".exe", ".bat", ".cmd":
			return true
		}

	case "wsl":
		// Register executable handlers based on file extension
		// TODO: Add more filenames? Or perhaps improve executable detection?
		switch strings.ToLower(filepath.Ext(path)) {
		case ".exe", ".sh", ".py":
			return true
		}

	default:
		// TODO: Does it make sense to look for files executable by any user?
		filemode := fileinfo.Mode()
		if !filemode.IsDir() && filemode.Perm()&0111 != 0 {
			return true
		}
	}

	return false
}

// ExecutePath executes the file at the path, passes the request body via
// standard input, gets the response via standard output and writes that as the
// response body.
//
// NOTE: Expects the input path to be rooted with forward slashes as the
// separator (HTTP request style)
func ExecutePath(ctx context.Context, execPath string, w http.ResponseWriter, r *http.Request) {
	logger.Println("Executing:", execPath)

	// Clean up the path and make it un-rooted
	if strings.HasPrefix(execPath, "/") {
		execPath = "." + execPath
	}
	execPath = path.Clean(execPath)
	abspath, err := filepath.Abs(filepath.FromSlash(execPath))
	if err != nil {
		logger.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
	dir, _ := filepath.Split(abspath)

	// Create the command using all environment variables. Include a
	// REQUEST_METHOD environment variable in imitation of CGI
	//
	// I tried to do exec.CommandContext here, but it doesn't kill child
	// processes, so anything run from a script keeps on going when the
	// connection terminates. Instead I use a goroutine listening to the context
	// with a custom package below.
	cmd := exec.Command(abspath)
	cmd.Env = append(os.Environ(), "REQUEST_METHOD="+r.Method)

	// Make the process children killable
	killfam.Augment(cmd)

	// Pass headers as environment variables in imitation of CGI
	for k, v := range r.Header {
		// The same header can have multiple values
		for _, s := range v {
			cmd.Env = append(cmd.Env, "HTTP_"+strings.ReplaceAll(k, "-", "_")+"="+s)
		}
	}

	// Execute the route in its own directory so relative paths in the executed
	// program behave sensibly
	cmd.Dir = dir

	// Pass request body on standard input
	stdin, err := cmd.StdinPipe()
	if err != nil {
		logger.Println(err)
		logger.Println("Couldn't pass the request body via stdin.")
		http.Error(w, http.StatusText(500), 500)
		return
	}
	go func() {
		defer stdin.Close()

		// TODO: Test with PUT/DELETE/other HTTP methods
		if r.Method != "POST" || (len(r.Header["Content-Type"]) > 0 &&
			r.Header["Content-Type"][0] == "application/x-www-form-urlencoded") {
			// If the submission is a non-POST request, or is a form
			// submission according to content type, treat it like a form
			err := r.ParseForm()
			if err != nil {
				logger.Println(err)
				logger.Println("Couldn't parse the request form.")
				http.Error(w, http.StatusText(500), 500)
				return
			}

			formData, err := DecodeForm(r.Form)
			if err != nil {
				logger.Println(err)
				logger.Println("Couldn't percent-decode the request form.")
				http.Error(w, http.StatusText(500), 500)
				return
			}
			_, err = io.Copy(stdin, bytes.NewReader(formData))
			if err != nil {
				logger.Println(err)
				logger.Println("Couldn't copy the form data to the program.")
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
				logger.Println("Couldn't copy the request body to the program.")
				http.Error(w, http.StatusText(500), 500)
				return
			}
		}
	}()

	// Print out stderror messages for debugging
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logger.Println(err)
		logger.Println("Couldn't get stderr output to print.")
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

	// Kill the process if the user terminates their connection
	cmdDone := make(chan error)
	go func() {
		select {
		case <-ctx.Done():
			logger.Println("User disconnected. Killing process.")
			if err := killfam.KillTree(cmd); err != nil {
				logger.Println(err)
			}

		case <-cmdDone:
			return
		}
	}()

	// Execute the command and write the output as the HTTP response
	out, err := cmd.Output()
	cmdDone <- err
	if err != nil {
		logger.Println(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	w.Write(out)
}

// FindIndexFile returns the path to the index file of the directory path given
// as input (if one exists). If there is no index file, or if there was a fatal
// error during the search, the the first returned value is the empty string and
// the second value is false.
//
// NOTE: The input dir is expected to be a rooted path with forward slashes, and
// the output has the same format
func FindIndexFile(dir string) (string, bool) {
	file, err := http.Dir(".").Open(dir)
	if err != nil {
		return "", false
	}
	defer file.Close()

	files, err := file.Readdir(-1)
	if err != nil {
		return "", false
	}
	for _, file := range files {
		filename := file.Name()
		if IsPathExecutable(filename, file) &&
			strings.ToLower(strings.TrimSuffix(filename, path.Ext(filename))) == "index" {
			return path.Join(dir, filename), true
		}
	}
	return "", false
}

// FindExecutablePaths walks the current directory and locates paths that will
// be executed when visited. It returns them as a map. In the map keys are paths
// that cause a file to be executed, and values are either the empty string or
// the file to be executed when the path is accessed.
func FindExecutablePaths(logfileName string) (map[string]string, error) {
	routes := make(map[string]string)

	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Clean up the path and format it HTTP-style
		path = filepath.Clean(path)
		path = "/" + filepath.ToSlash(path)

		// Ignore the executable for quickserv itself if it's in the directory.
		// Also ignore the logfile if it's present.
		_, filename := filepath.Split(path)
		fileinfo, err := d.Info()
		if err != nil {
			logger.Printf("Couldn't get file info for %v.\n", filename)
			return nil
		}
		switch strings.ToLower(filename) {
		case "quickserv", "quickserv.exe", logfileName:
			return nil
		}

		// Find the index file if path is a directory
		if fileinfo.IsDir() {
			index, found := FindIndexFile(path)
			if found {
				routes[path] = index
			}
			return nil
		}

		// Print a result if executable
		if IsPathExecutable(path, fileinfo) {
			routes[path] = ""
		}
		return nil
	})

	return routes, err
}

// NewMainHandler returns an http.Handler that looks at the file a user requests
// and decides whether to execute it, or pass it to an http.FileServer.
func NewMainHandler(filesystem http.FileSystem) http.Handler {
	fileserver := http.FileServer(filesystem)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write maximally permissive CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Expose-Headers", "*")

		// Clean up the request path
		reqPath := r.URL.Path
		if !strings.HasPrefix(reqPath, "/") {
			reqPath = "/" + reqPath
			r.URL.Path = reqPath
		}
		reqPath = path.Clean(reqPath)

		// Open the path in the filesystem for further inspection
		f, err := filesystem.Open(reqPath)
		if err != nil {
			// If we can't open the file, let the FileServer handle it correctly
			logger.Println(err)
			fileserver.ServeHTTP(w, r)
			return
		}
		defer f.Close()
		d, err := f.Stat()
		if err != nil {
			logger.Println(err)
			// If we can't open the file, let the FileServer handle it correctly
			fileserver.ServeHTTP(w, r)
			return
		}

		// If the path is a directory, look for an index file. If none found,
		// serve up the directory. Otherwise, act like the executable was the
		// original requested path.
		if d.IsDir() {
			index, found := FindIndexFile(reqPath)
			if !found {
				fileserver.ServeHTTP(w, r)
				return
			} else {
				reqPath = index
				fNew, err := filesystem.Open(reqPath)
				if err != nil {
					// If we can't open the file, let the FileServer handle it correctly
					logger.Println(err)
					fileserver.ServeHTTP(w, r)
					return
				}
				defer fNew.Close()
				d, err = fNew.Stat()
				if err != nil {
					logger.Println(err)
					// If we can't open the file, let the FileServer handle it correctly
					fileserver.ServeHTTP(w, r)
					return
				}
			}
		}

		if IsPathExecutable(reqPath, d) {
			// If the path is executable, run it
			ExecutePath(r.Context(), reqPath, w, r)
		} else {
			fileserver.ServeHTTP(w, r)
		}
	})
}

/******************************************************************************
 * Main Function
 *****************************************************************************/

func main() {
	// Parse command line arguments
	var logfileName, wd string
	var randomPort bool
	flag.StringVar(&logfileName, "logfile", "-", "Log file path. Stdout if unspecified.")
	flag.StringVar(&wd, "dir", ".", "Folder to serve files from.")
	flag.BoolVar(&randomPort, "random-port", false, "Use a random port instead of 42069.")
	flag.Parse()

	logger = NewLogFile(logfileName)

	// Switch directories and print the current working directory
	if err := os.Chdir(wd); err != nil {
		logger.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatal(err)
	}
	fmt.Printf("Running in folder:\n%v\n\n", wd)

	// Print non-static routes that will be executed (if any)
	routes, err := FindExecutablePaths(logfileName)
	if err != nil {
		logger.Fatal(err)
	}
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
		fmt.Println("When accessed, non-executable files will be viewed by the user instead of being run.")
		// TODO
		// fmt.Println("For more information see the documentation here: TODO")
	}
	fmt.Println("")

	// Pick a random port if the user wants -- for slightly more professional
	// demos where the number 42069 might be undesirable
	port := PickPort(randomPort)

	localIP := GetLocalIP()
	logger.Println("Starting a server...")
	fmt.Printf("Visit http://%v:%v to access the server from the local network.\n", localIP, port)
	fmt.Println("Press Control + C to stop the server.")
	fmt.Println()

	// Build a handler that decides whether to serve static files or dynamically
	// execute them
	handler := NewMainHandler(http.Dir("."))
	addr := ":" + strconv.FormatInt(port, 10)
	if err = http.ListenAndServe(addr, handler); err != nil {
		logger.Println("Make sure you are only running one instance of QuickServ!")
		logger.Fatal(err)
	}
}
