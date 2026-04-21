package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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
	NoImageArg         string
	ImageNotFound      string
	ConfigLoadError    string
	FindingsLoadError  string
	SelectTitle        string
	SelectText         string
	ColID              string
	ColTitle           string
	NewFinding         string
	NoTitle            string
	NewFindingFail     string
	NoTemplateUUID     string
	UploadFail         string
	UpdateFail         string
	Success            string
	UploadCaption      string
	ErrImageRead       string
	ErrFormFile        string
	ErrFormName        string
	ErrConnectUpload   string
	ErrAPIUpload       string
	ErrJSONParse       string
	ErrNoFilename      string
	ErrFindingFetch    string
	ErrFindingPatch    string
	ErrFindingPatchAPI string
	DeveloperCredit    string
	LastUsed           string
}

var translations = map[string]T{
	"pt-br": {
		NoImageArg:         "Nenhum arquivo de imagem passado como argumento pelo KSNIP.",
		ImageNotFound:      "O arquivo de imagem não foi encontrado no caminho:\n%s",
		ConfigLoadError:    "Erro ao carregar as configurações.\nVerifique se ~/.sysreptor/config.yaml existe e contém server e token,\nou configure server e token em ~/.config/reptor-widget/config.conf",
		FindingsLoadError:  "Não foi possível carregar os findings.\nVerifique a conexão e o Token.",
		SelectTitle:        "SysReptor - Selecione o Finding",
		SelectText:         "Onde deseja anexar esta evidência?",
		ColID:              "ID",
		ColTitle:           "Título do Finding",
		NewFinding:         "🌟 [CRIAR NOVO FINDING]",
		NoTitle:            "Sem Título",
		NewFindingFail:     "Falha ao criar um novo Finding.",
		NoTemplateUUID:     "Template UUID não configurado.\nDefina template_uuid em ~/.config/reptor-widget/config.conf",
		UploadFail:         "Falha no upload da imagem:\n\n%v",
		UpdateFail:         "Falha ao atualizar o Finding:\n\n%v",
		Success:            "Evidência anexada com sucesso no SysReptor!",
		UploadCaption:      "\n Upload via KSNIP-API - by Le Fort\n ![%s](/images/name/%s)\n",
		ErrImageRead:       "não foi possível ler a imagem local: %v",
		ErrFormFile:        "erro ao montar form-data (file): %v",
		ErrFormName:        "erro ao montar form-data (name): %v",
		ErrConnectUpload:   "falha de conexão ao enviar imagem: %v",
		ErrAPIUpload:       "API recusou o upload.\nStatus: %d\nResposta: %s",
		ErrJSONParse:       "upload concluído, mas falhou ao ler o JSON de resposta:\n%s",
		ErrNoFilename:      "upload concluído, mas a API não retornou o nome do arquivo.\nResposta: %s",
		ErrFindingFetch:    "erro ao buscar o finding atual para pegar os dados antigos",
		ErrFindingPatch:    "falha de conexão ao atualizar (PATCH)",
		ErrFindingPatchAPI: "API recusou a atualização (PATCH).\nStatus: %d\nResposta: %s",
		DeveloperCredit:    "Desenvolvido por leforense (aka Le Fort)",
		LastUsed:           "★ [ÚLTIMO USADO] ",
	},
	"en-us": {
		NoImageArg:         "No image file passed as argument by KSNIP.",
		ImageNotFound:      "Image file not found at path:\n%s",
		ConfigLoadError:    "Failed to load configuration.\nCheck that ~/.sysreptor/config.yaml exists and contains server and token,\nor set server and token in ~/.config/reptor-widget/config.conf",
		FindingsLoadError:  "Could not load findings.\nCheck connection and Token.",
		SelectTitle:        "SysReptor - Select Finding",
		SelectText:         "Where do you want to attach this evidence?",
		ColID:              "ID",
		ColTitle:           "Finding Title",
		NewFinding:         "🌟 [CREATE NEW FINDING]",
		NoTitle:            "No Title",
		NewFindingFail:     "Failed to create a new Finding.",
		NoTemplateUUID:     "Template UUID not configured.\nSet template_uuid in ~/.config/reptor-widget/config.conf",
		UploadFail:         "Image upload failed:\n\n%v",
		UpdateFail:         "Failed to update Finding:\n\n%v",
		Success:            "Evidence successfully attached in SysReptor!",
		UploadCaption:      "\n Upload via KSNIP-API - by Le Fort\n ![%s](/images/name/%s)\n",
		ErrImageRead:       "could not read local image: %v",
		ErrFormFile:        "error building form-data (file): %v",
		ErrFormName:        "error building form-data (name): %v",
		ErrConnectUpload:   "connection error while uploading image: %v",
		ErrAPIUpload:       "API rejected the upload.\nStatus: %d\nResponse: %s",
		ErrJSONParse:       "upload completed, but failed to parse JSON response:\n%s",
		ErrNoFilename:      "upload completed, but API did not return a filename.\nResponse: %s",
		ErrFindingFetch:    "error fetching current finding data",
		ErrFindingPatch:    "connection error during update (PATCH)",
		ErrFindingPatchAPI: "API rejected the update (PATCH).\nStatus: %d\nResponse: %s",
		DeveloperCredit:    "Developed by leforense (aka Le Fort)",
		LastUsed:           "★ [LAST USED] ",
	},
}

func getT(lang string) T {
	if t, ok := translations[lang]; ok {
		return t
	}
	return translations["en-us"]
}

type Finding struct {
	ID   string                 `json:"id"`
	Data map[string]interface{} `json:"data"`
}

func getSysreptorConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".sysreptor", "config.yaml")
}

func getWidgetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "reptor-widget", "config.conf")
}

type LastFinding struct {
	ProjectID string `json:"project_id"`
	FindingID string `json:"finding_id"`
}

func getLastFindingPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "reptor-widget", "last_finding.json")
}

func loadLastFinding() *LastFinding {
	data, err := os.ReadFile(getLastFindingPath())
	if err != nil {
		return nil
	}
	var lf LastFinding
	if err = json.Unmarshal(data, &lf); err != nil {
		return nil
	}
	return &lf
}

func saveLastFinding(projectID, findingID string) {
	path := getLastFindingPath()
	os.MkdirAll(filepath.Dir(path), 0755)
	data, _ := json.Marshal(LastFinding{ProjectID: projectID, FindingID: findingID})
	os.WriteFile(path, data, 0644)
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

func main() {
	wc := loadWidgetConfig()
	t := getT(wc.Language)

	if len(os.Args) < 2 {
		showError(t.NoImageArg)
		return
	}
	imagePath := os.Args[1]

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		showError(fmt.Sprintf(t.ImageNotFound, imagePath))
		return
	}

	cfg, err := loadConfig(wc)
	if err != nil {
		showError(t.ConfigLoadError)
		return
	}

	findings := getFindings(cfg)
	if findings == nil {
		showError(t.FindingsLoadError)
		return
	}

	selectedFindingID := showMenu(findings, cfg.ProjectID, t)
	if selectedFindingID == "" {
		return
	}

	if selectedFindingID == "NEW_FINDING" {
		if wc.TemplateUUID == "" {
			showError(t.NoTemplateUUID)
			return
		}
		selectedFindingID = createNewFinding(cfg, wc.TemplateUUID)
		if selectedFindingID == "" {
			showError(t.NewFindingFail)
			return
		}
	}

	uploadedFilename, err := uploadImage(cfg, imagePath, t)
	if err != nil {
		showError(fmt.Sprintf(t.UploadFail, err))
		return
	}

	err = updateFindingWithImage(cfg, selectedFindingID, uploadedFilename, wc.APIUploadField, t)
	if err != nil {
		showError(fmt.Sprintf(t.UpdateFail, err))
		return
	}

	saveLastFinding(cfg.ProjectID, selectedFindingID)
	exec.Command("zenity", "--info", "--text="+t.Success, "--timeout=3").Run()
}

func getFindings(cfg *Config) []Finding {
	url := fmt.Sprintf("%s/api/v1/pentestprojects/%s/findings/", cfg.Server, cfg.ProjectID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Authorization", "Bearer "+cfg.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result []Finding
	json.Unmarshal(body, &result)

	return result
}

func showMenu(findings []Finding, projectID string, t T) string {
	textMsg := fmt.Sprintf("%s\n\n<span size='small' color='#888888'><i>%s</i></span>",
		t.SelectText, t.DeveloperCredit)

	args := []string{
		"--list",
		"--title=" + t.SelectTitle,
		"--text=" + textMsg,
		"--column=" + t.ColID,
		"--column=" + t.ColTitle,
		"--hide-column=1",
		"--print-column=1",
		"--width=600",
		"--height=400",
	}

	// Determine last-used finding for this project (different project = ignore)
	lastFindingID := ""
	if lf := loadLastFinding(); lf != nil && lf.ProjectID == projectID {
		lastFindingID = lf.FindingID
	}

	// Last-used finding goes to the top with a visual marker
	for _, f := range findings {
		if f.ID == lastFindingID {
			title := t.NoTitle
			if val, ok := f.Data["title"].(string); ok && val != "" {
				title = val
			}
			args = append(args, f.ID, t.LastUsed+title)
		}
	}

	// All other findings
	for _, f := range findings {
		if f.ID == lastFindingID {
			continue
		}
		title := t.NoTitle
		if val, ok := f.Data["title"].(string); ok && val != "" {
			title = val
		}
		args = append(args, f.ID, title)
	}

	// New finding always at the bottom
	args = append(args, "NEW_FINDING", t.NewFinding)

	cmd := exec.Command("zenity", args...)
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return ""
	}

	selectedID := strings.TrimSpace(out.String())
	return strings.Split(selectedID, "|")[0]
}

func createNewFinding(cfg *Config, templateUUID string) string {
	url := fmt.Sprintf("%s/api/v1/pentestprojects/%s/findings/fromtemplate/", cfg.Server, cfg.ProjectID)
	payload := []byte(fmt.Sprintf(`{"template": "%s"}`, templateUUID))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Add("Authorization", "Bearer "+cfg.Token)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil || (resp.StatusCode != 200 && resp.StatusCode != 201) {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var newFinding Finding
	json.Unmarshal(body, &newFinding)

	return newFinding.ID
}

func uploadImage(cfg *Config, imagePath string, t T) (string, error) {
	url := fmt.Sprintf("%s/api/v1/pentestprojects/%s/upload/", cfg.Server, cfg.ProjectID)

	file, err := os.Open(imagePath)
	if err != nil {
		return "", fmt.Errorf(t.ErrImageRead, err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileName := filepath.Base(imagePath)
	part, err := writer.CreateFormFile("file", fileName)
	if err != nil {
		return "", fmt.Errorf(t.ErrFormFile, err)
	}
	io.Copy(part, file)

	if err = writer.WriteField("name", fileName); err != nil {
		return "", fmt.Errorf(t.ErrFormName, err)
	}
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Add("Authorization", "Bearer "+cfg.Token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf(t.ErrConnectUpload, err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf(t.ErrAPIUpload, resp.StatusCode, string(respBody))
	}

	var uploadResp struct {
		Name string `json:"name"`
	}
	if err = json.Unmarshal(respBody, &uploadResp); err != nil {
		return "", fmt.Errorf(t.ErrJSONParse, string(respBody))
	}

	if uploadResp.Name == "" {
		return "", fmt.Errorf(t.ErrNoFilename, string(respBody))
	}

	return uploadResp.Name, nil
}

func updateFindingWithImage(cfg *Config, findingID, uploadedFilename, apiUploadField string, t T) error {
	url := fmt.Sprintf("%s/api/v1/pentestprojects/%s/findings/%s/", cfg.Server, cfg.ProjectID, findingID)

	reqGet, _ := http.NewRequest("GET", url, nil)
	reqGet.Header.Add("Authorization", "Bearer "+cfg.Token)
	client := &http.Client{Timeout: 10 * time.Second}
	respGet, err := client.Do(reqGet)
	if err != nil || respGet.StatusCode != 200 {
		return fmt.Errorf("%s", t.ErrFindingFetch)
	}
	defer respGet.Body.Close()

	bodyGet, _ := io.ReadAll(respGet.Body)
	var currentFinding Finding
	json.Unmarshal(bodyGet, &currentFinding)

	apiUpload := ""
	if val, ok := currentFinding.Data[apiUploadField].(string); ok {
		apiUpload = val
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	newImageMarkdown := fmt.Sprintf(t.UploadCaption, timestamp, uploadedFilename)

	if apiUpload != "" && !strings.HasSuffix(apiUpload, "\n") {
		apiUpload += "\n"
	}
	apiUpload += newImageMarkdown

	currentFinding.Data[apiUploadField] = apiUpload

	patchPayload := map[string]interface{}{
		"data": currentFinding.Data,
	}
	patchBytes, _ := json.Marshal(patchPayload)

	reqPatch, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(patchBytes))
	reqPatch.Header.Add("Authorization", "Bearer "+cfg.Token)
	reqPatch.Header.Add("Content-Type", "application/json")

	respPatch, err := client.Do(reqPatch)
	if err != nil {
		return fmt.Errorf("%s", t.ErrFindingPatch)
	}
	defer respPatch.Body.Close()

	if respPatch.StatusCode != 200 {
		respBody, _ := io.ReadAll(respPatch.Body)
		return fmt.Errorf(t.ErrFindingPatchAPI, respPatch.StatusCode, string(respBody))
	}

	return nil
}

func showError(msg string) {
	// Width 500 fits longer API error responses without wrapping awkwardly
	exec.Command("zenity", "--error", "--width=500", "--text="+msg).Run()
}
