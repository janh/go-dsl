# Supported devices

This is a list of supported device types.
Some devices are supported directly, in this case they are listed by the device or manufacturer name.
Others are supported by accessing diagnostic tools from the DSL modem chipset manufacturer.
So, if you can't find your device in the list, find out which kind of modem it uses and look for that instead.

For some devices, there are additional configuration options.
Check the respective device section for details, as these options may be required to connect successfully.
Each device section also includes an example for command line usage.

## Broadcom (chipset vendor)

*Device types: `broadcom_ssh`, `broadcom_telnet`*

Requires access to the system command line via SSH or Telnet.

If the command on the device is not named `xdslctl`, you need to specify the correct name using the `Commmand` option.

	./dsl -d broadcom_ssh -u root 192.168.1.1
	./dsl -d broadcom_telnet -u root 192.168.1.1

## DrayTek (manufacturer)

*Device type: `draytek_telnet`*

Only Telnet is supported for connection to the device, as the SSH server suffers from various issues depending on the device (such as outdated cipher suites or crashes).

	./dsl -d draytek_telnet 192.168.1.1

## FRITZ!Box (device)

*Device type: `fritzbox`*

Should work with any DSL FRITZ!Box running a somewhat recent firmware (but for now requires the web interface language being set to German).

It is possible to access additional information that is not available from the web interface, such as QLN and Hlog data.
To do so, set the option `LoadSupportData` to `1`.
However, note that this will significantly increase loading times.

	./dsl -d fritzbox -o LoadSupportData=0 fritz.box

## Lantiq, Infineon, Intel, MaxLinear (chipset vendor)

*Device types: `lantiq_ssh`, `lantiq_telnet`*

Requires command line access via SSH or Telnet.

You'll likely need to specify the `Command` option.
If unspecified, `dsl_cpe_pipe` is used, but the actual name of the command varies between devices.

	./dsl -d lantiq_ssh -o Command="dsl_cpe_pipe.sh" -u root openwrt.lan # OpenWrt
	./dsl -d lantiq_telnet -o Command="/ifx/vdsl2/dsl_pipe" 192.168.16.249 # VINAX modem

## MediaTek, TrendChip, EcoNet, Airoha (chipset vendor)

*Device types: `mediatek_ssh`, `mediatek_telnet`*

Requires access to the Linux command line via SSH or Telnet.

	./dsl -d mediatek_ssh -u admin 192.168.1.1
	./dsl -d mediatek_telnet -u admin 192.168.1.1

## Sagemcom (manufacturer)

*Device type: `sagemcom`*

Unfortunately, there are data quality issues at least on some devices or firmware versions.
For example, it is possible that only the "Actual Data Rate" is reported instead of the "Actual Net Data Rate".
This means that the reported data rates are not comparable with other devices when retransmission is active.

	./dsl -d sagemcom speedport.ip

## Speedport (device)

*Device type: `speedport`*

Tested with Speedport Smart 2, other devices may or may not work.
For the Speedport Pro series, use the Sagemcom client instead.

Only a limited set of data is available, most notably SNR, QLN and Hlog data are missing.

	./dsl -d speedport speedport.ip
