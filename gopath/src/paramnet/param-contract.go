package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ParamContract is contract
type ParamContract struct {
	contractapi.Contract
}

// CreateParam creates the new asset and returns it
func (cc *ParamContract) CreateParam(ctx contractapi.TransactionContextInterface, id string, name string, minval float32,
	maxVal float32, goalVal float32) (*Parameter, error) {

	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, errors.New("Unable to communicate wtih world state")
	}

	if existing != nil {
		return nil, fmt.Errorf("Asset with id %s, already exists", id)
	}

	p := new(Parameter)
	p.ParamID = id
	p.Name = name
	p.MinValue = minval
	p.MaxValue = maxVal
	p.GoalValue = goalVal
	p.SetStausShared()

	// Convert to JSON
	pBytes, _ := json.Marshal(p)

	// Commit to the ledger
	err = ctx.GetStub().PutState(id, pBytes)

	if err != nil {
		return nil, errors.New("Unable to commit the asset to the world state")
	}

	return p, nil
}

// CreatePackage creates parameter package to share
func (cc *ParamContract) CreatePackage(ctx contractapi.TransactionContextInterface, pkgID string,
	paramID []string) (*ParamPackage, error) {

	// Get the start end
	existingPkg, err := ctx.GetStub().GetState(pkgID)

	if err != nil {
		return nil, errors.New("Unable to communicate with the World state")
	}

	if existingPkg != nil {
		return nil, fmt.Errorf("Asset with ID %s already exists", pkgID)
	}

	// Create package
	paramPkg := new(ParamPackage)
	paramPkg.ID = pkgID

	// Loop over IDs and load parameters
	for i := 0; i < len(paramID); i++ {

		existing, _ := ctx.GetStub().GetState(paramID[i])

		param := new(Parameter)
		json.Unmarshal(existing, param)

		paramPkg.Parameters = append(paramPkg.Parameters, param)
	}

	// convert to JSON
	pkgJSON, _ := json.Marshal(paramPkg)

	// save the start back into world state
	err = ctx.GetStub().PutState(pkgID, pkgJSON)

	return paramPkg, nil
}

// GetPackage returns the basic asset with id given from the world state
func (cc *ParamContract) GetPackage(ctx contractapi.TransactionContextInterface, pkgID string) (*ParamPackage, error) {

	existing, err := ctx.GetStub().GetState(pkgID)

	if err != nil {
		return nil, errors.New("Unable to interact with the world state")
	}

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", pkgID)
	}

	// Load the package from ledger
	paramPkg := new(ParamPackage)

	err = json.Unmarshal(existing, paramPkg)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Requirement", pkgID)
	}

	return paramPkg, nil
}

func main() {
	paramContract := new(ParamContract)

	paramcc, err := contractapi.NewChaincode(paramContract)

	if err != nil {
		panic(err.Error())
	}

	if err := paramcc.Start(); err != nil {
		panic(err.Error())
	}
}
