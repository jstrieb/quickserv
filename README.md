# QuickServ

**Quick**, no-setup web **Serv**er


## About

QuickServ makes creating web applications *dangerously* easy, no matter what
programming language you use. QuickServ:

- Has sensible defaults 
- Requires no configuration
- Prints helpful error messages
- Runs on any modern computer, with no setup or installation
- Knows which files to run server-side, and which to serve statically
- Passes data through standard input and standard output
- Doesn't require understanding the intricacies of HTTP

[QuickServ should not be used in production.](#disclaimer) 


## Get Started

Using QuickServ is as easy as downloading the program, dragging it to your
project folder, and double clicking it. It automatically detects which files to
run, and which to serve directly to the user. 

### Windows

<details>
<summary>Click to view details</summary>

[Download for Windows]()

</details>

### Mac

<details>
<summary>Click to view details</summary>

[Download for Mac]()

</details>

### Linux

<details>
<summary>Click to view details</summary>

[Download for Linux]()

</details>


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


## Disclaimer

QuickServ is not designed for production use. It was not created to be fast or
secure. Using QuickServ in production puts your users and yourself at risk,
please do not do it.

QuickServ lets people build dangerously insecure things. It does not sanitize
any inputs or outputs. Its security model presumes web users are trustworthy.
This is safe for prototypes, but not on the Internet in general.

To deter using QuickServ in production, it runs on port 42069. Hopefully that
makes everyone think twice when entering it into a reverse proxy or port forward
config. To run a more professional demo, the command-line flag `--random-port`
will instead run it on a random port, determined at runtime.