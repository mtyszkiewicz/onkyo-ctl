package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mtyszkiewicz/eiscp/internal/config"
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

func NewServer(client *eiscp.EISCPClient, profiles map[string]Profile) *Server {
	return &Server{
		client:   client,
		profiles: profiles,
	}
}

func (s *Server) Routes() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger, middleware.Recoverer)

	r.Route("/power", func(r chi.Router) {
		r.Get("/", s.getPowerStatus)
		r.Put("/on", s.powerOn)
		r.Put("/off", s.powerOff)
	})

	r.Route("/volume", func(r chi.Router) {
		r.Get("/", s.getVolume)
		r.Put("/", s.setVolume)
		r.Put("/up", s.volumeUp)
		r.Put("/down", s.volumeDown)
	})

	r.Route("/subwoofer", func(r chi.Router) {
		r.Get("/", s.getSubwoofer)
		r.Put("/", s.setSubwoofer)
		r.Put("/up", s.subwooferUp)
		r.Put("/down", s.subwooferDown)
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

func handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, eiscp.ErrValidation):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, eiscp.ErrTimeout):
		http.Error(w, err.Error(), http.StatusGatewayTimeout)
	case errors.Is(err, eiscp.ErrConnection), errors.Is(err, eiscp.ErrTransport):
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) getPowerStatus(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Power status: on")
}

func (s *Server) powerOn(w http.ResponseWriter, r *http.Request) {
	if err := s.client.PowerOn(); err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprint(w, "Power turned on")
}

func (s *Server) powerOff(w http.ResponseWriter, r *http.Request) {
	if err := s.client.PowerOff(); err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprint(w, "Power turned off")
}

func (s *Server) getVolume(w http.ResponseWriter, r *http.Request) {
	volume, err := s.client.QueryVolume()
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "Volume level: %d", volume)
}

func (s *Server) volumeUp(w http.ResponseWriter, r *http.Request) {
	if err := s.client.VolumeUp(); err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprint(w, "Volume level: Up")
}

func (s *Server) volumeDown(w http.ResponseWriter, r *http.Request) {
	if err := s.client.VolumeDown(); err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprint(w, "Volume level: Down")
}

func (s *Server) setVolume(w http.ResponseWriter, r *http.Request) {
	levelStr := r.URL.Query().Get("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		handleError(w, fmt.Errorf("%w: invalid volume level format", eiscp.ErrValidation))
		return
	}

	if err := s.client.PowerOn(); err != nil {
		handleError(w, err)
		return
	}

	if err := s.client.SetMasterVolume(level); err != nil {
		handleError(w, err)
		return
	}

	fmt.Fprintf(w, "Volume set to %d", level)
}

func (s *Server) getSubwoofer(w http.ResponseWriter, r *http.Request) {
	level, err := s.client.QuerySubwooferLevel()
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "Subwoofer level: %d", level)
}

func (s *Server) setSubwoofer(w http.ResponseWriter, r *http.Request) {
	levelStr := r.URL.Query().Get("level")
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		handleError(w, fmt.Errorf("%w: invalid subwoofer level format", eiscp.ErrValidation))
		return
	}

	if err := s.client.PowerOn(); err != nil {
		handleError(w, err)
		return
	}

	if err := s.client.SetSubwooferLevel(level); err != nil {
		handleError(w, err)
		return
	}

	fmt.Fprintf(w, "Subwoofer level set to %d", level)
}

func (s *Server) subwooferUp(w http.ResponseWriter, r *http.Request) {
	if err := s.client.SubwooferUp(); err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprint(w, "Subwoofer level: Up")
}

func (s *Server) subwooferDown(w http.ResponseWriter, r *http.Request) {
	if err := s.client.SubwooferDown(); err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprint(w, "Subwoofer level: Down")
}

func (s *Server) getInput(w http.ResponseWriter, r *http.Request) {
	input, err := s.client.QueryInputSelector()
	if err != nil {
		handleError(w, err)
		return
	}
	fmt.Fprintf(w, "Current input: %s", input)
}

func (s *Server) setInput(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		handleError(w, fmt.Errorf("%w: input name cannot be empty", eiscp.ErrValidation))
		return
	}

	if err := s.client.PowerOn(); err != nil {
		handleError(w, err)
		return
	}

	if err := s.client.SetInputSelector(name); err != nil {
		handleError(w, err)
		return
	}

	fmt.Fprintf(w, "Input set to %s", name)
}

func (s *Server) getProfile(w http.ResponseWriter, r *http.Request) {
	currentInput, err := s.client.QueryInputSelector()
	if err != nil {
		handleError(w, err)
		return
	}

	currentVolume, err := s.client.QueryVolume()
	if err != nil {
		handleError(w, err)
		return
	}

	currentSubwoofer, err := s.client.QuerySubwooferLevel()
	if err != nil {
		handleError(w, err)
		return
	}

	profile, exists := s.profiles[currentInput]
	if !exists {
		handleError(w, fmt.Errorf("%w: profile not found for input '%s'", eiscp.ErrValidation, currentInput))
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
		handleError(w, fmt.Errorf("%w: profile '%s' does not exist", eiscp.ErrValidation, name))
		return
	}

	if err := s.client.PowerOn(); err != nil {
		handleError(w, err)
		return
	}

	if err := s.client.SetMasterVolume(profile.VolumeLevel); err != nil {
		handleError(w, err)
		return
	}

	if err := s.client.SetSubwooferLevel(profile.SubwooferLevel); err != nil {
		handleError(w, err)
		return
	}

	if err := s.client.SetInputSelector(name); err != nil {
		handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func main() {
	cfg, err := config.Load(os.Getenv("ONKYO_CONFIG"))
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	client, err := eiscp.NewEISCPClient(cfg.Onkyo.Host, cfg.Onkyo.Port, cfg.InputCodes(), cfg.InputNames())
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer client.Conn.Close()

	profiles := make(map[string]Profile, len(cfg.Profiles))
	for name, p := range cfg.Profiles {
		profiles[name] = Profile{
			Name:           name,
			VolumeLevel:    p.VolumeLevel,
			SubwooferLevel: p.SubwooferLevel,
			MaxVolume:      p.MaxVolume,
		}
	}

	log.Println("Connected to server")
	server := NewServer(client, profiles)
	log.Fatal(http.ListenAndServe(":8080", server.Routes()))
}