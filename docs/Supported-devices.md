# Supported devices

This is a list of supported device types.
Some devices are supported directly, in this case they are listed by the device or manufacturer name.
Others are supported by accessing diagnostic tools from the DSL modem chipset manufacturer.
So, if you can't find your device in the list, find out which kind of modem it uses and look for that instead.

For some devices, there are additional configuration options.
Check the respective device section for details, as these options may be required to connect successfully.
Each device section also includes an example for command line usage.

## Bintec Elmeg (manufacturer)

*Device types: `bintecelmeg_ssh`, `bintecelmeg_telnet`*

Requires access to the command line interface via SSH or Telnet.

	./dsl -d bintecelmeg_ssh -u admin 192.168.0.251
	./dsl -d bintecelmeg_telnet -u admin 192.168.0.251

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

For some error counters and transmit power information, TR-064 needs to be enabled on the device.

It is possible to access additional information that is not available from the web interface, such as QLN and Hlog data.
To do so, set the option `LoadSupportData` to `1`.
However, note that this will significantly increase loading times.

	./dsl -d fritzbox -o LoadSupportData=0 fritz.box

## LANCOM (manufacturer)

*Device type: `lancom_snmpv3`*

SNMP access needs to be configured on the device beforehand.
To do so, go to "Configuration > Logging/Monitoring > Protocols" and follow these steps:

1. Add view entries with access to the OID subtrees "1.3.6.1.4.1.2356.11.1.75", "1.3.6.1.4.1.2356.11.1.41", and "1.3.6.1.4.1.2356.11.1.99".
   Use the same name for all of them.
2. Create new access rights and select the just created view as read-only view.
   Configure the security model as "SNMPv3 (USM)" and choose the desired security level.
3. Create a new user with the desired security options.
4. Create a new group and select the just created access rights and user.
   Configure the security model as "SNMPv3 (USM)".

The options `AuthProtocol` and `PrivacyProtocol` need to be specified in the client, and must match the configuration on the device.

You may also set the `Subtree` option to choose which subtree is loaded from the device.
If no value is specified, both "/Status/VDSL" and "/Status/ADSL" will be tried.

	./dsl -d lancom_snmpv3 -u user -o AuthProtocol=sha -o PrivacyProtocol=aes256 172.23.56.254

## Lantiq, Infineon, Intel, MaxLinear (chipset vendor)

*Device types: `lantiq_ssh`, `lantiq_telnet`*

Requires command line access via SSH or Telnet.

Some FRITZ!Box devices are also supported, if a modified firmware with command line access is installed.
However, it is necessary to run `dsl_pipe ccadbgmls 13 ff` (the index 13 may be different) before connecting, as command output will be hidden otherwise.

On some devices, you may need to specify the actual name of the `dsl_cpe_pipe` command using the `Command` option.
If unspecified, a few common variants are tried.

	./dsl -d lantiq_ssh -o Command="dsl_cpe_pipe.sh" -u root openwrt.lan # OpenWrt
	./dsl -d lantiq_ssh -o Command="/usr/sbin/dsl_pipe" -u root fritz.box # FRITZ!Box (modified firmware)
	./dsl -d lantiq_telnet -o Command="/ifx/vdsl2/dsl_pipe" 192.168.16.249 # VINAX modem

## MediaTek, TrendChip, EcoNet, Airoha (chipset vendor)

*Device types: `mediatek_ssh`, `mediatek_telnet`*

Requires access to the Linux command line via SSH or Telnet.

Some of the data can only be read in a hacky way from the kernel log.
If any other messages are written to the kernel log at the same time, the data may not be parsed correctly.
This also means that connecting multiple clients to the same device won't work.

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
