package main

//NewAssetPayload for events newAsset and assetAccessed
type NewAssetPayload struct {
	AssetID    string `json:"assetid"`
	CreateTime string `json:"createime"`
}

//ReadAssetPayload for event readasset
type ReadAssetPayload struct {
	AssetID   string `json:"assetid"`
	ShareTime string `json:"sharetime"`
	ReadTime  string `json:"readtime"`
}

// ShareAssetPayload for event shareasset
type ShareAssetPayload struct {
	AssetID    string   `json:"assetid"`
	ShareTime  string   `json:"sharetime"`
	Dependents []string `json:"dependents"`
}

//DepEventPayload event payload
type DepEventPayload struct {
	SourceID   string   `json:"source"`
	ListOfDeps []string `json:"dependents"`
}
