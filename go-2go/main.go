package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	UpURL    string
	PURL   string
	AutoA   bool
	FPath     string
	SPath      string
	Port         string
	UUID         string
	NServer  string
	NPort    string
	NKey     string
	ErgouDomain   string
	ErgouAuth     string
	ErgouPort     int
	CFIP         string
	CFPort       int
	Name         string
}

func loadConfig() *Config {
	return &Config{
		UpURL:    getEnv("UP_URL", ""), 
		PURL:   getEnv("P_URL", ""),
		AutoA:   getEnvAsBool("AUTO_A", false),
		FPath:     getEnv("F_PATH", "./tmp"),      
		SPath:      getEnv("S_PATH", "sub"),         
		Port:         getEnv("SERVER_PORT", getEnv("PORT", "3000")),
		UUID:         getEnv("UUID", "2faaf996-d2b0-440d-8258-81f2b05dd0e4"),
		NServer:  getEnv("N_SERVER", ""),
		NPort:    getEnv("N_PORT", ""),  
		NKey:     getEnv("N_KEY", ""),   
		ErgouDomain:   getEnv("ERGOU_DOMAIN", ""), 
		ErgouAuth:     getEnv("ERGOU_AUTH", ""),   
		ErgouPort:     getEnvAsInt("ERGOU_PORT", 8001),
		CFIP:         getEnv("CFIP", "ip.sb"),
		CFPort:       getEnvAsInt("CFPORT", 443),       
		Name:         getEnv("NAME", "123"),            
	}
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

func deleteNodes(cfg *Config) error {
	if cfg.UpURL == "" {
		return nil
	}

	SPath := filepath.Join(cfg.FPath, "sub.txt")
	if _, err := os.Stat(SPath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(SPath)
	if err != nil {
		return nil
	}

	decoded, err := base64.StdEncoding.DecodeString(string(content))
	if err != nil {
		return nil
	}

	nodes := []string{}
	for _, line := range strings.Split(string(decoded), "\n") {
		if matched, _ := regexp.MatchString(`(vless|vmess|trojan|hysteria2|tuic)://`, line); matched {
			nodes = append(nodes, strings.TrimSpace(line))
		}
	}

	if len(nodes) == 0 {
		return nil
	}

	jsonData := map[string]interface{}{
		"nodes": nodes,
	}
	
	jsonBytes, _ := json.Marshal(jsonData)
	resp, err := http.Post(cfg.UpURL+"/api/delete-nodes", 
		"application/json", 
		bytes.NewBuffer(jsonBytes))
	
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	return nil
}

func uploadNodes(cfg *Config) {
	if cfg.UpURL == "" && cfg.PURL == "" {
		return
	}

	if cfg.UpURL != "" && cfg.PURL != "" {
		
		subscriptionUrl := fmt.Sprintf("%s/%s", cfg.PURL, cfg.SPath)
		jsonData := map[string]interface{}{
			"subscription": []string{subscriptionUrl},
		}
		
		jsonBytes, _ := json.Marshal(jsonData)
		resp, err := http.Post(cfg.UpURL+"/api/add-subscriptions", 
			"application/json", 
			bytes.NewBuffer(jsonBytes))
		
		if err == nil && resp.StatusCode == 200 {
			log.Println("Subscription uploaded successfully")
		}
		if resp != nil {
			resp.Body.Close()
		}
	} else if cfg.UpURL != "" {
		
		SPath := filepath.Join(cfg.FPath, "sub.txt")
		if _, err := os.Stat(SPath); os.IsNotExist(err) {
			return
		}
	
		content, err := os.ReadFile(SPath)
		if err != nil {
			return
		}
				
		decoded, err := base64.StdEncoding.DecodeString(string(content))
		if err != nil {
			return
		}
	
		nodes := []string{}
		for _, line := range strings.Split(string(decoded), "\n") { 
			if matched, _ := regexp.MatchString(`(vless|vmess|trojan|hysteria2|tuic)://`, line); matched {
				nodes = append(nodes, strings.TrimSpace(line))
			}
		}

		if len(nodes) == 0 {
			return
		}

		jsonData := map[string]interface{}{
			"nodes": nodes,
		}
		
		jsonBytes, _ := json.Marshal(jsonData)
		resp, err := http.Post(cfg.UpURL+"/api/add-nodes", 
			"application/json", 
			bytes.NewBuffer(jsonBytes))
		
		if err == nil && resp.StatusCode == 200 {
			log.Println("Nodes uploaded successfully")
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func addVisitTask(cfg *Config) {
	if !cfg.AutoA || cfg.PURL == "" {
		log.Println("Skipping adding automatic access task")
		return
	}

	jsonData := map[string]string{
		"url": cfg.PURL,
	}
	
	jsonBytes, _ := json.Marshal(jsonData)
	resp, err := http.Post("https://gifted-steel-cheek.glitch.me/add-url", 
		"application/json", 
		bytes.NewBuffer(jsonBytes))
	
	if err != nil {
		log.Printf("添加URL失败: %v", err)
		return
	}
	defer resp.Body.Close()

	log.Println("automatic access task added successfully")
}

type XRayConfig struct {
	Log       LogConfig       `json:"log"`
	Inbounds  []Inbound      `json:"inbounds"`
	DNS       DNSConfig      `json:"dns"`
	Outbounds []Outbound     `json:"outbounds"`
	Routing   RoutingConfig  `json:"routing"`
}

type LogConfig struct {
	Access   string `json:"access"`
	Error    string `json:"error"`
	Loglevel string `json:"loglevel"`
}

type Inbound struct {
	Port           int                    `json:"port"`
	Protocol       string                 `json:"protocol"`
	Settings       map[string]interface{} `json:"settings"`
	StreamSettings map[string]interface{} `json:"streamSettings,omitempty"`
	Listen         string                 `json:"listen,omitempty"`
	Sniffing       map[string]interface{} `json:"sniffing,omitempty"`
}

type DNSConfig struct {
	Servers []string `json:"servers"`
}

type Outbound struct {
	Protocol string                 `json:"protocol"`
	Settings map[string]interface{} `json:"settings,omitempty"`
	Tag      string                `json:"tag,omitempty"`
}

type RoutingConfig struct {
	DomainStrategy string        `json:"domainStrategy"`
	Rules          []interface{} `json:"rules"`
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}

func cleanupOldFiles(FPath string) {
	pathsToDelete := []string{"web", "bot", "npm", "sub.txt", "boot.log"}
	for _, file := range pathsToDelete {
		fullPath := filepath.Join(FPath, file)
		os.Remove(fullPath)  
	}
}

func downloadFile(FPath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Download failed: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(FPath)
	if err != nil {
		return fmt.Errorf("Failed to create file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to write file: %v", err)
	}

	return nil
}

func getSystemArchitecture() string {
	arch := runtime.GOARCH
	if arch == "arm" || arch == "arm64" || arch == "aarch64" {
		return "arm"
	}
	return "amd"
}

func getFilesForArchitecture(architecture string) []struct {
	fileName string
	fileUrl  string
} {
	var baseFiles []struct {
		fileName string
		fileUrl  string
	}

	if architecture == "arm" {
		baseFiles = []struct {
			fileName string
			fileUrl  string
		}{
			{"web", "https://arm64.ssss.nyc.mn/web"},
			{"bot", "https://arm64.ssss.nyc.mn/2go"},
		}
	} else {
		baseFiles = []struct {
			fileName string
			fileUrl  string
		}{
			{"web", "https://amd64.ssss.nyc.mn/web"},
			{"bot", "https://amd64.ssss.nyc.mn/2go"},
		}
	}

	cfg := loadConfig()
	if cfg.NServer != "" && cfg.NKey != "" {
		if cfg.NPort != "" {
			npmUrl := "https://amd64.ssss.nyc.mn/agent"
			if architecture == "arm" {
				npmUrl = "https://arm64.ssss.nyc.mn/agent"
			}
			baseFiles = append([]struct {
				fileName string
				fileUrl  string
			}{{"npm", npmUrl}}, baseFiles...)
		} else {
			phpUrl := "https://amd64.ssss.nyc.mn/v1"
			if architecture == "arm" {
				phpUrl = "https://arm64.ssss.nyc.mn/v1"
			}
			baseFiles = append([]struct {
				fileName string
				fileUrl  string
			}{{"php", phpUrl}}, baseFiles...)
		}
	}

	return baseFiles
}

func generateXRayConfig(cfg *Config) {
	config := XRayConfig{
		Log: LogConfig{
			Access:   "/dev/null",
			Error:    "/dev/null",
			Loglevel: "none",
		},
		Inbounds: []Inbound{
			{
				Port:     cfg.ErgouPort,
				Protocol: "vless",
				Settings: map[string]interface{}{
					"clients": []map[string]interface{}{
						{"id": cfg.UUID, "flow": "xtls-rprx-vision"},
					},
					"decryption": "none",
					"fallbacks": []map[string]interface{}{
						{"dest": 3001},
						{"path": "/vless-ERGOU", "dest": 3002},
						{"path": "/vmess-ERGOU", "dest": 3003},
						{"path": "/trojan-ERGOU", "dest": 3004},
					},
				},
				StreamSettings: map[string]interface{}{
					"network": "tcp",
				},
			},
		},
		DNS: DNSConfig{
			Servers: []string{"https+local://8.8.8.8/dns-query"},
		},
		Outbounds: []Outbound{
			{
				Protocol: "freedom",
				Tag:      "direct",
			},
			{
				Protocol: "blackhole",
				Tag:      "block",
			},
		},
	}

	additionalInbounds := []Inbound{
		{
			Port:     3001,
			Listen:   "127.0.0.1",
			Protocol: "vless",
			Settings: map[string]interface{}{
				"clients":     []map[string]interface{}{{"id": cfg.UUID}},
				"decryption": "none",
			},
			StreamSettings: map[string]interface{}{
				"network":  "tcp",
				"security": "none",
			},
		},
		{
			Port:     3002,
			Listen:   "127.0.0.1",
			Protocol: "vless",
			Settings: map[string]interface{}{
				"clients": []map[string]interface{}{
					{"id": cfg.UUID, "level": 0},
				},
				"decryption": "none",
			},
			StreamSettings: map[string]interface{}{
				"network":  "ws",
				"security": "none",
				"wsSettings": map[string]interface{}{
					"path": "/vless-ERGOU",
				},
			},
			Sniffing: map[string]interface{}{
				"enabled":      true,
				"destOverride": []string{"http", "tls", "quic"},
				"metadataOnly": false,
			},
		},
		{
			Port:     3003,
			Listen:   "127.0.0.1",
			Protocol: "vmess",
			Settings: map[string]interface{}{
				"clients": []map[string]interface{}{
					{"id": cfg.UUID, "alterId": 0},
				},
			},
			StreamSettings: map[string]interface{}{
				"network": "ws",
				"wsSettings": map[string]interface{}{
					"path": "/vmess-ERGOU",
				},
			},
			Sniffing: map[string]interface{}{
				"enabled":      true,
				"destOverride": []string{"http", "tls", "quic"},
				"metadataOnly": false,
			},
		},
		{
			Port:     3004,
			Listen:   "127.0.0.1",
			Protocol: "trojan",
			Settings: map[string]interface{}{
				"clients": []map[string]interface{}{
					{"password": cfg.UUID},
				},
			},
			StreamSettings: map[string]interface{}{
				"network":  "ws",
				"security": "none",
				"wsSettings": map[string]interface{}{
					"path": "/trojan-ERGOU",
				},
			},
			Sniffing: map[string]interface{}{
				"enabled":      true,
				"destOverride": []string{"http", "tls", "quic"},
				"metadataOnly": false,
			},
		},
	}
	config.Inbounds = append(config.Inbounds, additionalInbounds...)

	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		log.Printf("Failed to serialize config: %v", err)
		return
	}

	configPath := filepath.Join(cfg.FPath, "config.json")
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		log.Printf("Failed to write config file: %v", err)
		return
	}
}

func startServer(cfg *Config) {
	arch := getSystemArchitecture()
	files := getFilesForArchitecture(arch)

	for _, file := range files {
		fpath := filepath.Join(cfg.FPath, file.fileName)
		if err := downloadFile(fpath, file.fileUrl); err != nil {
			log.Printf("Failed to download %s: %v", file.fileName, err)
