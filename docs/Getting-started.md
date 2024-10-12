# Getting started

The easiest way to use the application is the graphical user interface.
If the application does not run, see the troubleshooting section below.

## Configuration

For all devices you will need to specify the host (either the hostname or an IP address, optionally including port), and most also need a user name.

Depending on the device additional options may be required.
Check the [list of supported devices](Supported-devices.md) to find out if this is the case for your device.
When using the graphical user interface, device-specific options can be configured in the "Advanced options" dropdown.

If any credentials (such as a password or passphrase) are required, you will be asked for them after the connection is started.

## SSH authentication

For devices using SSH, public key authentication is also supported.
By default, the application tries to use your OpenSSH private key and known hosts file.
This means that the connection should just work, as long as you connected to the device before using SSH.

Alternatively you may customize the SSH client options using the "Advanced options" dropdown in the graphical user interface.
Enter the path of a private key file or leave the field blank to disable public key authentication.
The path of the known hosts file is also configurable, but you cannot leave it blank as it is necessary to validate the host key of the device.
The special value `IGNORE` disables host key validation, but it is strongly recommended not to do this as it is insecure.

## Command line and web interface

As an alternative to the graphical user interface, you can run the application from the command line.
If you want to use the web interface, pass the `-web` option.

For information about available command line options, run `./dsl -help`.
Additional options may also be specified using a [configuration file](Configuration-files.md).

## Troubleshooting

In some cases additional steps may be required to get the application to run on your system:

- **Linux:**
  The GTK3 and WebKit2GTK 4.0 libraries need to be installed for the graphical user interface.
  On recent Linux distribution releases, the GTK4 variant using WebKitGTK 6.0 is more suitable.
- **Windows:**
  If SmartScreen blocks the application, you may need to click on "More info" and choose "Run anyway".
  Make sure that [Microsoft Edge WebView2](https://go.microsoft.com/fwlink/p/?LinkId=2124703) is installed if you want to run the graphical user interface.
- **macOS:**
  You may need to right-click the application and select "Open" to start it for the first time.
