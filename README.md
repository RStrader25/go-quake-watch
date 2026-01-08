# 🌍 Real-time Earthquake Monitor

![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)
![SSE](https://img.shields.io/badge/SSE-Enabled-green?style=for-the-badge)
![Tailwind](https://img.shields.io/badge/Tailwind-CSS-38B2AC?style=for-the-badge&logo=tailwind-css)
![Render](https://img.shields.io/badge/Render-Deployed-46E3B7?style=for-the-badge&logo=render&logoColor=white)

A minimalist real-time earthquake monitoring dashboard powered by USGS API. Track global seismic activity with live updates via Server-Sent Events.

---

## ✨ Features

- 🔴 **Real-time Updates** - SSE streaming for instant earthquake data
- 🗺️ **Interactive Map** - Visualize earthquakes on a world map
- 📊 **Magnitude Filtering** - Filter by earthquake intensity
- 🎨 **Beautiful UI** - Clean Tailwind CSS design
- 🚀 **Lightweight** - Minimal dependencies, maximum performance
- 🌐 **USGS Data** - Official US Geological Survey API

---

## 🚀 Quick Start

Clone the repository:

```bash
git clone https://github.com/smart-developer1791/go-quake-watch
cd go-quake-watch
```

Initialize dependencies and run:

```bash
go mod tidy
go run .
```

Open your browser at `http://localhost:8080`

---

## 📦 Tech Stack

- **Backend:** Go 1.21+ with Chi router
- **Frontend:** Vanilla JS + Tailwind CSS
- **Maps:** Leaflet.js
- **API:** USGS Earthquake API
- **Real-time:** Server-Sent Events (SSE)

---

## 🛠️ Configuration

Environment variables (optional):

```bash
PORT=8080                    # Server port
UPDATE_INTERVAL=30           # Update interval in seconds
```

---

## 📡 API Endpoints

- `GET /` - Main dashboard
- `GET /api/earthquakes` - Fetch latest earthquakes (JSON)
- `GET /api/stream` - SSE stream for real-time updates

---

## 🎯 Use Cases

- Monitor seismic activity in your region
- Educational tool for geology students
- Emergency preparedness dashboard
- Data visualization practice

---

## 📸 Preview

The dashboard displays:
- Latest earthquakes with magnitude, location, and time
- Interactive world map with earthquake markers
- Color-coded severity levels
- Auto-refreshing data every 30 seconds

---

## 🤝 Contributing

Contributions are welcome! Feel free to:
- Report bugs
- Suggest features
- Submit pull requests

---

## 🙏 Credits

- Earthquake data provided by [USGS](https://earthquake.usgs.gov/)
- Maps powered by [Leaflet](https://leafletjs.com/)
- UI styled with [Tailwind CSS](https://tailwindcss.com/)

---

## Deploy in 10 seconds

[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)
