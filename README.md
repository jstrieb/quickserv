# QuickServ

**Quick**, no-setup web **Serv**er


## About

QuickServ makes creating web applications easy, no matter what programming
language you use. 

By relying on using sensible defaults, printing helpful error messages, and
passing data through standard input and output channels QuickServ lowers the
barrier to creating interactive programs on the web.


## Get Started

Using QuickServ is as easy as downloading the program and running it in your
project folder. It will automatically detect which files to run, and which to
serve directly to the user. 

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


## Disclaimer

QuickServ is not designed for production use. It was not created to be fast or
secure. Using QuickServ in production puts your users and yourself at risk,
please do not do it.

To deter using it QuickServ production, it runs on port 42069. Hopefully that
will make you think twice before entering it into a reverse proxy or port
forward config. To run a demo (like for a hackathon) where you may need to be
more professional, the command-line flag `--random-port` will run it on a random
port, decided at runtime.