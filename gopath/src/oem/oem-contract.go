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

	// Create and persist new Content
	contents := Content{ID: strconv.FormatInt(time.Now().Unix(), 10), Text: text}
	contentBytes, _ := json.Marshal(contents)
	ctx.GetStub().PutState(contents.ID, contentBytes)

	// Create Dependents and persist
	dependents := Dependents{ID: strconv.FormatInt(time.Now().Unix(), 10), DepIDs: []string{}}
	dependentBytes, _ := json.Marshal(dependents)
	ctx.GetStub().PutState(dependents.ID, dependentBytes)

	ba := Requirement{}
	ba.ID = id
	ba.Owner = owner
	ba.ContentID = contents.ID
	ba.DepID = dependents.ID
	ba.IsAccessed = false
	ba.setInitialStatus()

	// Get the time at creating the asset in World state
	createTime := strconv.FormatInt(time.Now().Unix(), 10)
	ba.CreateTime = createTime

	// Convert to JSON
	baBytes, _ := json.Marshal(ba)

	// Commit to the ledger
	err = ctx.GetStub().PutState(id, []byte(baBytes))

	if err != nil {
		return errors.New("Unable to commit the asset to the world state")
	}

	// Emit the event
	newAssetPayload := NewAssetPayload{AssetID: ba.ID, CreateTime: createTime}
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

// ShareAssetsBulk creates requirements in BULK on peer
func (cc *OEMContract) ShareAssetsBulk(ctx contractapi.TransactionContextInterface, input []string, owner Owner) error {

	// Call the ShareAsset function for each element of the requirement array
	for i := 0; i < len(input); i++ {

		depIDs, err := cc.ShareAsset(ctx, input[i], owner)
		if err != nil {
			return err
		}

		// Build the payload
		shareAssetPayload := ShareAssetPayload{AssetID: input[i], ShareTime: strconv.FormatInt(time.Now().Unix(), 10),
			Dependents: depIDs}

		eventPayload, err := json.Marshal(shareAssetPayload)
		if err != nil {
			return errors.New("Unable to marshal event payload to JSON")
		}

		// Emit the event
		err = ctx.GetStub().SetEvent("assetShared", []byte(eventPayload))

		if err != nil {
			return errors.New("Unable to raise event")
		}
	}

	return nil
}

// ShareAsset shares asset by changing status
func (cc *OEMContract) ShareAsset(ctx contractapi.TransactionContextInterface, assetID string, owner Owner) ([]string, error) {

	// Get the current asset
	existing, err := ctx.GetStub().GetState(assetID)

	if err != nil {
		return nil, errors.New("Unable to interact with the world state")
	}

	if existing == nil {
		return nil, fmt.Errorf("Unable to find asset with id %s", assetID)
	}

	// convert to the BasicAsset
	req := Requirement{}
	json.Unmarshal(existing, &req)

	// set new owner
	req.Owner = owner

	// Update the status
	req.setStatusShared()

	// Get the time at sharing the asset in World state
	req.ShareTime = strconv.FormatInt(time.Now().Unix(), 10)

	// Commit back to ledger
	baBytes, _ := json.Marshal(req)
	err = ctx.GetStub().PutState(assetID, []byte(baBytes))

	if err != nil {
		return nil, errors.New("Unable to update the world state")
	}

	// Return dependents for shared assets
	existingDep, _ := ctx.GetStub().GetState(req.DepID)

	dependents := Dependents{}
	json.Unmarshal(existingDep, &dependents)

	return dependents.DepIDs, nil
}

// CreateDependent using from and to id
func (cc *OEMContract) CreateDependent(ctx contractapi.TransactionContextInterface, fromID string,
	toIDs []string) (*Requirement, error) {

	// Get the start end
	fromEnd, err := ctx.GetStub().GetState(fromID)

	if err != nil {
		return nil, errors.New("Unable to communicate with the World state")
	}

	if fromEnd == nil {
		return nil, fmt.Errorf("Unable to find the asset with %s", fromID)
	}

	// load the start end
	start := new(Requirement)
	err = json.Unmarshal(fromEnd, start)

	if err != nil {
		return nil, errors.New("Unable to load start end from JSON")
	}

	// Pesrsist Dependent to World state
	dependents := new(Dependents)
	dependents.ID = strconv.FormatInt(time.Now().Unix(), 10)
	dependents.DepIDs = toIDs

	dependentBytes, _ := json.Marshal(dependents)
	ctx.GetStub().PutState(dependents.ID, dependentBytes)

	// Add the end as dependency on the start
	start.DepID = dependents.ID

	// convert to JSON
	startBytes, _ := json.Marshal(start)

	// save the start back into world state
	err = ctx.GetStub().PutState(fromID, startBytes)

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

	// Get the contents for this requirement
	existingContent, _ := ctx.GetStub().GetState(ba.DepID)
	content := new(Content)
	json.Unmarshal(existingContent, content)

	// set the new text
	content.Text = newText

	// Commit back to ledger
	contentBytes, _ := json.Marshal(content)
	err = ctx.GetStub().PutState(content.ID, contentBytes)

	if err != nil {
		return errors.New("Unable to update the world state")
	}

	// Get the dependents
	dependents := new(Dependents)
	existingDep, _ := ctx.GetStub().GetState(ba.DepID)
	json.Unmarshal(existingDep, dependents)

	// Build the payload
	depPayload := DepEventPayload{SourceID: ba.ID, ListOfDeps: dependents.DepIDs}

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
		assetAccessPayload := ReadAssetPayload{AssetID: ba.ID, ShareTime: ba.ShareTime,
			ReadTime: strconv.FormatInt(time.Now().Unix(), 10)}

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
