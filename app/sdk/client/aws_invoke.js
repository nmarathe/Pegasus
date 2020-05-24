/**
 * Solution to challenge exercise
 * ==============================
 * This code demonstrates how to invoke chaincode using client & channel
 * 
 * The code is "Highly Simplified" to make it easy to read - no promises used deliberately :)
 * 
 * THIS IS NOT PRODUCTION READY CODE
 * 
 * + As there is NO error checks
 * + As the timeout not implemented i.e., indefinite wait if commit event not received
 */

'use strict';
const fs = require('fs');
const Client = require('fabric-client');

// Constants for profile
const CONNECTION_PROFILE_PATH = '../profiles/aws-dev-connection.yaml'
// Client section configuration
const REQ_CLIENT_CONNECTION_PROFILE_PATH = '../profiles/requirements-client.yaml'
const DESGRP_CLIENT_CONNECTION_PROFILE_PATH = '../profiles/designgroup-client.yaml'

// Org & User
const ORG_NAME = 'requirements.oem.com'
const USER_NAME = 'Admin'   
const PEER_NAME = 'peer0.oem.requirements.com'
const CHANNEL_NAME = 'oem-channel'

const CHAINCODE_ID = 'oemcc'
// const CHAINCODE_EVENT = 'transfer'


// Variable to hold the client
var client = {}
// Variable to hold the channel
var channel = {}

main()

async function main(){
    // setup client
    client = await setupClient()

    // Setup the channel instance
    channel = await setupChannel()

    const DATA_FILE_PATH = '../data/Data_128'

    var data = fs.readFileSync(DATA_FILE_PATH)

    var totalSize = 5000
    var assets = []
    for(var i = 4504; i <= 9504; i++) {
        assets.push("Req-1" + i)
    }
    
    console.log('# of transactions - ' + totalSize + ' Data size 128 KB')
    console.log('Transaction start time ' + Math.round((new Date()).getTime() / 1000))

    // for(var i = 4504; i <= 9504; i++) {
    //     await createAssets("Req-1" + i, data.toString())
    // }

    // Share assets
    await shareAssets(assets)

    console.log('Transaction end time ' + Math.round((new Date()).getTime() / 1000))
}

async function shareAssets(assets) {

    let peerName = channel.getChannelPeer(PEER_NAME)

    var tx_id = client.newTransactionID();

    var request = {
        targets: peerName,
        chaincodeId: CHAINCODE_ID,
        fcn: "ShareAssetsBulk",
        args: [JSON.stringify(assets),"{\"firstname\":\"Sam\",\"lastname\":\"Designer\"}"],
        chainId: CHANNEL_NAME,
        txId: tx_id
    };

    let results = await channel.sendTransactionProposal(request);
    // console.log(results)

    // Array of proposal responses *or* error @ index=0
    var proposalResponses = results[0];

    // Original proposal @ index = 1
    var proposal = results[1];
    
     // Broadcast request
     var orderer_request = {
        txId: tx_id,
        proposalResponses: proposalResponses,
        proposal: proposal
    };

    // #4 Request orderer to broadcast the txn
    await channel.sendTransaction(orderer_request);
}

async function createAssets(assetID, payload) {

    // Get the peer for channel. 
    let peerName = channel.getChannelPeer(PEER_NAME)

    // Create a transaction ID
    var tx_id = client.newTransactionID();

    // Create the ChaincodeInvokeRequest - used as arg for sending proposal to endorser(s)
    // https://fabric-sdk-node.github.io/release-1.4/global.html#ChaincodeInvokeRequest
    var request = {
        targets: peerName,
        chaincodeId: CHAINCODE_ID,
        fcn:'NewAsset',
        args: [assetID,"{\"firstname\":\"John\",\"lastname\":\"Doe\"}",payload],
        chainId: CHANNEL_NAME,
        txId: tx_id
    };

    // PHASE-1 of Transaction Flow
    // #1  Send the txn proposal
    // console.log("Channel.sendTransactionProposal " + assetID + " - Done.")
    let results = await channel.sendTransactionProposal(request);

    // Array of proposal responses *or* error @ index=0
    var proposalResponses = results[0];

    // console.log(proposalResponses)

    // Original proposal @ index = 1
	var proposal = results[1];

    // Broadcast request
    var orderer_request = {
        txId: tx_id,
        proposalResponses: proposalResponses,
        proposal: proposal
    };

    // PHASE-2 of Transaction Flow

    // #4 Request orderer to broadcast the txn
    await channel.sendTransaction(orderer_request);
    // console.log("#4 channel.sendTransaction - waiting for Tx Event")
}

/**
 * Initialize the file system credentials store
 * 1. Creates the instance of client using <static> loadFromConfig
 * 2. Loads the client connection profile based on org name
 * 3. Initializes the credential store
 * 4. Loads the user from credential store
 * 5. Sets the user on client instance and returns it
 */
async function setupClient() {

    // setup the instance
    const client = Client.loadFromConfig(CONNECTION_PROFILE_PATH)

    // setup the client part
    if (ORG_NAME == 'requirements.oem.com') {
        client.loadFromConfig(REQ_CLIENT_CONNECTION_PROFILE_PATH)
    } else if (ORG_NAME == 'designgroup.oem.com') {
        client.loadFromConfig(DESGRP_CLIENT_CONNECTION_PROFILE_PATH)
    } else {
        console.log("Invalid Org: ", ORG_NAME)
        process.exit(1)
    }

    // Call the function for initializing the credentials store on file system
    await client.initCredentialStores()
        .then((done) => {
            //console.log("initCredentialStore(): ", done)
        })

    let userContext = await client.loadUserFromStateStore(USER_NAME)
    if (userContext == null) {
        console.log("User NOT found in credstore: ", USER_NAME)
        process.exit(1)
    }

    // set the user context on client
    client.setUserContext(userContext, true)

    return client
}

/**
 * Creates an instance of the Channel class
 */
async function setupChannel() {
    try {
        // Get the Channel class instance from client
        channel = await client.getChannel(CHANNEL_NAME, true)
    } catch (e) {
        console.log("Could NOT create channel: ", CHANNEL_NAME)
        process.exit(1)
    }
    console.log("Created channel object.")

    return channel
}
