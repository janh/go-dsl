# xDSL Stats Parser

This is a library and command line tool for reading basic xDSL stats, as well bitloading, SNR, QLN and Hlog data. It supports many different kinds of modems.

Please note that this is unfinished software at the moment. The command line tool works, but the API of the library is not final yet and there will be incompatible changes.

## Building

Go needs to be installed. Run the following command to build the command line tool:

	go build -o dsl cmd/main.go

Note: If you are using Windows, use the option `-o dsl.exe` instead, and adjust the commands listed below as necessary.

## Usage

A few usage examples for different modems are listed below.

In general, if the device requires a username, you need to specify it on the command line using the `-u` option. When connecting using SSH, your OpenSSH private key and `known_hosts` file are tried by default, but it is possible to override this. Otherwise, if the device uses password authentication, you will be asked for it.

By default, the output contains a summary of the loaded data, and additional data including graphs is written to files. Alternatively, a web interface can be started using the `-web` option, and a link to access it is shown. The data on the web interface is automatically refreshed.

To get more detailed usage information, see `./dsl -help`.

### Broadcom

	./dsl -d broadcom_ssh -u root 192.168.1.1
	./dsl -d broadcom_telnet -u root 192.168.1.1

If the command on the device is not named `xdslctl`, you need to specify the correct name using the `-o Commmand` option.

### DrayTek

	./dsl -d draytek_telnet 192.168.1.1

While there is an SSH server on DrayTek devices, it is not supported at the moment, as the Go SSH client is incompatible with the tested device due to an unsupported DSA key size.

### FRITZ!Box

	./dsl -d fritzbox -o LoadSupportData=0 fritz.box

The option `-o LoadSupportData` is optional, and it is off by default. Set it to `1` to enable loading additional information from the support data, such as QLN and Hlog data. This will take significantly longer to load.

### Lantiq

	./dsl -d lantiq_ssh -o Command="dsl_cpe_pipe.sh" -u root openwrt.lan # OpenWrt
	./dsl -d lantiq_telnet -o Command="/ifx/vdsl2/dsl_pipe" 192.168.16.249 # VINAX modem

If no command is specified, `dsl_cpe_pipe` is used. Since the actual name varies a lot on different devices, you'll have to find out the right one for your device.

### MediaTek

	./dsl -d mediatek_ssh -u admin 192.168.1.1
	./dsl -d mediatek_telnet -u admin 192.168.1.1

### Sagemcom

	./dsl -d sagemcom speedport.ip

There tend to be data quality issues on at least some of the supported devices. For example, only the "Actual Data Rate" is reported instead of the "Actual Net Data Rate". This means that the reported data rate is not comparable with other devices when retransmission is active.

### Speedport

	./dsl -d speedport speedport.ip

Only a limited set of data is available on these devices. This is known to work with a Speedport Smart 2, but it may also work for other devices. For Speedport Pro, use the Sagemcom client.
