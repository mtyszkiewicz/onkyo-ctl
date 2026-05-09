package eiscp

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Custom error types
var (
	ErrValidation = errors.New("validation error")
	ErrTimeout    = errors.New("timeout error")
	ErrConnection = errors.New("connection error")
	ErrTransport  = errors.New("transport error")
)

type EISCPClient struct {
	Conn          net.Conn
	responseQueue chan string
	inputCodes    map[string]string
	inputNames    map[string]string
}

func NewEISCPClient(host, port string, inputCodes, inputNames map[string]string) (*EISCPClient, error) {
	serverAddress := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", serverAddress, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnection, err)
	}
	client := &EISCPClient{
		Conn:          conn,
		responseQueue: make(chan string, 100),
		inputCodes:    inputCodes,
		inputNames:    inputNames,
	}
	go client.listen()
	return client, nil
}

// Constatnly puts incoming data into responseQueue
func (c *EISCPClient) listen() {
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
func (c *EISCPClient) SendCommand(msg string) error {
	// Clear the response queue
	for len(c.responseQueue) > 0 {
		<-c.responseQueue
	}

	packet := NewEISCPPacket(msg)
	_, err := c.Conn.Write(packet.Bytes())
	if err != nil {
		return fmt.Errorf("%w: %v", ErrTransport, err)
	}
	return nil
}

// Sends ISCP message and waits for response
func (c *EISCPClient) SendReceiveCommand(command string) (string, error) {
	err := c.SendCommand(command)
	if err != nil {
		return "", err
	}

	select {
	case response := <-c.responseQueue:
		return UnpackEISCPMessage(response), nil
	case <-time.After(2 * time.Second):
		return "", fmt.Errorf("%w: no response received within timeout", ErrTimeout)
	}
}

func (c *EISCPClient) SetInputSelector(input string) error {
	code, ok := c.inputCodes[input]
	if !ok {
		return fmt.Errorf("%w: invalid input selector '%s'", ErrValidation, input)
	}
	return c.SendCommand("SLI" + code)
}

func (c *EISCPClient) QueryInputSelector() (string, error) {
	response, err := c.SendReceiveCommand("SLIQSTN")
	if err != nil {
		return "", err
	}

	code := strings.TrimPrefix(response, "SLI")

	name, ok := c.inputNames[code]
	if !ok {
		return "", fmt.Errorf("%w: unknown input code '%s'", ErrValidation, code)
	}
	return name, nil
}

func (c *EISCPClient) ListInputs() []string {
	names := make([]string, 0, len(c.inputCodes))
	for name := range c.inputCodes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (c *EISCPClient) PowerOn() error {
	return c.SendCommand("PWR01")
}

func (c *EISCPClient) PowerOff() error {
	return c.SendCommand("PWR00")
}

func (c *EISCPClient) VolumeUp() error {
	return c.SendCommand("MVLUP")
}

func (c *EISCPClient) VolumeDown() error {
	return c.SendCommand("MVLDOWN")
}

func (c *EISCPClient) SubwooferUp() error {
	return c.SendCommand("SWLUP")
}

func (c *EISCPClient) SubwooferDown() error {
	return c.SendCommand("SWLDOWN")
}

func (c *EISCPClient) SetMasterVolume(level int) error {
	if level < 0 || level > 50 {
		return fmt.Errorf("%w: volume level %d must be between 0 and 50", ErrValidation, level)
	}
	hexLevel := fmt.Sprintf("%02X", level)
	return c.SendCommand("MVL" + hexLevel)
}

func (c *EISCPClient) SetSubwooferLevel(level int) error {
	if level < -8 || level > 8 {
		return fmt.Errorf("%w: subwoofer level %d must be between -8 and 8", ErrValidation, level)
	}

	var command string
	if level >= 0 {
		command = fmt.Sprintf("SWL+%02d", level)
	} else {
		command = fmt.Sprintf("SWL-%02d", -level)
	}

	return c.SendCommand(command)
}

func (c *EISCPClient) QueryVolume() (int, error) {
	response, err := c.SendReceiveCommand("MVLQSTN")
	if err != nil {
		return 0, err
	}

	hexValue := strings.TrimPrefix(response, "MVL")

	result, err := strconv.ParseInt(hexValue, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to parse volume response", ErrTransport)
	}

	return int(result), nil
}

func (c *EISCPClient) QuerySubwooferLevel() (int, error) {
	response, err := c.SendReceiveCommand("SWLQSTN")
	if err != nil {
		return 0, err
	}
	response = strings.TrimPrefix(response, "SWL")
	response = strings.TrimSuffix(response, "C")
	result, err := strconv.Atoi(response)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to parse subwoofer response", ErrTransport)
	}
	return result, nil
}

func (c *EISCPClient) SetBrightness(level int) error {
	if !(level == 0 || level == 1 || level == 2) {
		return fmt.Errorf("%w: brightness level must be either: 0 - bright, 1 - dim, 2 - dark", ErrValidation)
	}
	return c.SendCommand(fmt.Sprintf("DIM0%d", level))
}

func (c *EISCPClient) AnimateBlink() error {
	var err error

	err = c.SendCommand("DIM01")
	if err != nil {
		return fmt.Errorf("failed to set brightness: %w", err)
	}

	time.Sleep(60 * time.Millisecond)
	err = c.SendCommand("DIM00")
	if err != nil {
		return fmt.Errorf("failed to set brightness: %w", err)
	}

	time.Sleep(80 * time.Millisecond)
	err = c.SendCommand("DIM01")
	if err != nil {
		return fmt.Errorf("failed to set brightness: %w", err)
	}

	time.Sleep(40 * time.Millisecond)
	err = c.SendCommand("DIM02")
	if err != nil {
		return fmt.Errorf("failed to set brightness: %w", err)
	}

	return nil
}
