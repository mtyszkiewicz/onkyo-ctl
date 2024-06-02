package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Profile struct {
	Name           string `json:"profile"`
	VolumeLevel    int    `json:"volumeLevel"`
	SubwooferLevel int    `json:"subwooferLevel"`
	MaxVolume      int    `json:"maxVolume"`
}

var profiles = map[string]Profile{
	"tv": {
		Name:           "tv",
		VolumeLevel:    20,
		SubwooferLevel: 0,
		MaxVolume:      28,
	},
	"dj": {
		Name:           "dj",
		VolumeLevel:    27,
		SubwooferLevel: -8,
		MaxVolume:      35,
	},
	"vinyl": {
		Name:           "vinyl",
		VolumeLevel:    20,
		SubwooferLevel: 0,
		MaxVolume:      30,
	},
	"spotify": {
		Name:           "spotify",
		VolumeLevel:    38,
		SubwooferLevel: -6,
		MaxVolume:      50,
	},
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

type EISCPClient struct {
	conn          net.Conn
	responseQueue chan string
}

func NewEISCPClient(host, port string) (*EISCPClient, error) {
	serverAddress := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", serverAddress, 5*time.Second)
	if err != nil {
		return nil, err
	}

	client := &EISCPClient{
		conn:          conn,
		responseQueue: make(chan string, 100),
	}
	go client.readResponses()
	return client, nil
}

func unpackEISCPMessage(packet string) string {
	if len(packet) < 16 {
		return packet
	}
	header := packet[:16]
	dataSize := binary.BigEndian.Uint32([]byte(header[8:12]))
	data := packet[16 : 16+dataSize]

	// Remove the ISCP start character '!1' and the trailing '\r'
	message := strings.TrimSpace(string(data[2 : len(data)-3]))
	return message
}

func (c *EISCPClient) readResponses() {
	buf := make([]byte, 1024)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			close(c.responseQueue)
			return
		}
		c.responseQueue <- string(buf[:n])
	}
}

func (c *EISCPClient) sendCommand(command string) error {
	// Clear the response queue
	for len(c.responseQueue) > 0 {
		<-c.responseQueue
	}

	packet := NewEISCPPacket(command)
	_, err := c.conn.Write(packet.Bytes())
	return err
}

func (c *EISCPClient) sendCommandAndGetResponse(command string) (string, error) {
	err := c.sendCommand(command)
	if err != nil {
		return "", err
	}

	select {
	case response := <-c.responseQueue:
		return unpackEISCPMessage(response), nil
	case <-time.After(2 * time.Second):
		return "", fmt.Errorf("timeout waiting for response")
	}
}

func (c *EISCPClient) PowerOn() error {
	return c.sendCommand("PWR01")
}

func (c *EISCPClient) PowerOff() error {
	return c.sendCommand("PWR00")
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
	response, err := c.sendCommandAndGetResponse("SLIQSTN")
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
	response, err := c.sendCommandAndGetResponse("MVLQSTN")
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
	response, err := c.sendCommandAndGetResponse("SWLQSTN")
	if err != nil {
		return 0, err
	}
	result, err := strconv.Atoi(strings.TrimPrefix(response, "SWL"))
	if err != nil {
		return 0, err
	}
	return result, nil
}

func main() {
	client, err := NewEISCPClient("10.205.0.163", "60128")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer client.conn.Close()
	fmt.Println("Connected to server")

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/power", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Power status: on"))
	})

	r.Put("/power/on", func(w http.ResponseWriter, r *http.Request) {
		if err := client.PowerOn(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Power turned on"))
	})

	r.Put("/power/off", func(w http.ResponseWriter, r *http.Request) {
		if err := client.PowerOff(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("Power turned off"))
	})

	r.Get("/volume", func(w http.ResponseWriter, r *http.Request) {
		volume, err := client.QueryCurrentVolume()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("Volume level: %d", volume)))
	})

	r.Put("/volume", func(w http.ResponseWriter, r *http.Request) {
		levelStr := r.URL.Query().Get("level")
		level, err := strconv.Atoi(levelStr)
		if err != nil || level < 0 || level > 50 {
			http.Error(w, "Invalid volume level", http.StatusBadRequest)
			return
		}
		if err := client.PowerOn(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := client.SetMasterVolume(level); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("Volume set to %d", level)))
	})

	r.Get("/subwoofer", func(w http.ResponseWriter, r *http.Request) {
		subwoofer, err := client.QueryCurrentSubwooferLevel()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("Subwoofer level: %d", subwoofer)))
	})

	r.Put("/subwoofer", func(w http.ResponseWriter, r *http.Request) {
		levelStr := r.URL.Query().Get("level")
		level, err := strconv.Atoi(levelStr)
		if err != nil || level < -8 || level > 8 {
			http.Error(w, "Invalid subwoofer level", http.StatusBadRequest)
			return
		}
		if err := client.PowerOn(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := client.SetSubwooferLevel(level); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("Subwoofer level set to %d", level)))
	})

	r.Get("/input", func(w http.ResponseWriter, r *http.Request) {
		input, err := client.QueryCurrentInput()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte(fmt.Sprintf("Current input: %s", input)))
	})

	r.Put("/input", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if err := client.PowerOn(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := client.SetInputSelector(name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte(fmt.Sprintf("Input set to %s", name)))
	})

	r.Get("/profile", func(w http.ResponseWriter, r *http.Request) {
		currentInput, err := client.QueryCurrentInput()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		currentVolume, err := client.QueryCurrentVolume()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		currentSubwoofer, err := client.QueryCurrentSubwooferLevel()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		profile, exists := profiles[currentInput]
		if !exists {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}

		response := Profile{
			Name:           currentInput,
			VolumeLevel:    currentVolume,
			SubwooferLevel: currentSubwoofer,
			MaxVolume:      profile.MaxVolume,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	r.Put("/profile", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		profile, exists := profiles[name]
		if !exists {
			http.Error(w, "Invalid profile name", http.StatusBadRequest)
			return
		}
		if err := client.PowerOn(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := client.SetMasterVolume(profile.VolumeLevel); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := client.SetSubwooferLevel(profile.SubwooferLevel); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := client.SetInputSelector(name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(profile)
	})

	http.ListenAndServe(":8080", r)
}

type EISCPPacket struct {
	Magic      [4]byte
	HeaderSize uint32
	DataSize   uint32
	Version    byte
	Reserved   [3]byte
	Data       []byte
}

func NewEISCPPacket(iscpMessage string) *EISCPPacket {
	iscpMessage = "!1" + iscpMessage + "\r"
	iscpMessageBytes := []byte(iscpMessage)
	packet := &EISCPPacket{
		Magic:      [4]byte{'I', 'S', 'C', 'P'},
		HeaderSize: 16,
		DataSize:   uint32(len(iscpMessageBytes)),
		Version:    0x01,
		Reserved:   [3]byte{0x00, 0x00, 0x00},
		Data:       iscpMessageBytes,
	}
	return packet
}

func (p *EISCPPacket) Bytes() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, p.Magic)
	binary.Write(buf, binary.BigEndian, p.HeaderSize)
	binary.Write(buf, binary.BigEndian, p.DataSize)
	binary.Write(buf, binary.BigEndian, p.Version)
	binary.Write(buf, binary.BigEndian, p.Reserved)
	buf.Write(p.Data)
	return buf.Bytes()
}
