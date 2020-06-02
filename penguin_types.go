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
