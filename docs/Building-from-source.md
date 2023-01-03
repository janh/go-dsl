# Building from source

[Go](https://go.dev/) needs to be installed.
A basic version of the application without graphical user interface can be built with a single command:

	go build -o dsl ./cmd # Linux, macOS, …
	go build -o dsl.exe ./cmd # Windows

## Graphical user interface

To build the application with the graphical user interface, you need to additionally specify the `gui` build tag with the option `-tags gui`.
On Windows, you should also add the option `-ldflags="-H windowsgui"` to get rid of the command prompt window which otherwise appears when running the application.
In the end the command should like like this:

	go build -tags gui -o dsl-gui ./cmd # Linux, macOS, …
	go build -tags gui -ldflags="-H windowsgui" -o dsl-gui.exe ./cmd # Windows

As the graphical user interface uses cgo, additional build tools are required in addition to the Go toolchain:

- Linux: basic C build tools, headers for `gtk+-3.0` and `webkit2gtk-4.0`
- Windows: MinGW-w64 toolchain and WebView2 SDK, see documentation of the Go package [github.com/webview/webview](https://github.com/webview/webview) for details
- macOS: Xcode command line developer tools (the system should automatically offer to install them)

## Documentation

The script `docs/build.sh` allows to build HTML documentation from the Markdown files.
Run it from the main folder with the target directory as argument.
Note that the script requires [pandoc](https://pandoc.org/) and a POSIX-compatible shell.
