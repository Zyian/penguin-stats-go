package penguinstats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

type Server string
type DropType string

const (
	ServerUS Server = "US"
	ServerCN Server = "CN"
	ServerJP Server = "JP"
	ServerKR Server = "KR"

	NormalDrop    DropType = "NORMAL_DROP"
	SpecialDrop   DropType = "SPECIAL_DROP"
	ExtraDrop     DropType = "EXTRA_DROP"
	FurnitureDrop DropType = "FURNITURE"

	BaseURL = "https://penguin-stats.io/PenguinStats/api/v2"
)

// PenguinClient is the singleton struct that handles data exchange with Penguin Stats API
type PenguinClient struct {
	client *http.Client
}

// Drop represents the results of an item obtained at the end of a given stage
// At least one drop is required when sending results to the report API
// 		DropType is one of the four static types (see constants)
//		ItemID is a provided item ID obtained from either the data
//		Quantity is how many of the items you obtained for the stage
type Drop struct {
	DropType DropType `json:"dropType,omitempty"`
	ItemID   string   `json:"itemId,omitempty"`
	Quantity int      `json:"quantity,omitempty"`
}

type DropMatrix struct {
	rawData []rawData
}

type rawData struct {
	StageID  string    `json:"stageId,omitempty"`
	ItemID   string    `json:"itemId,omitempty"`
	Quantity int       `json:"quantity,omitempty"`
	Times    int       `json:"times,omitempty"`
	Start    time.Time `json:"start,omitempty"`
	End      time.Time `json:"end,omitempty"`
}

type reportPayload struct {
	Server  string `json:"server,omitempty"`
	StageID string `json:"stageId,omitempty"`
	Drops   []Drop `json:"drops,omitempty"`
	Source  string `json:"source,omitempty"`
	Version string `json:"version,omitempty"`
}

// NewClient creates a new Penguin Stats client with a default timeout of 5 seconds
// This is the same as calling NewClientWithTimeout(5)
func NewClient() *PenguinClient {
	return NewClientWithTimeout(5)
}

// NewClientWithTimeout creates a new Penguin Stats client with a specified timeout
func NewClientWithTimeout(timeout float64) *PenguinClient {
	return &PenguinClient{
		client: &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: time.Duration(timeout) * time.Second,
				}).Dial,
				TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
			},
		},
	}
}

// ReportDrop submits a report of item drops obtained during a stage referenced by stageID
// Setting `source` is **optional** and allows a user to define where the information comes from
// Setting `version` is **optional** and allows a user to define what version of the source the report came from
func (pc *PenguinClient) ReportDrop(ctx context.Context, server Server, stageID string, drops []Drop, source, version string) (string, error) {
	report := reportPayload{
		Server:  string(server),
		StageID: stageID,
		Drops:   drops,
		Source:  source,
		Version: version,
	}

	p, err := json.Marshal(report)
	if err != nil {
		return "", errors.Wrap(err, "unable to marshal data")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", BaseURL+"/report", bytes.NewBuffer(p))
	if err != nil {
		return "", errors.Wrap(err, "could not build request")
	}

	if source != "" {
		req.Header.Set("User-Agent", source)
	}

	resp, err := pc.client.Do(req)
	if err != nil {
		return "", errors.Wrap(err, "unable to complete request")
	}

	if resp.StatusCode != 201 { //TODO: Better error handling, nasty response comes back when not compliant
		return "", fmt.Errorf("PenguinStats returned a bad status code: %d", resp.StatusCode)
	}

	respBody := map[string]string{}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", errors.Wrap(err, "unable to decode response body")
	}

	return respBody["reportHash"], nil
}

// RecallLastReport recalls an item drop report given by `reportHash` when reported within 24 hours of original submission
func (pc *PenguinClient) RecallLastReport(ctx context.Context, reportHash, source string) error {
	hash := struct {
		ReportHash string `json:"reportHash"`
	}{
		ReportHash: reportHash,
	}

	p, err := json.Marshal(hash)
	if err != nil {
		return errors.Wrap(err, "unable to marshal hash body")
	}

	req, err := http.NewRequestWithContext(ctx, "POST", BaseURL+"/report/recall", bytes.NewBuffer(p))
	if err != nil {
		return errors.Wrap(err, "could not build request")
	}

	if source != "" {
		req.Header.Set("User-Agent", source)
	}

	resp, err := pc.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "unable to complete request")
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("PenguinStats returned bad status code: %d", resp.StatusCode)
	}

	return nil
}
