# xDSL Stats Parser

This is a library and application for reading basic **xDSL stats**, as well bitloading, SNR, QLN and Hlog data.
It **supports various kinds of modems** (see below for details).
The application includes a **graphical user interface**, but it can also be used via the **command line**, and a **web interface** is also available.

---

**You just want to use the application?**  
[Go here to **download**](https://github.com/janh/go-dsl/releases) binaries for Linux, Windows and macOS.

**You are a developer and want to build your own project based on the library?**  
[Use the package 3e8.eu/go/dsl](https://3e8.eu/go/dsl).
Note that there is no stable release yet, so incompatible changes should be expected occasionally.

---

## Documentation

- **[Getting started](docs/Getting-started.md)**  
  Basic information about how to use the application.
  Includes tips for troubleshooting if it doesn't work right away.

- **[Supported devices](docs/Supported-devices.md)**  
  List of supported device types, including some device-specific notes for configuration.

- **[Configuration files](docs/Configuration-files.md)**  
  Description of all available options.
  Especially useful for the web server which can only be fully configured via config file.

- **[Building from source](docs/Building-from-source.md)**  
  Instructions for compiling the code yourself.
  Usually not necessary if binaries for your platform are available.
