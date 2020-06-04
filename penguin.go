package penguinstats

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

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
// 	Setting source is optional and allows a user to define where the information comes from
// 	Setting version is optional and allows a user to define what version of the source the report came from
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
	defer resp.Body.Close()

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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("PenguinStats returned bad status code: %d", resp.StatusCode)
	}

	return nil
}

// GetMatrixData requests the item drop rate matrix for a given Arknights Server, if no server is provided, default is CN (as stated in PenguinStats docs)
func (pc *PenguinClient) GetMatrixData(ctx context.Context, server ...Server) (*DropMatrix, error) {
	if len(server) > 1 {
		return nil, fmt.Errorf("Bad parameter: server parameter must be 0 or 1")
	}

	s := ServerCN
	if len(server) > 0 {
		s = server[0]
	}

	return pc.requestMatrixData(ctx, s, false, false, "")
}

// GetMatrixDataCustomOptions allows specifying additional options, if isPersonal is set to true userID must be set
func (pc *PenguinClient) GetMatrixDataCustomOptions(ctx context.Context, server Server, showClosedZones, isPersonal bool, userID string) (*DropMatrix, error) {
	if isPersonal && userID == "" {
		return nil, fmt.Errorf("personal stats specified but no user ID provided")
	}
	return pc.requestMatrixData(ctx, server, showClosedZones, isPersonal, userID)
}

func (pc *PenguinClient) requestMatrixData(ctx context.Context, server Server, showClosedZones, isPersonal bool, userID string) (*DropMatrix, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", BaseURL+"/result/matrix", nil)
	if err != nil {
		return nil, err
	}

	if userID != "" {
		req.AddCookie(&http.Cookie{
			Name:  "userID",
			Value: userID,
		})
	}

	req.URL.Query().Add("server", string(server))
	req.URL.Query().Add("show_closed_zones", strconv.FormatBool(showClosedZones))
	req.URL.Query().Add("is_personal", strconv.FormatBool(isPersonal))

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "could not get matrix data")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("recieved bad status code and could not read body: %v", err)
		}
		return nil, fmt.Errorf("recieved bad status code: %d %v", resp.StatusCode, string(b))
	}

	var data []StageDrop
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("could not decode body: %v", err)
	}

	return &DropMatrix{
		rawData:   data,
		stageMap:  make(map[string][]StageDrop),
		processed: false,
	}, nil
}

func (pc *PenguinClient) GetAllStages(ctx context.Context, server Server) ([]Stage, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", BaseURL+"/stages", nil)
	if err != nil {
		return nil, err
	}

	req.URL.Query().Add("server", string(server))

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	stages := []Stage{}
	if err := json.NewDecoder(resp.Body).Decode(&stages); err != nil {
		return nil, err
	}

	return stages, nil
}

func (pc *PenguinClient) SendArkPlan(ctx context.Context, payload ArkPlannerRequest) (*ArkPlannerPlan, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", PlannerURL, bytes.NewBuffer(p))
	if err != nil {
		return nil, err
	}

	resp, err := pc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	plan := ArkPlannerPlan{}
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, err
	}

	return &plan, nil
}
