package penguinstats

import (
	"net/http"
	"time"
)

type Server string
type DropType string

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
	processed bool
	stageMap  map[string][]StageDrop
	rawData   []StageDrop
}

type StageDrop struct {
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

type Stage struct {
	Type      string            `json:"stageType,omitempty"`
	ID        string            `json:"stageId,omitempty"`
	ZoneID    string            `json:"zoneId,omitempty"`
	Code      string            `json:"code,omitempty"`
	CodeI18N  map[string]string `json:"code_i18n,omitempty"`
	APCost    int               `json:"apCost,omitempty"`
	DropInfos []StageDropInfo   `json:"dropInfos,omitempty"`
}

type StageDropInfo struct {
	ItemID string   `json:"itemId,omitempty"`
	Type   DropType `json:"dropType,omitempty"`
	Bounds struct {
		Lower int `json:"lower,omitempty"`
		Upper int `json:"upper,omitempty"`
	} `json:"bounds,omitempty"`
}

type ArkPlannerRequest struct {
	ExpDemand       bool           `json:"exp_demand,omitempty"`
	ExtraByProducts bool           `json:"extra_outc,omitempty"`
	LMDDemand       bool           `json:"gold_demand,omitempty"`
	Owned           map[string]int `json:"owned,omitempty"`
	Request         map[string]int `json:"request,omitempty"`
}

type ArkPlannerPlan struct {
	SanityCost int                `json:"cost,omitempty"`
	LMDIncome  int                `json:"gold,omitempty"`
	ExpIncome  int                `json:"exp,omitempty"`
	Stages     []ArkPlanStage     `json:"stages,omitempty"`
	Syntheses  []ArkPlanSynthesis `json:"syntheses,omitempty"`
	Values     []ArkPlanValues    `json:"values,omitempty"`
}

type ArkPlanStage struct {
	Stage         string            `json:"stage,omitempty"`
	EstimatedRuns string            `json:"count,omitempty"`
	ItemWeights   map[string]string `json:"items,omitempty"`
}

type ArkPlanSynthesis struct {
}

type ArkPlanValues struct {
	Level string `json:"level,omitempty"`
	Items []struct {
		Name  string `json:"name,omitempty"`
		Value string `json:"value,omitempty"`
	} `json:"items,omitempty"`
}
