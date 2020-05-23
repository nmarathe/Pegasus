package main

//NewAssetPayload for events newAsset and assetAccessed
type NewAssetPayload struct {
	AssetID    string `json:"assetid"`
	TimeShared string `json:"sharetime"`
	AccessTime string `json:"accesstime"`
}

//DepEventPayload event payload
type DepEventPayload struct {
	SourceID   string   `json:"source"`
	ListOfDeps []string `json:"dependents"`
}
