# QuickServ

**Quick**, no-setup web **Serv**er


## About

QuickServ makes creating web applications [*dangerously*](#disclaimer) easy, no
matter what programming language you use. QuickServ:

- Has sensible defaults 
- Prints helpful error messages directly to the console
- Runs on any modern computer, with no setup or installation
- Needs no configuration
- Knows which files to run server-side, and which to serve plain
- Works with any programming language that can `read` and `write`
- Doesn't require understanding the intricacies of HTTP

QuickServ brings the heady fun of the 1990s Internet to the 2020s. It is
inspired by the [Common Gateway Interface
(CGI)](https://en.wikipedia.org/wiki/Common_Gateway_Interface), but is much
easier to set up and use. Unlike CGI, it works out of the box with no searching
for obscure log files, no learning how HTTP headers work, no fiddling with
permission bits, no worrying about CORS, no wondering where to put your scripts,
and no struggling with Apache `mod_cgi` configurations. 

<!-- I promise I'm not jaded about CGI or anything ;) -->

It is perfect for:

- Building hackathon projects without learning a web framework
- Creating internal tools
- Prototyping applications using any language
- Giving scripts web interfaces
- Controlling hardware with Raspberry Pis on your local network
- Trying out web development without being overwhelmed

[QuickServ should not be used on the open Internet.](#disclaimer) 


## Get Started

Using QuickServ is as easy as downloading the program, dragging it to your
project folder, and double clicking to run. It automatically detects which files
to execute, and which to serve directly to the user. 

### Windows

<details>
<summary>Click to view details</summary>

1. [Download for
   Windows](https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_windows_x64.exe).

2. Make a project folder and add files to it. For example, if Python is
   installed, create a file called `test.py` containing:

   ``` python
   #!python

   import random
   print(random.randint(0, 420))
   ```

   Since `test.py` starts with `#!something`, where `something test.py` is the
   command to execute the file, QuickServ will know to run it. If QuickServ is
   not running your file, make sure to add this to the beginning. 
   
   On Windows, QuickServ also knows to automatically run files that end in
   `.exe` and `.bat`. Any other file type needs to start with `#!something` if
   it should be run.

3. Move the downloaded `quickserv_windows_x64.exe` file to the project folder.

   <!-- TODO image -->

4. Double click `quickserv_windows_x64.exe` in the project folder to start
   QuickServ.

   <!-- TODO image -->

5. Go to <http://127.0.0.1:42069> (or the address shown by QuickServ) to connect
   to your web application. In the example, to run `test.py`, go to
   <http://127.0.0.1:42069/test.py>.

</details>

### Mac

<details>
<summary>Click to view details</summary>

[Download for Intel
Mac](https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_macos_x64).
[Download for Arm
Mac](https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_macos_arm).

</details>

### Raspberry Pi

<details>
<summary>Click to view details</summary>

<!-- TODO -->

It's easiest to install and run via the command line. Open the Terminal.

<!-- TODO Image -->

Enter the following commands. A password may be required for the first command. 

``` bash
# Download
sudo curl \
    --location \
    --output /usr/local/bin/quickserv 
    https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_raspi_x64

# Make a project folder
mkdir -p my/project/folder

# Go to project folder
cd my/project/folder

# Add a test file 
cat <<EOF > temp.py
#!python3

import random
print(random.randint(0, 420))
EOF

# Run QuickServ
quickserv
```

Go to <http://127.0.0.1:42069> (or the address shown by QuickServ) to connect to
your web application. For example, to run `test.py`, go to
<http://127.0.0.1:42069/test.py>.

</details>

### Others

<details>
<summary>Click to view details</summary>

Clicking to run executables does not have consistent behavior across Linux
distros, so it's easiest to install and run via the command line. It may be
necessary to change the filename at the end of the `curl` HTTP request URL
below.

See all download options on the [releases
page](https://github.com/jstrieb/quickserv/releases/latest).

``` bash
# Download
sudo curl \
    --location \
    --output /usr/local/bin/quickserv 
    https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_linux_x64

# Make a project folder
mkdir -p /my/project/folder

# Go to project folder
cd /my/project/folder

# Add a test file 
cat <<EOF > temp.py
#!python3

import random
print(random.randint(0, 420))
EOF

# Run QuickServ
quickserv
```

Go to <http://127.0.0.1:42069> (or the address shown by QuickServ) to connect to
your web application. For example, to run `test.py`, go to
<http://127.0.0.1:42069/test.py>.

</details>


## Examples

TODO


## How It Works

<details>
<summary>Click to view details</summary>

TODO

</details>


## Disclaimer

Do not run QuickServ on the public Internet. Only run it on private networks.

QuickServ is not designed for production use. It was not created to be fast or
secure. Using QuickServ in production puts your users and yourself at risk,
please do not do it.

QuickServ lets people build dangerously insecure things. It does not sanitize
any inputs or outputs. It uses one process per request, and is susceptible to a
denial of service attack. Its security model presumes web users are trustworthy.
These characteristics make prototyping easier, but are not safe on the public
Internet.

To deter using QuickServ in production, it runs on port `42069`. Hopefully that
makes everyone think twice before entering it into a reverse proxy or port
forward config. For a more professional demo, the command-line flag
`--random-port` will instead use a random port, determined at runtime.

QuickServ is similar to the ancient CGI protocol. There are many
well-articulated, well-established [reasons that CGI is bad in
production](https://www.embedthis.com/blog/posts/stop-using-cgi/stop-using-cgi.html),
and they all apply to QuickServ in production.


## Advanced

QuickServ has advanced options configured via command line flags. These
change how and where QuickServ runs, as well as where it saves its output.

```
Usage: 
quickserv [options]

Options:
  --dir string
        Folder to serve files from. (default ".")
  --logfile string
        Log file path. Stdout if unspecified. (default "-")
  --no-pause
        Don't pause before exiting after fatal error.
  --random-port
        Use a random port instead of 42069.
```


<!--
## Motivation & Philosophy

The idea came from spending way too much time getting set up during a hackathon
with friends in college.

I started this project in C, but I finished it in Golang. It leans heavily on
the Go standard library. Go's easy web server integration meant that I could
spend most of my time optimizing the user experience. Thankfully, Go shoulders
much of the complexity for the end-user.

At home, I constantly use it to give my shell scripts simple web front-ends.
-->


## Support the Project

There are a few ways to support the project:

- Star the repository and follow me on GitHub
- Share and upvote on sites like Twitter, Reddit, and Hacker News
- Report any bugs, glitches, or errors that you find

These things motivate me to to keep sharing what I build, and they provide
validation that my work is appreciated! They also help me improve the project.
Thanks in advance!

If you are insistent on spending money to show your support, I encourage you to
instead make a generous donation to one of the following organizations. By
advocating for Internet freedoms, organizations like these help me to feel
comfortable releasing work publicly on the Web.

- [Electronic Frontier Foundation](https://supporters.eff.org/donate/)
- [Signal Foundation](https://signal.org/donate/)
- [Mozilla](https://donate.mozilla.org/en-US/)
- [The Internet Archive](https://archive.org/donate/index.php)


## Acknowledgments

This project would not be possible without the help and support of:

- [Logan Snow](https://github.com/lsnow99)
- [Amy Liu](https://www.linkedin.com/in/amyjl/)
- Everyone who [supports the project](#support-the-project)