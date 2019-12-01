package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/signal"

	log "github.com/Sirupsen/logrus"
	"github.com/tarm/serial"

	"github.com/yosssi/gmq/mqtt"
	"github.com/yosssi/gmq/mqtt/client"
)

var (
	mqttURL        string
	mqttLogin      string
	mqttPassword   string
	teleinfoDevice string
	cli            *client.Client
	sigc           chan os.Signal
)

type frame struct {
	tokens map[string]string
}

func getNextFrame(port *serial.Port) (*frame, error) {
	buffer := bufio.NewReader(port)
	_, err := buffer.ReadSlice(0x2)
	if err != nil {
		return nil, fmt.Errorf("Error looking for 0x2")
	}
	rawFrame, err := buffer.ReadBytes(0x3)
	if err != nil {
		return nil, fmt.Errorf("Error looking for 0x3")
	}
	if len(rawFrame) == 0 {
		return nil, fmt.Errorf("read empty frame")
	}
	rawFrame = bytes.Trim(rawFrame[0:len(rawFrame)-1], "\r\n")
	fields := bytes.Split(rawFrame, []byte("\r\n"))
	result := frame{
		tokens: map[string]string{},
	}
	for _, field := range fields {
		tokens := bytes.Split(field, []byte(" "))
		if len(tokens) != 3 {
			return nil, fmt.Errorf("invalid number of items => %s", field)
		}
		name, value, checksum := tokens[0], tokens[1], tokens[2]
		if len(checksum) != 1 {
			return nil, fmt.Errorf("invalide checksum: %s", checksum)
		}
		readChecksum := byte(checksum[0])
		expectedChecksum := computeChecksum(name, value)
		if readChecksum != expectedChecksum {
			return nil, fmt.Errorf("invalid checksum for (%s, %s). Should be %c, get %c", name, value, readChecksum, expectedChecksum)
		}
		result.tokens[string(name)] = string(value)
	}
	return &result, nil
}

func sum(a []byte) (res byte) {
	res = 0
	for _, c := range a {
		res += c
	}
	return
}

func computeChecksum(name []byte, value []byte) byte {
	// NOTE: 0x20 == ASCII space char
	checksum := sum(name) + byte(0x20) + sum(value)

	// Map to a single char E [0x20;0x7F]
	checksum = (checksum & 0x3F) + 0x20
	return checksum
}

func readFrames(port *serial.Port, frameChan chan<- *frame) {
	for {
		frame, err := getNextFrame(port)
		if err != nil {
			log.Printf("Error reading Teleinfo frame: %s\n", err)
			continue
		}
		frameChan <- frame
	}
}

func initMqtt() {
	// Create an MQTT Client.
	cli = client.New(&client.Options{
		// Define the processing of the error handler.
		ErrorHandler: func(err error) {
			log.Fatal(err)
		},
	})
	// Connect to the MQTT Server.
	err := cli.Connect(&client.ConnectOptions{
		Network:  "tcp",
		Address:  mqttURL,
		UserName: []byte(mqttLogin),
		Password: []byte(mqttPassword),
		ClientID: []byte("mqtt-teleinfo"),
	})
	if err != nil {
		log.Fatal(err)
	}

}

// send a message
func publish(topic, message string) error {
	// Publish a message.
	err := cli.Publish(&client.PublishOptions{
		QoS:       mqtt.QoS0,
		TopicName: []byte(topic),
		Message:   []byte(message),
	})
	if err != nil {
		log.Warn(err)
	}
	return err
}

func main() {
	// Set up channel on which to send signal notifications.
	sigc = make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill)
	if os.Getenv("DEBUG") != "" {
		log.SetLevel(log.DebugLevel)
	}
	mqttURL = os.Getenv("MQTT_URL")
	mqttLogin = os.Getenv("MQTT_LOGIN")
	mqttPassword = os.Getenv("MQTT_PASSWORD")
	teleinfoDevice = os.Getenv("TELEINFO_DEVICE")

	initMqtt()
	log.Info("Mqtt ... OK")
	defer cli.Terminate()

	cfg := &serial.Config{
		Name:     teleinfoDevice,
		Baud:     1200,
		Size:     7,
		Parity:   serial.ParityEven,
		StopBits: serial.Stop1,
	}
	port, err := serial.OpenPort(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer port.Close()

	frameChan := make(chan *frame)

	// Read Teleinfo frames and send them into framesChan
	go readFrames(port, frameChan)

	// Enqueue teleinfo.Frame into a fixed-length ring buffer
	go func() {
		for {
			frame := <-frameChan
			for token, value := range frame.tokens {
				publish(fmt.Sprintf("teleinfo/%s", token), value)
			}
		}
	}()

	<-sigc

	// Disconnect the Network Connection.
	if err := cli.Disconnect(); err != nil {
		panic(err)
	}
}
