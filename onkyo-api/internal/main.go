package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mtyszkiewicz/eiscp/internal/pkg/eiscp"
)

type Profile struct {
	Name           string `json:"profile"`
	VolumeLevel    int    `json:"volumeLevel"`
	SubwooferLevel int    `json:"subwooferLevel"`
	MaxVolume      int    `json:"maxVolume"`
}

type Server struct {
	client   *eiscp.EISCPClient
	profiles map[string]Profile
}

func NewServer(client *eiscp.EISCPClient) *Server {
	return &Server{
		client: client,
		profiles: map[string]Profile{
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
		},
	}
}

func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/power", func(r chi.Router) {
		r.Get("/", s.getPowerStatus)
		r.Put("/on", s.powerOn)
		r.Put("/off", s.powerOff)
	})

	r.Route("/volume", func(r chi.Router) {
		r.Get("/", s.getVolume)
		r.Put("/", s.setVolume)
	})

	r.Route("/subwoofer", func(r chi.Router) {
		r.Get("/", s.getSubwoofer)
		r.Put("/", s.setSubwoofer)
	})

	r.Route("/input", func(r chi.Router) {
		r.Get("/", s.getInput)
		r.Put("/", s.setInput)
	})

	r.Route("/profile", func(r chi.Router) {
		r.Get("/", s.getProfile)
		r.Put("/", s.setProfile)
	})

	return r
}

// Power handlers
func (s *Server) getPowerStatus(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Power status: on"))
}

func (s *Server) powerOn(w http.ResponseWriter, r *http.Request) {
	if err := s.client.PowerOn(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Power turned on"))
}

func (s *Server) powerOff(w http.ResponseWriter, r *http.Request) {
	if err := s.client.PowerOff(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("Power turned off"))
}

// Volume handlers
func (s *Server) getVolume(w http.ResponseWriter, r *http.Request) {
	volume, err := s.client.QueryCurrentVolume()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("Volume level: %d", volume)))
}

func (s *Server) setVolume(w http.ResponseWriter, r *http.Request) {
	levelStr := r.URL.Query().Get("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < 0 || level > 50 {
		http.Error(w, "Invalid volume level", http.StatusBadRequest)
		return
	}
	if err := s.client.PowerOn(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.client.SetMasterVolume(level); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("Volume set to %d", level)))
}

// Subwoofer handlers
func (s *Server) getSubwoofer(w http.ResponseWriter, r *http.Request) {
	subwoofer, err := s.client.QueryCurrentSubwooferLevel()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("Subwoofer level: %d", subwoofer)))
}

func (s *Server) setSubwoofer(w http.ResponseWriter, r *http.Request) {
	levelStr := r.URL.Query().Get("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil || level < -8 || level > 8 {
		http.Error(w, "Invalid subwoofer level", http.StatusBadRequest)
		return
	}
	if err := s.client.PowerOn(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.client.SetSubwooferLevel(level); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("Subwoofer level set to %d", level)))
}

// Input handlers
func (s *Server) getInput(w http.ResponseWriter, r *http.Request) {
	input, err := s.client.QueryCurrentInput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(fmt.Sprintf("Current input: %s", input)))
}

func (s *Server) setInput(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if err := s.client.PowerOn(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.client.SetInputSelector(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Write([]byte(fmt.Sprintf("Input set to %s", name)))
}

// Profile handlers
func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	currentInput, err := s.client.QueryCurrentInput()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	currentVolume, err := s.client.QueryCurrentVolume()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	currentSubwoofer, err := s.client.QueryCurrentSubwooferLevel()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	profile, exists := s.profiles[currentInput]
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
}

func (s *Server) setProfile(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	profile, exists := s.profiles[name]
	if !exists {
		http.Error(w, "Invalid profile name", http.StatusBadRequest)
		return
	}
	if err := s.client.PowerOn(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.client.SetMasterVolume(profile.VolumeLevel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.client.SetSubwooferLevel(profile.SubwooferLevel); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.client.SetInputSelector(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func main() {
	client, err := eiscp.NewEISCPClient("10.205.0.163", "60128")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer client.Conn.Close()
	fmt.Println("Connected to server")

	server := NewServer(client)
	http.ListenAndServe(":8080", server.Routes())
}
