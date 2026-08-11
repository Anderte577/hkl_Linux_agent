package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type Config struct {
	DeviceName        string `json:"device_name"`
	ServerURL         string `json:"server_url"`
	HeartbeatInterval int    `json:"heartbeat_interval"`
}

type Heartbeat struct {
	DeviceName string `json:"device_name"`
	Hostname   string `json:"hostname"`
	Timestamp  string `json:"timestamp"`
	Status     string `json:"status"`
}

func getHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	return hostname
}

func getIP() string {
	cmd := exec.Command("hostname", "-I")

	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}

	return string(bytes.TrimSpace(output))
}

func sendHeartbeat(config Config) {
	hostname := getHostname()
	ip := getIP()

	data := Heartbeat{
		DeviceName: config.DeviceName,
		Hostname:   hostname,
		Timestamp:  time.Now().Format(time.RFC3339),
		Status:     "ONLINE",
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("Błąd JSON:", err)
		return
	}

	url := config.ServerURL + "/heartbeat"

	req, err := http.NewRequest(
		"POST",
		url,
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		log.Println("Błąd tworzenia requestu:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-HKL-IP", ip)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Println("Nie można połączyć z serwerem:", err)
		return
	}

	defer resp.Body.Close()

	fmt.Println(
		"Heartbeat wysłany:",
		config.DeviceName,
		hostname,
		"IP:",
		ip,
		"HTTP:",
		resp.StatusCode,
	)
}

func loadConfig() Config {
	configPath := "/opt/hkl-agent/config.json"

	data, err := os.ReadFile(configPath)

	if err != nil {
		log.Fatal("Nie można odczytać config.json:", err)
	}

	var config Config

	err = json.Unmarshal(data, &config)

	if err != nil {
		log.Fatal("Błąd konfiguracji:", err)
	}

	if config.DeviceName == "" {
		config.DeviceName = getHostname()
	}

	if config.HeartbeatInterval <= 0 {
		config.HeartbeatInterval = 30
	}

	return config
}

func main() {
	log.Println("HKL-Agent Linux uruchomiony")

	config := loadConfig()

	log.Println("Nazwa urządzenia:", config.DeviceName)
	log.Println("Hostname:", getHostname())
	log.Println("Serwer:", config.ServerURL)

	sendHeartbeat(config)

	ticker := time.NewTicker(
		time.Duration(config.HeartbeatInterval) * time.Second,
	)

	defer ticker.Stop()

	for {
		<-ticker.C
		sendHeartbeat(config)
	}
}
