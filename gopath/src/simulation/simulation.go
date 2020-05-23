package main

//TestCase represets test data
type TestCase struct {
	ID        string  `json:"testid"`
	ParamID   string  `json:"paramid"`
	GoalVal   float32 `json:"goal"`
	ActualVal float32 `json:"actual"`
	Result    string  `json:"result"`
}

//SimulationRun represents collection of test cases
type SimulationRun struct {
	RunID     string      `json:"runid"`
	TestCases []*TestCase `json:"testcases"`
	Result    string      `json:"result"`
}

// SimulationReport summarises the activity
type SimulationReport struct {
	ReportID   string           `json:"reportid"`
	Runs       []*SimulationRun `json:"runs"`
	Acceptable bool             `json:"acceptable"`
}
