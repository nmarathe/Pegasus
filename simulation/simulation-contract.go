package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimulationContract creates test cases and generates simulation report
type SimulationContract struct {
	contractapi.Contract
}

//CreateTest creates a test case
func (sc *SimulationContract) CreateTest(ctx contractapi.TransactionContextInterface, testID string, paramID string,
	goalVal float32, actualVal float32) (*TestCase, error) {

	// Check if test case with the id already exist
	existing, err := ctx.GetStub().GetState(testID)

	if err != nil {
		return nil, errors.New("Unable to communicate with World state")
	}

	if existing != nil {
		return nil, fmt.Errorf("Asset with ID %s already available in World state", testID)
	}

	// Create new test case
	tc := new(TestCase)
	tc.ID = testID
	tc.ParamID = paramID
	tc.GoalVal = goalVal
	tc.ActualVal = actualVal

	// The test case can only pass if the diff between goal and actuals is +/- 2.0 units
	lower := goalVal - 1.0
	upper := goalVal + 1.0

	if lower <= actualVal && actualVal <= upper {
		tc.Result = "Pass"
	} else {
		tc.Result = "Fail"
	}

	// Convert to json
	testBytes, _ := json.Marshal(tc)

	// Store in the World state
	err = ctx.GetStub().PutState(testID, testBytes)

	if err != nil {
		return nil, errors.New("Unable to save test case to the World state")
	}

	return tc, nil
}

//CreateRun creates a simulation run
func (sc *SimulationContract) CreateRun(ctx contractapi.TransactionContextInterface, runID string,
	testIDs []string, minTests int, minTestPass int) (*SimulationRun, error) {

	//Check if simulation run with ID exists
	existing, err := ctx.GetStub().GetState(runID)

	if err != nil {
		return nil, errors.New("Unable to communicate with the world state")
	}

	if existing != nil {
		return nil, fmt.Errorf("The asset with ID %s already exists", runID)
	}

	// Create the run
	run := new(SimulationRun)
	run.RunID = runID

	// Collect results for each test
	passResults := 0

	// Go over the loop to get the test cases for this run
	for i := 0; i < len(testIDs); i++ {
		testCase, _ := sc.findTest(ctx, testIDs[i])
		run.TestCases = append(run.TestCases, testCase)

		if testCase.Result == "Pass" {
			passResults++
		}
	}

	// A run only passes if minmum number of test cases are run and a certain number of them passes. This is a
	// business constraint
	if minTests <= len(testIDs) && minTestPass <= passResults {
		run.Result = "Pass"
	} else {
		run.Result = "Fail"
	}

	// Conver to JSON bytes to save
	runBytes, err := json.Marshal(run)

	if err != nil {
		return nil, errors.New("Unable to convert to JSON")
	}

	// Write to world state
	err = ctx.GetStub().PutState(runID, runBytes)

	if err != nil {
		return nil, errors.New("Unable to save the state to ledgers")
	}

	return run, nil
}

//CreateReport creates the simulation report
func (sc *SimulationContract) CreateReport(ctx contractapi.TransactionContextInterface, reportID string,
	runIDs []string, minRuns int, minRunsPass int) (*SimulationReport, error) {

	// Check if there is already a report with ID provided
	existing, err := ctx.GetStub().GetState(reportID)

	if err != nil {
		return nil, errors.New("Unable to communicate with World state")
	}

	if existing != nil {
		return nil, fmt.Errorf("The report with ID %s already exists", reportID)
	}

	// Create the Report instance
	simReport := new(SimulationReport)
	simReport.ReportID = reportID
	simReport.Acceptable = false
	passCount := 0

	// Loop through run IDs to load runs
	for i := 0; i < len(runIDs); i++ {
		run, _ := sc.findRun(ctx, runIDs[i])

		if run.Result == "Pass" {
			passCount++
		}
	}

	// Simulation is acceptable only if certain number of runs are carried out a min pass rate
	if passCount >= minRunsPass && minRuns <= len(runIDs) {
		simReport.Acceptable = true
	}

	// Save the state
	simReportBytes, _ := json.Marshal(simReport)

	err = ctx.GetStub().PutState(reportID, simReportBytes)

	if err != nil {
		return nil, errors.New("Unable to store report to World state")
	}

	return simReport, nil
}

//findRun returns the run
func (sc *SimulationContract) findRun(ctx contractapi.TransactionContextInterface, runID string) (*SimulationRun, error) {

	//Check if simulation run with ID exists
	existing, err := ctx.GetStub().GetState(runID)

	if err != nil {
		return nil, errors.New("Unable to communicate with the world state")
	}

	if existing != nil {
		return nil, fmt.Errorf("The asset with ID %s already exists", runID)
	}

	// Convert JSON to struct
	sr := new(SimulationRun)

	err = json.Unmarshal(existing, sr)

	if err != nil {
		return nil, errors.New("Unable to convert JSON to structure")
	}

	return sr, nil
}

//find tests returns test case
func (sc *SimulationContract) findTest(ctx contractapi.TransactionContextInterface, testID string) (*TestCase, error) {
	// Check if test case with the id already exist
	existing, err := ctx.GetStub().GetState(testID)

	if err != nil {
		return nil, errors.New("Unable to communicate with World state")
	}

	if existing != nil {
		return nil, fmt.Errorf("Asset with ID %s already available in World state", testID)
	}

	// Convert from JSON to struct
	tc := new(TestCase)
	err = json.Unmarshal(existing, tc)

	if err != nil {
		return nil, errors.New("Unable to covnert to testcase")
	}

	return tc, nil
}

func main() {
	simContract := new(SimulationContract)

	simcc, err := contractapi.NewChaincode(simContract)

	if err != nil {
		panic(err.Error())
	}

	if err := simcc.Start(); err != nil {
		panic(err.Error())
	}
}
