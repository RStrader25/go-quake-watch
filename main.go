package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

type Earthquake struct {
	ID        string    `json:"id"`
	Magnitude float64   `json:"magnitude"`
	Place     string    `json:"place"`
	Time      time.Time `json:"time"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Depth     float64   `json:"depth"`
	URL       string    `json:"url"`
	Alert     string    `json:"alert"`
}

type USGSResponse struct {
	Features []struct {
		ID         string `json:"id"`
		Properties struct {
			Mag   float64 `json:"mag"`
			Place string  `json:"place"`
			Time  int64   `json:"time"`
			URL   string  `json:"url"`
			Alert string  `json:"alert"`
		} `json:"properties"`
		Geometry struct {
			Coordinates []float64 `json:"coordinates"`
		} `json:"geometry"`
	} `json:"features"`
}

var (
	earthquakeCache []Earthquake
	lastUpdate      time.Time
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Routes
	r.Get("/", serveHome)
	r.Get("/api/earthquakes", getEarthquakes)
	r.Get("/api/stream", streamEarthquakes)

	// Initial fetch
	fetchEarthquakes()

	// Background updater
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fetchEarthquakes()
		}
	}()

	log.Printf("🌍 Earthquake Monitor running on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func fetchEarthquakes() {
	url := "https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson"

	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching earthquakes: %v", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	var usgsData USGSResponse
	if err := json.Unmarshal(body, &usgsData); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return
	}

	earthquakes := make([]Earthquake, 0)
	for _, feature := range usgsData.Features {
		if feature.Properties.Mag > 0 {
			eq := Earthquake{
				ID:        feature.ID,
				Magnitude: feature.Properties.Mag,
				Place:     feature.Properties.Place,
				Time:      time.Unix(feature.Properties.Time/1000, 0),
				Longitude: feature.Geometry.Coordinates[0],
				Latitude:  feature.Geometry.Coordinates[1],
				Depth:     feature.Geometry.Coordinates[2],
				URL:       feature.Properties.URL,
				Alert:     feature.Properties.Alert,
			}
			earthquakes = append(earthquakes, eq)
		}
	}

	earthquakeCache = earthquakes
	lastUpdate = time.Now()
	log.Printf("Updated: %d earthquakes fetched", len(earthquakes))
}

func getEarthquakes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"earthquakes": earthquakeCache,
		"count":       len(earthquakeCache),
		"lastUpdate":  lastUpdate,
	})
}

func streamEarthquakes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			data, _ := json.Marshal(map[string]interface{}{
				"earthquakes": earthquakeCache,
				"count":       len(earthquakeCache),
				"timestamp":   time.Now(),
			})
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlTemplate))
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>🌍 Earthquake Monitor</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <link rel="stylesheet" href="https://unpkg.com/leaflet@1.9.4/dist/leaflet.css" />
    <script src="https://unpkg.com/leaflet@1.9.4/dist/leaflet.js"></script>
</head>
<body class="bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 min-h-screen text-white">
    <div class="container mx-auto px-4 py-8">
        <!-- Header -->
        <div class="text-center mb-8">
            <h1 class="text-5xl font-bold mb-2">🌍 Earthquake Monitor</h1>
            <p class="text-slate-400">Real-time seismic activity powered by USGS</p>
            <div class="mt-4 inline-flex items-center gap-2 bg-green-500/20 px-4 py-2 rounded-full">
                <div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
                <span class="text-sm text-green-400" id="status">Live</span>
            </div>
        </div>

        <!-- Stats -->
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4 mb-8">
            <div class="bg-slate-800/50 backdrop-blur rounded-lg p-6 border border-slate-700">
                <div class="text-slate-400 text-sm mb-1">Total Earthquakes</div>
                <div class="text-3xl font-bold" id="totalCount">0</div>
            </div>
            <div class="bg-slate-800/50 backdrop-blur rounded-lg p-6 border border-slate-700">
                <div class="text-slate-400 text-sm mb-1">Strongest (M)</div>
                <div class="text-3xl font-bold text-red-400" id="strongest">0.0</div>
            </div>
            <div class="bg-slate-800/50 backdrop-blur rounded-lg p-6 border border-slate-700">
                <div class="text-slate-400 text-sm mb-1">Last Update</div>
                <div class="text-xl font-bold" id="lastUpdate">-</div>
            </div>
        </div>

        <!-- Map -->
        <div class="bg-slate-800/50 backdrop-blur rounded-lg p-4 border border-slate-700 mb-8">
            <div id="map" class="w-full h-96 rounded-lg"></div>
        </div>

        <!-- Earthquake List -->
        <div class="bg-slate-800/50 backdrop-blur rounded-lg p-6 border border-slate-700">
            <h2 class="text-2xl font-bold mb-4">Recent Activity</h2>
            <div id="earthquakeList" class="space-y-3"></div>
        </div>
    </div>

    <script>
        let map;
        let markers = [];

        // Initialize map
        function initMap() {
            map = L.map('map').setView([20, 0], 2);
            L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
                attribution: '© OpenStreetMap contributors'
            }).addTo(map);
        }

        function getMagnitudeColor(mag) {
            if (mag >= 6) return 'bg-red-500';
            if (mag >= 4.5) return 'bg-orange-500';
            if (mag >= 3) return 'bg-yellow-500';
            return 'bg-green-500';
        }

        function getMarkerColor(mag) {
            if (mag >= 6) return 'red';
            if (mag >= 4.5) return 'orange';
            if (mag >= 3) return 'yellow';
            return 'green';
        }

        function updateMap(earthquakes) {
            markers.forEach(marker => map.removeLayer(marker));
            markers = [];

            earthquakes.forEach(eq => {
                const color = getMarkerColor(eq.magnitude);
                const marker = L.circleMarker([eq.latitude, eq.longitude], {
                    radius: eq.magnitude * 2,
                    fillColor: color,
                    color: '#fff',
                    weight: 1,
                    opacity: 1,
                    fillOpacity: 0.6
                }).addTo(map);

                marker.bindPopup(
                    '<b>M ' + eq.magnitude.toFixed(1) + '</b><br>' +
                    eq.place + '<br>' +
                    new Date(eq.time).toLocaleString()
                );
                markers.push(marker);
            });
        }

        function updateUI(data) {
            const earthquakes = data.earthquakes || [];
            
            // Update stats
            document.getElementById('totalCount').textContent = earthquakes.length;
            
            const strongest = earthquakes.length > 0 
                ? Math.max(...earthquakes.map(e => e.magnitude))
                : 0;
            document.getElementById('strongest').textContent = strongest.toFixed(1);
            
            document.getElementById('lastUpdate').textContent = 
                new Date().toLocaleTimeString();

            // Update map
            updateMap(earthquakes);

            // Update list
            const list = document.getElementById('earthquakeList');
            if (earthquakes.length === 0) {
                list.innerHTML = '<p class="text-slate-400">No earthquakes in the last hour</p>';
                return;
            }

            list.innerHTML = earthquakes
                .sort((a, b) => b.magnitude - a.magnitude)
                .map(eq => {
                    const color = getMagnitudeColor(eq.magnitude);
                    const timeAgo = Math.floor((Date.now() - new Date(eq.time)) / 60000);
                    return '<div class="flex items-center gap-4 p-4 bg-slate-700/30 rounded-lg hover:bg-slate-700/50 transition">' +
                        '<div class="' + color + ' w-16 h-16 rounded-lg flex items-center justify-center font-bold text-xl">' +
                        eq.magnitude.toFixed(1) +
                        '</div>' +
                        '<div class="flex-1">' +
                        '<div class="font-semibold">' + eq.place + '</div>' +
                        '<div class="text-sm text-slate-400">' +
                        timeAgo + ' min ago • Depth: ' + eq.depth.toFixed(1) + ' km' +
                        '</div>' +
                        '</div>' +
                        '<a href="' + eq.url + '" target="_blank" ' +
                        'class="px-4 py-2 bg-blue-500 rounded hover:bg-blue-600 transition text-sm">' +
                        'Details' +
                        '</a>' +
                        '</div>';
                }).join('');
        }

        // Initialize
        initMap();

        // Fetch initial data
        fetch('/api/earthquakes')
            .then(r => r.json())
            .then(updateUI);

        // Connect to SSE
        const eventSource = new EventSource('/api/stream');
        eventSource.onmessage = (event) => {
            const data = JSON.parse(event.data);
            updateUI(data);
        };

        eventSource.onerror = () => {
            document.getElementById('status').textContent = 'Reconnecting...';
            setTimeout(() => {
                document.getElementById('status').textContent = 'Live';
            }, 3000);
        };
    </script>
</body>
</html>`
