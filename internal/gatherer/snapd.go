package gatherer

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"
)

type snapInfo struct {
	Name        string `json:"name,omitempty"`
	Channel     string `json:"channel"`
	Developer   string `json:"developer"`
	Id          string `json:"id"`
	InstallDate string `json:"install-date"`
	Revision    string `json:"revision"`
	Status      string `json:"status"`
	Summary     string `json:"summary"`
	Version     string `json:"version"`
}

type SnapsResponse struct {
	Result []snapInfo `json:"result"`
}

type ConfResponse struct {
	Result struct {
		Experimental map[string]bool `json:"experimental"`
	} `json:"result"`
}

func createUnixSocketClient(socketPath string) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 10 * time.Second,
	}
}

func getSnapData() (map[string]snapInfo, error) {
	client := createUnixSocketClient("/run/snapd.socket")

	resp, err := client.Get("http://localhost/v2/snaps")
	if err != nil {
		return nil, fmt.Errorf("cannot get snaps: %w", err)
	}
	defer resp.Body.Close()

	var snapsResp SnapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&snapsResp); err != nil {
		return nil, fmt.Errorf("cannot decode snaps response: %w", err)
	}

	snapData := make(map[string]snapInfo)
	for _, snap := range snapsResp.Result {
		snapData[snap.Name] = snapInfo{
			Channel:     snap.Channel,
			Developer:   snap.Developer,
			Id:          snap.Id,
			InstallDate: snap.InstallDate,
			Revision:    snap.Revision,
			Status:      snap.Status,
			Summary:     snap.Summary,
			Version:     snap.Version,
		}
	}

	return snapData, nil
}

func getExperimentalFlagData() (map[string]bool, error) {
	client := createUnixSocketClient("/run/snapd.socket")

	resp, err := client.Get("http://localhost/v2/snaps/system/conf?keys=experimental")
	if err != nil {
		return nil, fmt.Errorf("cannot get config: %w", err)
	}
	defer resp.Body.Close()

	var confResp ConfResponse
	if err := json.NewDecoder(resp.Body).Decode(&confResp); err != nil {
		return nil, fmt.Errorf("cannot decode config response: %w", err)
	}

	return confResp.Result.Experimental, nil
}
