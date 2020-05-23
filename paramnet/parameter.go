package main

// Parameter represents an engineering parameter
type Parameter struct {
	ParamID       string  `json:"id"`
	Name          string  `json:"name"`
	MinValue      float32 `json:"min"`
	MaxValue      float32 `json:"max"`
	GoalValue     float32 `json:"goal"`
	ReleaseStatus string  `json:"status"`
}

// SetStausShared sets the ReleaseStatus as shared
func (p *Parameter) SetStausShared() {
	p.ReleaseStatus = "Shared"
}

// ParamPackage represents a collection of sharable parameters
type ParamPackage struct {
	ID         string       `json:"id"`
	Parameters []*Parameter `json:"parameters"`
}
