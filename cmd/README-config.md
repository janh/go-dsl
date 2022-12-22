# Configuration files

The applications uses two different config files for the main configuration and secrets.
Both config files use the TOML format, supported options are described in the sections below.

## Main configuration

Most options are configured in the main config file.
By default it is loaded from the user directory with the exact location depending on the operating system.
Alternatively, a path can be specified using the `-config` command line option.
This file is also used to persist configuration from the graphical user interface.

- **DeviceType**:  
  Which kind of device to connect to, use the `-help` command line option to see a list of available options.  
  *(equivalent to the `-d` command line option)*

- **Host**:  
  Host name to connect to.  
  *(equivalent to the last argument on the command line)*

- **User**:  
  User name to use if required for the selected device type.  
  *(equivalent to `-u` command line option)*

- **PrivateKeyPath**:  
  Path to private key file or directory containing private keys for SSH authentication.  
  *(equivalent to `-private-key` command line option)*

- **KnownHostsPath**:  
  Path to known hosts file for SSH host key validation.
  Validation is skipped if the special value "IGNORE" is specified.  
  *(equivalent to `-known-hosts` command line option)*

### Options table

All device-specific options are specified in a table called **Option**, which is equivalent to the `-o` command line options.
The value always needs to be a quoted string.
For details about the available options, run the application with the `-help` option.

### Web table

The **Web** table contains options for the web server.
There are no equivalent command line options for these settings.

- **ListenAddress**:  
  Address for the web server to listen on.
  If unspecified or empty, the server will listen at a random port on localhost.

- **HideErrorMessages**:  
  Only show generic error messages in the web interface and print a log message with the original message instead.
  Recommended when the web server is publicly exposed, to avoid access to sensitive information.

- **DisableInteractiveAuth**:  
  Disable password and passphrase entry in the web interface.
  All necessary authentication data needs to be specified in the configuration or secrets file instead.
  Recommended when the web server is publicly exposed, to reduce attack surface.

- **HideRawData**:  
  Make raw data inaccessible.
  Depending on the device type, this may be useful to prevent access to sensitive information.

## Secrets configuration

A separate file can be used for secrets such as passwords or passphrases.
It is only used if its is specified on the command line using the `-secrets` option.
Any non-empty value means that the application won't ask for that type of secret interactively.

- **Password**:  
  Password to use for authentication, if requested by the device.

- **PrivateKeyPassphrase**:  
  Passphrase to use for decryption of SSH private key, if required.
