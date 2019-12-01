# teleinfo2mqtt

Analyzes the teleinfo frames from serial device and sends the result to MQTT

    teleinfo/ADCO 012345678901
    teleinfo/OPTARIF BASE
    teleinfo/ISOUSC 30
    teleinfo/BASE 003404751
    teleinfo/PTEC TH..
    teleinfo/IINST 001
    teleinfo/HHPHC A
    teleinfo/IMAX 090
    teleinfo/PAPP 00270
    teleinfo/MOTDETAT 000000

## setup

The setup is done thanks to the following environment variables

	MQTT_URL
	MQTT_LOGIN
	MQTT_PASSWORD
	TELEINFO_DEVICE

The teleinfo device can be /dev/ttyUSB0

It has been tested successfully with a Micro Teleinfo v2.0 (https://www.tindie.com/products/hallard/micro-teleinfo-v20/)