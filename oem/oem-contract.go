package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// OEMContract represents the contract
type OEMContract struct {
	contractapi.Contract
}

// NewAsset creates the new asset
func (cc *OEMContract) NewAsset(ctx contractapi.TransactionContextInterface, id string, owner Owner, text string) error {
	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return errors.New("Unable to communicate wtih world state")
	}

	if existing != nil {
		return fmt.Errorf("Asset with id %s, already exists", id)
	}

	dependents := []*Requirement{}

	ba := new(Requirement)
	ba.ID = id
	ba.Owner = owner
	ba.Text = text
	ba.Dependents = dependents
	ba.IsAccessed = false
	ba.setStatusShared()

	// Get the time at sharing the asset in World state
	shareTime := strconv.FormatInt(time.Now().Unix(), 10)
	ba.ShareTime = shareTime

	// Convert to JSON
	baBytes, _ := json.Marshal(ba)

	// Commit to the ledger
	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to commit the asset to the world state")
	}

	// Emit the event
	newAssetPayload := NewAssetPayload{AssetID: ba.ID, TimeShared: shareTime, AccessTime: ""}
	eventPayload, err := json.Marshal(newAssetPayload)

	if err != nil {
		return errors.New("Unable to marshal event payload to JSON")
	}

	err = ctx.GetStub().SetEvent("newAsset", []byte(eventPayload))

	if err != nil {
		return errors.New("Unable to raise event")
	}

	return nil
}

// CreateDependent using from and to id
func (cc *OEMContract) CreateDependent(ctx contractapi.TransactionContextInterface, fromID string, toID string) (*Requirement, error) {

	// Get the start end
	fromEnd, err := ctx.GetStub().GetState(fromID)

	if err != nil {
		return nil, errors.New("Unable to communicate with the World state")
	}

	if fromEnd == nil {
		return nil, fmt.Errorf("Unable to find the asset with %s", fromID)
	}

	// find the to end
	toEnd, err := ctx.GetStub().GetState(toID)

	if err != nil {
		return nil, errors.New("Unable to communicate with World state for to end")
	}

	if toEnd == nil {
		return nil, fmt.Errorf("Unable to find to end with id %s", toID)
	}

	// load the start end
	start := new(Requirement)
	err = json.Unmarshal(fromEnd, start)

	if err != nil {
		return nil, errors.New("Unable to load start end from JSON")
	}

	// load the to end
	end := new(Requirement)
	err = json.Unmarshal(toEnd, end)

	if err != nil {
		return nil, errors.New("Unable to load the end from JSON")
	}

	// Add the end as dependency on the start
	start.Dependents = append(start.Dependents, end)

	// convert to JSON
	startJSON, _ := json.Marshal(start)

	// save the start back into world state
	err = ctx.GetStub().PutState(fromID, []byte(startJSON))

	return start, nil
}

// UpdateValue updates the asset value
func (cc *OEMContract) UpdateValue(ctx contractapi.TransactionContextInterface, id string, newText string) error {

	// Get the current asset
	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return errors.New("Unable to interact with the world state")
	}

	if existing == nil {
		return fmt.Errorf("Unable to find asset with id %s", id)
	}

	// convert to the BasicAsset
	ba := new(Requirement)
	json.Unmarshal(existing, ba)

	// Update the value
	ba.Text = newText

	// Commit back to ledger
	baBytes, _ := json.Marshal(ba)

	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to update the world state")
	}

	// Go through the dependents and collect their ids
	depIds := []string{}
	for _, req := range ba.Dependents {
		depIds = append(depIds, req.ID)
	}

	// Build the payload
	depPayload := DepEventPayload{SourceID: ba.ID, ListOfDeps: depIds}

	eventPayload, err := json.Marshal(depPayload)
	if err != nil {
		return errors.New("Unable to marshal event payload to JSON")
	}

	// Emit the event
	err = ctx.GetStub().SetEvent("assetModified", []byte(eventPayload))

	if err != nil {
		return errors.New("Unable to raise event")
	}

	return nil
}

// ReadAsset returns the basic asset with id given from the world state
func (cc *OEMContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Requirement, error) {
	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, errors.New("Unable to interact with the world state")
	}

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(Requirement)

	err = json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Requirement", id)
	}

	// Raise the event if this asset is accessed first time after sharing
	if !ba.IsAccessed {
		ba.IsAccessed = true

		// Update world state / ledger
		baBytes, _ := json.Marshal(ba)
		err = ctx.GetStub().PutState(id, []byte(baBytes))

		if err != nil {
			return nil, errors.New("Unable to update World state")
		}

		// Build the payload
		assetAccessPayload := NewAssetPayload{AssetID: ba.ID, TimeShared: ba.ShareTime,
			AccessTime: strconv.FormatInt(time.Now().Unix(), 10)}

		eventPayload, err := json.Marshal(assetAccessPayload)

		if err != nil {
			return nil, errors.New("Unable to marshal event payload to JSON")
		}

		// Emit the event
		err = ctx.GetStub().SetEvent("assetAccessed", []byte(eventPayload))

		if err != nil {
			return nil, errors.New("Unable to raise event")
		}
	}

	return ba, nil
}

// GetAsset returns the basic asset with id given from the world state
func (cc *OEMContract) GetAsset(ctx contractapi.TransactionContextInterface, id string) (*Requirement, error) {
	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, errors.New("Unable to interact with the world state")
	}

	if existing == nil {
		return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", id)
	}

	ba := new(Requirement)

	err = json.Unmarshal(existing, ba)

	if err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type Requirement", id)
	}

	return ba, nil
}

// GetEvaluateTransactions returns functions of ComplexContract not to be tagged as submit
func (cc *OEMContract) GetEvaluateTransactions() []string {
	return []string{"GetAsset"}
}

func main() {
	oemContract := new(OEMContract)

	cc, err := contractapi.NewChaincode(oemContract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
