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

// 环境变量配置
type Config struct {
	UploadURL    string
	ProjectURL   string
	AutoAccess   bool
	FilePath     string
	SubPath      string
	Port         string
	UUID         string
	NzServer  string
	NzPort    string
	NzKey     string
	ErgouDomain   string
	ErgouAuth     string
	ErgouPort     int
	CFIP         string
	CFPort       int
	Name         string
}

func loadConfig() *Config {
	return &Config{
		UploadURL:    getEnv("UPLOAD_URL", ""),
		ProjectURL:   getEnv("PROJECT_URL", ""),
		AutoAccess:   getEnvAsBool("AUTO_ACCESS", false),
		FilePath:     getEnv("FILE_PATH", "./tmp"),
		SubPath:      getEnv("SUB_PATH", "sub"),
		Port:         getEnv("SERVER_PORT", getEnv("PORT", "3000")),
		UUID:         getEnv("UUID", "f30a03c9-cf1c-48ce-9f6d-f4b6cd2382d9"),
		NzServer:  getEnv("N_SERVER", ""),
		NzPort:    getEnv("N_PORT", ""),
		NzKey:     getEnv("N_KEY", ""),
		ErgouDomain:   getEnv("ERGOU_DOMAIN", ""),
		ErgouAuth:     getEnv("ERGOU_AUTH", ""),
		ErgouPort:     getEnvAsInt("ERGOU_PORT", 8001),
		CFIP:         getEnv("CFIP", "www.visa.com.tw"),
		CFPort:       getEnvAsInt("CFPORT", 443),
		Name:         getEnv("NAME", "Vls"),
	}
}


// 创建运行目录
func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
}

// 删除历史
func deleteNodes(cfg *Config) error {
	if cfg.UploadURL == "" {
		return nil
	}

	subPath := filepath.Join(cfg.FilePath, "sub.txt")
	if _, err := os.Stat(subPath); os.IsNotExist(err) {
		return nil
	}

	content, err := os.ReadFile(subPath)
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
	resp, err := http.Post(cfg.UploadURL+"/api/delete-nodes", 
		"application/json", 
		bytes.NewBuffer(jsonBytes))
	
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	
	return nil
}

// 上传或订阅
func uploadNodes(cfg *Config) {
	if cfg.UploadURL == "" && cfg.ProjectURL == "" {
		return
	}

	if cfg.UploadURL != "" && cfg.ProjectURL != "" {
		// 上传
		subscriptionUrl := fmt.Sprintf("%s/%s", cfg.ProjectURL, cfg.SubPath)
		jsonData := map[string]interface{}{
			"subscription": []string{subscriptionUrl},
		}
		
		jsonBytes, _ := json.Marshal(jsonData)
		resp, err := http.Post(cfg.UploadURL+"/api/add-subscriptions", 
			"application/json", 
			bytes.NewBuffer(jsonBytes))
		
		if err == nil && resp.StatusCode == 200 {
			log.Println("Subscription uploaded successfully")
		}
		if resp != nil {
			resp.Body.Close()
		}
	} else if cfg.UploadURL != "" {
		// 上传
		subPath := filepath.Join(cfg.FilePath, "sub.txt")
		if _, err := os.Stat(subPath); os.IsNotExist(err) {
			return
		}
	
		content, err := os.ReadFile(subPath)
		if err != nil {
			return
		}
				
		decoded, err := base64.StdEncoding.DecodeString(string(content))
		if err != nil {
			return
		}
	
		nodes := []string{}
		for _, line := range strings.Split(string(decoded), "\n") {  // Changed from content to decoded
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
		resp, err := http.Post(cfg.UploadURL+"/api/add-nodes", 
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

// 添加自动访问任务
func addVisitTask(cfg *Config) {
	if !cfg.AutoAccess || cfg.ProjectURL == "" {
		log.Println("Skipping adding automatic access task")
		return
	}

	jsonData := map[string]string{
		"url": cfg.ProjectURL,
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

// Xserver配置结构
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

func cleanupOldFiles(filePath string) {
	pathsToDelete := []string{"web", "bot", "npm", "sub.txt", "boot.log"}
	for _, file := range pathsToDelete {
		fullPath := filepath.Join(filePath, file)
		os.Remove(fullPath)  
	}
}

func downloadFile(filePath string, url string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Download failed: %v", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(filePath)
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
	if cfg.NzServer != "" && cfg.NzKey != "" {
		if cfg.NzPort != "" {
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
						{"path": "/vless-argo", "dest": 3002},
						{"path": "/vmess-argo", "dest": 3003},
						{"path": "/trojan-argo", "dest": 3004},
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

	// 添加其他inbounds
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
					"path": "/vless-argo",
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
					"path": "/vmess-argo",
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
					"path": "/trojan-argo",
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

	configPath := filepath.Join(cfg.FilePath, "config.json")
	if err := os.WriteFile(configPath, configBytes, 0644); err != nil {
		log.Printf("Failed to write config file: %v", err)
		return
	}
}

func startServer(cfg *Config) {
	// 下载并运行依赖文件
	arch := getSystemArchitecture()
	files := getFilesForArchitecture(arch)

	// 下载所有文件
	for _, file := range files {
		filePath := filepath.Join(cfg.FilePath, file.fileName)
		if err := downloadFile(filePath, file.fileUrl); err != nil {
			log.Printf("Failed to download %s: %v", file.fileName, err)
			continue
		}
		log.Printf("Successfully downloaded %s", file.fileName)

		if err := os.Chmod(filePath, 0755); err != nil {
			log.Printf("Failed to set permissions for %s: %v", filePath, err)
		}
	}

	// 运行nServer
	if cfg.NzServer != "" && cfg.NzKey != "" {
		if cfg.NzPort == "" {
			// 生成 config.yaml
			configYaml := fmt.Sprintf(`
client_secret: %s
debug: false
disable_auto_update: true
disable_command_execute: false
disable_force_update: true
disable_nat: false
disable_send_query: false
gpu: false
insecure_tls: false
ip_report_period: 1800
report_delay: 1
server: %s
skip_connection_count: false
skip_procs_count: false
temperature: false
tls: false
use_gitee_to_upgrade: false
use_ipv6_country_code: false
uuid: %s`, cfg.NzKey, cfg.NzServer, cfg.UUID)

			if err := os.WriteFile(filepath.Join(cfg.FilePath, "config.yaml"), []byte(configYaml), 0644); err != nil {
				log.Printf("Failed to write config.yaml: %v", err)
			}

			cmd := exec.Command(filepath.Join(cfg.FilePath, "php"), "-c", filepath.Join(cfg.FilePath, "config.yaml"))
			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start php: %v", err)
			} else {
				log.Println("php is running")
			}
		} else {
			nzArgs := []string{"-s", fmt.Sprintf("%s:%s", cfg.NzServer, cfg.NzPort), "-p", cfg.NzKey}
			
			// 检查是否需要TLS
			tlsPorts := []string{"443", "8443", "2096", "2087", "2083", "2053"}
			for _, port := range tlsPorts {
				if cfg.NzPort == port {
					nzArgs = append(nzArgs, "--tls")
					break
				}
			}

			cmd := exec.Command(filepath.Join(cfg.FilePath, "npm"), nzArgs...)
			if err := cmd.Start(); err != nil {
				log.Printf("Failed to start npm: %v", err)
			} else {
				log.Println("npm is running")
			}
		}
	} else {
		log.Println("NEZHA variable is empty, skipping running")
	}

	// 运行
	cmd := exec.Command(filepath.Join(cfg.FilePath, "web"), "-c", filepath.Join(cfg.FilePath, "config.json"))
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err == nil {
		cmd.Stdout = devNull
		cmd.Stderr = devNull
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start Web: %v", err)
	} else {
		log.Println("web is running")
	}

	// 运行cfserver
	if _, err := os.Stat(filepath.Join(cfg.FilePath, "bot")); err == nil {
		var args []string

		if matched, _ := regexp.MatchString(`^[A-Z0-9a-z=]{120,250}$`, cfg.ErgouAuth); matched {
			args = []string{"tunnel", "--edge-ip-version", "auto", "--no-autoupdate", "--protocol", "http2", "run", "--token", cfg.ErgouAuth}
		} else if strings.Contains(cfg.ErgouAuth, "TunnelSecret") {
			args = []string{"tunnel", "--edge-ip-version", "auto", "--config", filepath.Join(cfg.FilePath, "tunnel.yml"), "run"}
		} else {
			args = []string{"tunnel", "--edge-ip-version", "auto", "--no-autoupdate", "--protocol", "http2", 
				"--logfile", filepath.Join(cfg.FilePath, "boot.log"), "--loglevel", "info",
				"--url", fmt.Sprintf("http://localhost:%d", cfg.ErgouPort)}
		}

		cmd := exec.Command(filepath.Join(cfg.FilePath, "bot"), args...)
		// 重定向输出到boot.log
		logFile, err := os.Create(filepath.Join(cfg.FilePath, "boot.log"))
		if err == nil {
			cmd.Stdout = logFile
			cmd.Stderr = logFile
		}

		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start bot: %v", err)
		} else {
			log.Println("bot is running")
		}
	}
}

func generateErgouConfig(cfg *Config) {
	if cfg.ErgouAuth == "" || cfg.ErgouDomain == "" {
		log.Println("ERGOU_DOMAIN or ERGOU_AUTH is empty, using quick tunnels")
		return
	}

	if strings.Contains(cfg.ErgouAuth, "TunnelSecret") {
		if err := os.WriteFile(filepath.Join(cfg.FilePath, "tunnel.json"), []byte(cfg.ErgouAuth), 0644); err != nil {
			log.Printf("Failed to write tunnel.json: %v", err)
			return
		}

		var tunnelData map[string]interface{}
		if err := json.Unmarshal([]byte(cfg.ErgouAuth), &tunnelData); err != nil {
			log.Printf("Failed to parse tunnel data: %v", err)
			return
		}
		tunnelID, ok := tunnelData["TunnelID"].(string)
		if !ok {
			log.Println("Failed to get TunnelID")
			return
		}

		tunnelYaml := fmt.Sprintf(`
tunnel: %s
credentials-file: %s
protocol: http2

ingress:
  - hostname: %s
    service: http://localhost:%d
    originRequest:
      noTLSVerify: true
  - service: http_status:404
`, tunnelID, filepath.Join(cfg.FilePath, "tunnel.json"), cfg.ErgouDomain, cfg.ErgouPort)

		if err := os.WriteFile(filepath.Join(cfg.FilePath, "tunnel.yml"), []byte(tunnelYaml), 0644); err != nil {
			log.Printf("Failed to write tunnel.yml: %v", err)
		}
	} else {
		log.Println("ERGOU_AUTH doesn't match TunnelSecret format, using token connection")
	}
}

// 提取Ergou域名
func extractDomains(cfg *Config) (string, error) {
	if cfg.ErgouAuth != "" && cfg.ErgouDomain != "" {
		log.Printf("ERGOU_DOMAIN: %s", cfg.ErgouDomain)
		return cfg.ErgouDomain, nil
	}

	// 等待boot.log生成并读取域名
	bootLogPath := filepath.Join(cfg.FilePath, "boot.log")
	for i := 0; i < 30; i++ { // 最多等待30秒
		content, err := os.ReadFile(bootLogPath)
		if err == nil {
			re := regexp.MustCompile(`https?://([^/]*trycloudflare\.com)/?`)
			matches := re.FindStringSubmatch(string(content))
			if len(matches) > 1 {
				domain := matches[1]
				log.Printf("ErgouDomain: %s", domain)
				return domain, nil
			}
		}
		time.Sleep(time.Second)
	}

	return "", fmt.Errorf("Failed to get ErgouDomain after 30 seconds")
}

// 生成订阅链接
func generateLinks(cfg *Config, argoDomain string) error {
	cmd := exec.Command("curl", "-s", "https://speed.cloudflare.com/meta")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("Failed to get ISP info: %v", err)
	}

	var meta map[string]interface{}
	if err := json.Unmarshal(output, &meta); err != nil {
		return fmt.Errorf("Failed to parse ISP info: %v", err)
	}

	isp := fmt.Sprintf("%s-%s", meta["country"], meta["asOrganization"])
	isp = strings.ReplaceAll(isp, " ", "_")

	// 生成VMESS配置
	vmess := map[string]interface{}{
		"v":    "2",
		"ps":   fmt.Sprintf("%s-%s", cfg.Name, isp),
		"add":  cfg.CFIP,
		"port": cfg.CFPort,
		"id":   cfg.UUID,
		"aid":  "0",
		"scy":  "none",
		"net":  "ws",
		"type": "none",
		"host": argoDomain,
		"path": "/vmess-argo?ed=2048",
		"tls":  "tls",
		"sni":  argoDomain,
		"alpn": "",
	}

	vmessBytes, err := json.Marshal(vmess)
	if err != nil {
		return fmt.Errorf("Failed to serialize VMESS config: %v", err)
	}

	// 生成订阅内容
	subContent := fmt.Sprintf(`
vless://%s@%s:%d?encryption=none&security=tls&sni=%s&type=ws&host=%s&path=%%2Fvless-argo%%3Fed%%3D2048#%s-%s

vmess://%s

trojan://%s@%s:%d?security=tls&sni=%s&type=ws&host=%s&path=%%2Ftrojan-argo%%3Fed%%3D2048#%s-%s
`,
		cfg.UUID, cfg.CFIP, cfg.CFPort, argoDomain, argoDomain, cfg.Name, isp,
		base64.StdEncoding.EncodeToString(vmessBytes),
		cfg.UUID, cfg.CFIP, cfg.CFPort, argoDomain, argoDomain, cfg.Name, isp,
	)

	// 保存到文件
	subPath := filepath.Join(cfg.FilePath, "sub.txt")
	encodedContent := base64.StdEncoding.EncodeToString([]byte(subContent))
	if err := os.WriteFile(subPath, []byte(encodedContent), 0644); err != nil {
		return fmt.Errorf("Failed to save sub.txt: %v", err)
	}
	fmt.Printf("\n%s\n\n", encodedContent)
	log.Printf("%s/sub.txt saved successfully\n", cfg.FilePath)
	uploadNodes(cfg)  // 上传节点或订阅
	// 添加/sub路由
	http.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, encodedContent)
	})

	return nil
}

// 清理临时文件
func cleanupTempFiles(cfg *Config) {
	time.Sleep(15 * time.Second)
	filesToDelete := []string{
		filepath.Join(cfg.FilePath, "boot.log"),
		filepath.Join(cfg.FilePath, "config.json"),
		filepath.Join(cfg.FilePath, "list.txt"),
		filepath.Join(cfg.FilePath, "npm"),
		filepath.Join(cfg.FilePath, "web"),
		filepath.Join(cfg.FilePath, "bot"),
		filepath.Join(cfg.FilePath, "php"),
	}

	for _, file := range filesToDelete {
		os.Remove(file) 
	}
	fmt.Print("\033[H\033[2J") // Clear screen
	log.Println("App is running")
	log.Println("Thank you for using this script, enjoy!")
}


// 启动所有服务
func startServices(cfg *Config) error {
	generateErgouConfig(cfg)
	startServer(cfg)

	// 提取域名并生成链接
	argoDomain, err := extractDomains(cfg)
	if err != nil {
		return fmt.Errorf("Failed to extract domain: %v", err)
	}

	if err := generateLinks(cfg, argoDomain); err != nil {
		return fmt.Errorf("Failed to generate links: %v", err)
	}

	// 清理临时文件
	go cleanupTempFiles(cfg)

	return nil
}

func main() {
	cfg := loadConfig()
	
	// 创建运行文件夹
	if err := os.MkdirAll(cfg.FilePath, 0775); err != nil {
		log.Printf("Failed to create directory: %v", err)
	}

	// 删除历史节点
	deleteNodes(cfg)

	// 清理历史文件
	cleanupOldFiles(cfg.FilePath)

	// 创建HTTP服务器
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello world!")
	})

	// 生成配置文件
	generateXRayConfig(cfg)

	// 启动核心服务
	if err := startServices(cfg); err != nil {
		log.Printf("Failed to start services: %v", err)
	}

	// 添加自动访问任务
	addVisitTask(cfg)

	log.Printf("http server is running on port: %s\n", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatal(err)
	}
}
