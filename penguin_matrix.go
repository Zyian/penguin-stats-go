package penguinstats

func (dm *DropMatrix) GetRaw() []StageDrop {
	return dm.rawData
}

// ProcessMap runs through the entire array returned by Penguin Stats and maps together stageID to an individual array of items drops
func (dm *DropMatrix) ProcessMap() {
	for _, v := range dm.rawData {
		dm.stageMap[v.StageID] = append(dm.stageMap[v.StageID], v)
	}
	dm.processed = true
}

// GetItemsForStage returns the item array for a given stage, if the raw data has not been processed
// it will search through then entire data set and return only the items for the given stage
func (dm *DropMatrix) GetItemsForStage(stageID string) []StageDrop {
	if dm.processed {
		return dm.stageMap[stageID]
	}

	var stageItems []StageDrop
	for _, v := range dm.rawData {
		if v.StageID == stageID {
			stageItems = append(stageItems, v)
		}
	}
	return stageItems
}
