package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	ProjectID string
	Server    string
	Token     string
}

type WidgetConfig struct {
	Language       string
	TemplateUUID   string
	APIUploadField string
	Server         string
	Token          string
}

type T struct {
	ConfigError        string
	NotConnected       string
	UnknownProject     string
	CouldNotLoad       string
	SelectProjectTitle string
	SelectProjectText  string
	DeveloperBy        string
}

var translations = map[string]T{
	"pt-br": {
		ConfigError:        "Erro de Configuração",
		NotConnected:       "Reptor Desconectado",
		UnknownProject:     "Projeto Desconhecido",
		CouldNotLoad:       "Não foi possível carregar os projetos.",
		SelectProjectTitle: "SysReptor - Trocar Projeto",
		SelectProjectText:  "Selecione o Projeto Ativo:",
		DeveloperBy:        "Desenvolvido por",
	},
	"en-us": {
		ConfigError:        "Config Error",
		NotConnected:       "Reptor Not Connected",
		UnknownProject:     "Unknown Project",
		CouldNotLoad:       "Could not load projects.",
		SelectProjectTitle: "SysReptor - Switch Project",
		SelectProjectText:  "Select Active Project:",
		DeveloperBy:        "Developed by",
	},
}

func getT(lang string) T {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations["en-us"]
}

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ProjectsResponse struct {
	Results []Project `json:"results"`
}

func getSysreptorConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".sysreptor", "config.yaml")
}

func getWidgetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "reptor-widget", "config.conf")
}

func parseField(content, key string) string {
	re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(key) + `:\s*(.+)$`)
	if m := re.FindStringSubmatch(content); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func loadWidgetConfig() *WidgetConfig {
	wc := &WidgetConfig{
		Language:       "en-us",
		APIUploadField: "api_upload",
	}

	data, err := os.ReadFile(getWidgetConfigPath())
	if err != nil {
		return wc
	}

	content := string(data)
	if v := parseField(content, "language"); v != "" {
		wc.Language = strings.ToLower(v)
	}
	if v := parseField(content, "template_uuid"); v != "" {
		wc.TemplateUUID = v
	}
	if v := parseField(content, "api_upload_field"); v != "" {
		wc.APIUploadField = v
	}
	if v := parseField(content, "server"); v != "" {
		wc.Server = v
	}
	if v := parseField(content, "token"); v != "" {
		wc.Token = v
	}

	return wc
}

func loadConfig(wc *WidgetConfig) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(getSysreptorConfigPath())
	if err == nil {
		content := string(data)
		if v := parseField(content, "project_id"); v != "" {
			cfg.ProjectID = v
		}
		if v := parseField(content, "server"); v != "" {
			cfg.Server = v
		}
		if v := parseField(content, "token"); v != "" {
			cfg.Token = v
		}
	}

	// Fall back to widget config for server/token if not found in sysreptor config.yaml
	if cfg.Server == "" && wc.Server != "" {
		cfg.Server = wc.Server
	}
	if cfg.Token == "" && wc.Token != "" {
		cfg.Token = wc.Token
	}

	if cfg.Server == "" || cfg.Token == "" {
		return nil, fmt.Errorf("server or token not configured")
	}

	return cfg, nil
}

func updateConfigProjectID(newID string) error {
	path := getSysreptorConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`(?m)^project_id:.*$`)
	newData := re.ReplaceAllString(string(data), fmt.Sprintf("project_id: %s", newID))

	return os.WriteFile(path, []byte(newData), 0644)
}

func getProjectName(cfg *Config) string {
	url := fmt.Sprintf("%s/api/v1/pentestprojects/%s", cfg.Server, cfg.ProjectID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+cfg.Token)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var proj Project
	json.Unmarshal(body, &proj)

	return proj.Name
}

func getAllProjects(cfg *Config) []Project {
	url := fmt.Sprintf("%s/api/v1/pentestprojects/", cfg.Server)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+cfg.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result ProjectsResponse
	json.Unmarshal(body, &result)

	return result.Results
}

// forcePanelRefresh dynamically discovers the GenMon plugin ID and triggers a refresh,
// so the panel updates immediately after switching projects without waiting for the next tick.
func forcePanelRefresh() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	pattern := filepath.Join(home, ".config", "xfce4", "panel", "genmon-*.rc")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return
	}

	// Match by executable name so the community can rename the binary freely.
	execName := filepath.Base(os.Args[0])

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err == nil && strings.Contains(string(content), execName) {
			filename := filepath.Base(file)
			pluginID := strings.TrimSuffix(filename, filepath.Ext(filename))
			eventCmd := fmt.Sprintf("%s:refresh:bool:true", pluginID)
			exec.Command("xfce4-panel", "--plugin-event="+eventCmd).Run()
		}
	}
}

func main() {
	wc := loadWidgetConfig()
	t := getT(wc.Language)

	cfg, err := loadConfig(wc)
	if err != nil {
		fmt.Printf("<txt><span weight='Bold' fgcolor='Red'>%s</span></txt>\n", t.ConfigError)
		return
	}

	if len(os.Args) > 1 && os.Args[1] == "--menu" {
		projects := getAllProjects(cfg)
		if len(projects) == 0 {
			exec.Command("zenity", "--error", "--text="+t.CouldNotLoad).Run()
			return
		}

		developerName := "leforense - aka Le Fort"
		menuText := fmt.Sprintf("%s\n\n<span size='small' color='#888888'><i>%s %s</i></span>",
			t.SelectProjectText, t.DeveloperBy, developerName)

		args := []string{
			"--list",
			"--title=" + t.SelectProjectTitle,
			"--text=" + menuText,
			"--column=ID",
			"--column=Name",
			"--hide-column=1",
			"--print-column=1",
			"--width=600",
			"--height=400",
		}

		for _, p := range projects {
			args = append(args, p.ID, p.Name)
		}

		cmd := exec.Command("zenity", args...)
		var out bytes.Buffer
		cmd.Stdout = &out

		if err := cmd.Run(); err == nil {
			selectedID := strings.TrimSpace(out.String())
			selectedID = strings.Split(selectedID, "|")[0]
			if selectedID != "" {
				updateConfigProjectID(selectedID)
				forcePanelRefresh()
			}
		}
		return
	}

	name := getProjectName(cfg)
	if name == "" {
		fmt.Printf("<txt><span weight='Bold' fgcolor='Red'>%s</span></txt>\n", t.NotConnected)
	} else {
		execName := filepath.Base(os.Args[0])
		fmt.Printf("<txt><span weight='Bold' fgcolor='Green'>%s</span></txt><txtclick>%s --menu</txtclick>\n", name, execName)
	}
}
