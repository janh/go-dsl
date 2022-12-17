# xDSL Stats Parser

This is a library and application for reading basic **xDSL stats**, as well bitloading, SNR, QLN and Hlog data.
It **supports various kinds of modems** (see below for details).
The application includes a **graphical user interface**, but it can also be used via the **command line**, and a **web interface** is also available.

If you want to use the library in your own project, please note that the API is not final yet and there are still going to be incompatible changes.

## Getting started

Binaries for Linux, Windows and macOS are automatically built by GitHub Actions.
You can [**download** them here](https://github.com/janh/go-dsl/releases).

In some cases additional steps may be required to get the application to run on your system:

- Linux: The GTK3 and WebKit2GTK libraries need to be installed for the graphical user interface.
- Windows: Make sure that [Microsoft Edge WebView2](https://go.microsoft.com/fwlink/p/?LinkId=2124703) is installed to run the graphical user interface. If SmartScreen blocks the application, you may need to click on "More info" and choose "Run anyway".
- macOS: You may need to right-click the application and select "Open" to start it for the first time.

If you want to build binaries from source yourself, see the next section.
Otherwise skip to the usage section for more information on how to use the application.

## Building from source

Go needs to be installed.
A basic version of the application without graphical user interface can be built with a single command:

	go build -o dsl ./cmd # Linux, macOS, …
	go build -o dsl.exe ./cmd # Windows

To build the application with the graphical user interface, you need to additionally specify the `gui` build tag with the option `-tags gui`.
On Windows, you should also add the option `-ldflags="-H windowsgui"` to get rid of the command prompt window which otherwise appears when running the application.
In the end the command should like like this:

	go build -tags gui -o dsl-gui ./cmd # Linux, macOS, …
	go build -tags gui -ldflags="-H windowsgui" -o dsl-gui.exe ./cmd # Windows

As the graphical user interface uses cgo, additional build tools are required in addition to the Go toolchain:

- Linux: basic C build tools, headers for `gtk+-3.0` and `webkit2gtk-4.0`
- Windows: a C compiler such as TDM-GCC, for building the webview dependency additionally Visual Studio C/C++
- macOS: Xcode command line developer tools (the system should automatically offer to install them)

## Usage

The easiest way to use the application is the graphical user interface.
Just launch it and configure as required for your device.

For all devices you will need to specify the host (either the hostname or an IP address), and most also need a user name.
If a password is required you will be asked for it after the connection is started.

For any device that uses SSH, the application needs to know the host key of the device to validate it.
By default, it tries to use your OpenSSH private key and `known_hosts` file.
This means that the connection should just work, as long as you connected to the device before using SSH.
You can also disable host key validation by specifying the special string `IGNORE` for the known hosts option, but this is not recommended as it is insecure.

As an alternative to the graphical user interface, you can run the application from the command line.
This is also needed if you want to use the web interface.
For information about available command line options, run `./dsl -help`.

Below is a list of supported devices, along with a few device-specfic notes and examples for command line usage.

### Broadcom

Devices with a Broadcom modem that allow to access the system command line via SSH or Telnet.

If the command on the device is not named `xdslctl`, you need to specify the correct name using the `Commmand` option.

	./dsl -d broadcom_ssh -u root 192.168.1.1
	./dsl -d broadcom_telnet -u root 192.168.1.1

### DrayTek

Various devices made by DrayTek.

Only Telnet is supported for connection to the device, as the built-in SSH server suffers from various issues depending on the device (such as outdated cipher suites or crashes).

	./dsl -d draytek_telnet 192.168.1.1

### FRITZ!Box

Should work with any DSL FRITZ!Box devices with somewhat recent firmware (for now the web interface needs to be in German).

It is possible to access additional information that is not available from the web interface, such as QLN and Hlog data.
To do so, set the option `LoadSupportData` to `1`.
However, note that this will significantly increase loading times.

	./dsl -d fritzbox -o LoadSupportData=0 fritz.box

### Lantiq

Devices with a Infineon/Lantiq/Intel/MaxLinear DSL modem and command line access via SSH or Telnet, including OpenWrt devices.

The name of the actual command on the device varies, so you'll need to find out the right one for your device, and specify it with the `Command` option.
By default, `dsl_cpe_pipe` is used.

	./dsl -d lantiq_ssh -o Command="dsl_cpe_pipe.sh" -u root openwrt.lan # OpenWrt
	./dsl -d lantiq_telnet -o Command="/ifx/vdsl2/dsl_pipe" 192.168.16.249 # VINAX modem

### MediaTek

Devices with TrendChip/MediaTek/EcoNet/Airoha DSL hardware and access to the Linux command line via SSH or Telnet.

	./dsl -d mediatek_ssh -u admin 192.168.1.1
	./dsl -d mediatek_telnet -u admin 192.168.1.1

### Sagemcom

Devices made by Sagemcom.

Unfortunately, there are data quality issues at least on some devices or firmware versions.
For example, it is possible that only the "Actual Data Rate" is reported instead of the "Actual Net Data Rate".
This means that the reported data rates are not comparable with other devices when retransmission is active.

	./dsl -d sagemcom speedport.ip

### Speedport

Some Speedport routers, such as Speedport Smart 2.
For the Speedport Pro series, you need to use the Sagemcom client instead.

Only a limited set of data is available, most notably SNR, QLN and Hlog data are missing.

	./dsl -d speedport speedport.ip
