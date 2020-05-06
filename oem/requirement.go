package main

// Owner contains the full name of the asset owner
type Owner struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
}

// Requirement an asset
type Requirement struct {
	ID         string         `json:"id"`
	Owner      Owner          `json:"owner"`
	Text       string         `json:"text"`
	Status     string         `json:"status"`
	ShareTime  string         `json:"sharetime"`
	Dependents []*Requirement `json:"dependents"`
	IsAccessed bool           `json:"isaccessed"`
}

// SetConditionUsed marks the asses as used
func (ba *Requirement) setStatusShared() {
	ba.Status = "shared"
}
