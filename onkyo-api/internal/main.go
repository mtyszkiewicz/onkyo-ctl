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

func main() {
	client, err := eiscp.NewEISCPClient("10.205.0.163", "60128")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		return
	}
	defer client.Conn.Close()
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
