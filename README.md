# QuickServ

**Quick**, no-setup web **Serv**er


## About

QuickServ makes creating web applications *dangerously* easy, no matter what
programming language you use. QuickServ:

- Has sensible defaults 
- Requires no configuration
- Prints helpful error messages directly to the console
- Runs on any modern computer, with no setup or installation
- Knows which files to run server-side, and which to serve statically
- Passes data through standard input and standard output
- Doesn't require understanding the intricacies of HTTP

[QuickServ should not be used in production.](#disclaimer) 

QuickServ brings the heady fun of the 90s Internet to the 2020s. It is inspired
by the [Common Gateway Interface
(CGI)](https://en.wikipedia.org/wiki/Common_Gateway_Interface), but is much
easier to set up and use. Unlike CGI, QuickServ works out of the box with no
searching for obscure log files, no learning how HTTP headers work, no fiddling
with permission bits, no wondering where to put your scripts, and no struggling
with Apache `mod_cgi` configurations.


## Get Started

Using QuickServ is as easy as downloading the program, dragging it to your
project folder, and double clicking it. It automatically detects which files to
run, and which to serve directly to the user. 

### Windows

<details>
<summary>Click to view details</summary>

[Download for
Windows](https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_windows_x64.exe).

</details>

### Mac

<details>
<summary>Click to view details</summary>

[Download for Intel
Mac](https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_macos_x64).
[Download for Arm
Mac](https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_macos_arm).

</details>

### Linux

<details>
<summary>Click to view details</summary>

Download for Linux. Run it in your project folder.

``` bash
# Download
sudo curl \
    --location \
    --output /usr/local/bin/quickserv 
    https://github.com/jstrieb/quickserv/releases/latest/download/quickserv_linux_x64

# Go to project folder and run
cd /my/project/folder
quickserv
```

</details>


## Examples

TODO


## How It Works

TODO


## Disclaimer

QuickServ is not designed for production use. It was not created to be fast or
secure. Using QuickServ in production puts your users and yourself at risk,
please do not do it.

QuickServ lets people build dangerously insecure things. It does not sanitize
any inputs or outputs. It uses one process per request, and is susceptible to a
denial of service attack. Its security model presumes web users are trustworthy.
These characteristics make prototyping easier, but are not safe on the public
Internet.

To deter using QuickServ in production, it runs on port 42069. Hopefully that
makes everyone think twice before entering it into a reverse proxy or port
forward config. For a more professional demo, the command-line flag
`--random-port` will instead use a random port, determined at runtime.

QuickServ is similar to the ancient CGI protocol. There are many
well-articulated, well-established [reasons that CGI is bad in
production](https://www.embedthis.com/blog/posts/stop-using-cgi/stop-using-cgi.html),
and they all apply to QuickServ in production.


## Advanecd

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


## Support the Project

There are a few things you can do to support the project:

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

TODO