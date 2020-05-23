package main

// Owner contains the full name of the asset owner
type Owner struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

// Content is the text
type Content struct {
	ID   string `json:"contentid"`
	Text string `json:"text"`
}

//Dependents store dependent information
type Dependents struct {
	ID     string   `json:"depid"`
	DepIDs []string `json:"depids"`
}

// Requirement an asset
type Requirement struct {
	ID         string `json:"id"`
	Owner      Owner  `json:"owner"`
	ContentID  string `json:"contentid"`
	Status     string `json:"status"`
	CreateTime string `createtime:"createtime"`
	ShareTime  string `json:"sharetime"`
	AccessTime string `json:"accesstime"`
	DepID      string `json:"depid"`
	IsAccessed bool   `json:"isaccessed"`
}

// SetConditionUsed marks the asses as used
func (req *Requirement) setStatusShared() {
	req.Status = "shared"
}

// Set the status to created on initializing
func (req *Requirement) setInitialStatus() {
	req.Status = "created"
}
