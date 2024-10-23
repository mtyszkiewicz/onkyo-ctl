package eiscp

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type EISCPClient struct {
	Conn          net.Conn
	responseQueue chan string
}

func NewEISCPClient(host, port string) (*EISCPClient, error) {
	serverAddress := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", serverAddress, 5*time.Second)
	if err != nil {
		return nil, err
	}
	client := &EISCPClient{
		Conn:          conn,
		responseQueue: make(chan string, 100),
	}
	go client.readResponses()
	return client, nil
}

// Constatnly puts incoming data into responseQueue
func (c *EISCPClient) readResponses() {
	buf := make([]byte, 1024)
	for {
		n, err := c.Conn.Read(buf)
		if err != nil {
			close(c.responseQueue)
			return
		}
		c.responseQueue <- string(buf[:n])
	}
}

// Sends ISCP message and returns without awaiting the response
func (c *EISCPClient) sendCommand(msg string) error {
	// Clear the response queue
	for len(c.responseQueue) > 0 {
		<-c.responseQueue
	}

	packet := NewEISCPPacket(msg)
	_, err := c.Conn.Write(packet.Bytes())
	return err
}

// Sends ISCP message and waits for response
func (c *EISCPClient) sendReceiveCommand(command string) (string, error) {
	err := c.sendCommand(command)
	if err != nil {
		return "", err
	}

	select {
	case response := <-c.responseQueue:
		return UnpackEISCPMessage(response), nil
	case <-time.After(2 * time.Second):
		return "", fmt.Errorf("timeout waiting for response")
	}
}

var inputCodes = map[string]string{
	"spotify": "01",
	"vinyl":   "22",
	"tv":      "12",
	"dj":      "10",
}

var inputNames = map[string]string{
	"01": "spotify",
	"22": "vinyl",
	"12": "tv",
	"10": "dj",
}

func (c *EISCPClient) PowerOn() error {
	return c.sendCommand("PWR01")
}

func (c *EISCPClient) PowerOff() error {
	return c.sendCommand("PWR00")
}

func (c *EISCPClient) VolumeUp() error {
	return c.sendCommand("MVLUP")
}

func (c *EISCPClient) VolumeDown() error {
	return c.sendCommand("MVLDOWN")
}

func (c *EISCPClient) SetMasterVolume(level int) error {
	if level < 0 || level > 50 {
		return fmt.Errorf("invalid volume level: %d, must be between 0 and 50", level)
	}
	hexLevel := fmt.Sprintf("%02X", level)
	return c.sendCommand("MVL" + hexLevel)
}

func (c *EISCPClient) SetSubwooferLevel(level int) error {
	if level < -8 || level > 8 {
		return fmt.Errorf("invalid subwoofer level: %d, must be between -8 and 8", level)
	}

	var command string
	if level >= 0 {
		command = fmt.Sprintf("SWL+%02d", level)
	} else {
		command = fmt.Sprintf("SWL-%02d", -level)
	}

	return c.sendCommand(command)
}

func (c *EISCPClient) SetInputSelector(input string) error {
	code, ok := inputCodes[input]
	if !ok {
		return fmt.Errorf("invalid input selector: %s", input)
	}
	return c.sendCommand("SLI" + code)
}

func (c *EISCPClient) QueryCurrentInput() (string, error) {
	response, err := c.sendReceiveCommand("SLIQSTN")
	if err != nil {
		return "", err
	}

	code := strings.TrimPrefix(response, "SLI")
	// fmt.Printf("Received input code: '%s' (length: %d)\n", code, len(code))

	name, ok := inputNames[code]
	if !ok {
		return "", fmt.Errorf("unknown input code: %s", code)
	}
	return name, nil
}

func (c *EISCPClient) QueryCurrentVolume() (int, error) {
	response, err := c.sendReceiveCommand("MVLQSTN")
	if err != nil {
		return 0, err
	}

	hexValue := strings.TrimPrefix(response, "MVL")

	result, err := strconv.ParseInt(hexValue, 16, 64)
	if err != nil {
		return 0, err
	}

	return int(result), nil
}

func (c *EISCPClient) QueryCurrentSubwooferLevel() (int, error) {
	response, err := c.sendReceiveCommand("SWLQSTN")
	if err != nil {
		return 0, err
	}
	result, err := strconv.Atoi(strings.TrimPrefix(response, "SWL"))
	if err != nil {
		return 0, err
	}
	return result, nil
}
