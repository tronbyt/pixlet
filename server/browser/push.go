package browser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const (
	TidbytAPIPush = "https://api.tidbyt.com/v0/devices/%s/push"
)

type TidbytPushJSON struct {
	DeviceID       string `json:"deviceID"`
	Image          string `json:"image"`
	InstallationID string `json:"installationID"`
	Background     bool   `json:"background"`
}

func (b *Browser) pushHandler(w http.ResponseWriter, r *http.Request) {
	var (
		deviceID       string
		apiToken       string
		installationID string
		background     bool
	)

	var result map[string]any
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, err)
		return
	}

	_ = json.Unmarshal(bodyBytes, &result)

	config := make(map[string]any)
	for k, val := range result {
		switch k {
		case "deviceID":
			deviceID = val.(string)
		case "apiToken":
			apiToken = val.(string)
		case "installationID":
			installationID = val.(string)
		case "background":
			background = val.(string) == "true"
		default:
			config[k] = val.(string)
		}
	}

	img, err := b.loader.LoadApplet(config)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, err)
		return
	}

	payload, err := json.Marshal(
		TidbytPushJSON{
			DeviceID:       deviceID,
			Image:          img,
			InstallationID: installationID,
			Background:     background,
		},
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, err)
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf(TidbytAPIPush, deviceID),
		bytes.NewReader(payload),
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, err)
		return
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiToken))

	resp, err := client.Do(req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintln(w, err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("Tidbyt API returned an error", "status", resp.Status)
		w.WriteHeader(resp.StatusCode)
		_, _ = fmt.Fprintln(w, err)

		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{}"))
}
