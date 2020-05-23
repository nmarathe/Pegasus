/**
 * Demonstrates the use of ChannelEventHub for
 * (a) Block events (b) Chaincode events
 */
const fs = require('fs');
const Client = require('fabric-client');

// Constants for profile
const CONNECTION_PROFILE_PATH = '../profiles/aws-dev-connection.yaml'
// Client section configuration
const CLIENT_CONNECTION_PROFILE_PATH = '../profiles/simulation-client.yaml'

// Org & User
// const ORG_NAME = 'simulation.com'
const USER_NAME = 'Admin'   
const PEER_NAME = 'peer0.simulation.com'
const CHANNEL_NAME = 'oem-channel'

const CHAINCODE_ID = 'oemcc'
const CHAINCODE_EVENT = 'assetAccessed'

// Variable to hold the client
var client = {}
// Variable to hold the channel
var channel = {}

// Call the main function
main()

async function main() {
    // Setup the client instance
    client = await setupClient()

    // Setup the channel instance
    channel = await setupChannel()

    // Setup block listener
    //setupBlockListener()

    // Setup the chaincode listener
    setupChaincodeListener()
}

/**
 * Setup the chaincode listener
 */
async function setupChaincodeListener(){

    // Create a new instance of the event hub 
    // let eventHub = channel.newChannelEventHub();
    // eventHub.setPeerAddr(PEER_NAME)
    let eventHub = channel.getChannelEventHub(PEER_NAME);

    // console.log(eventHub)

    // Register the Listener
    let chaincodeHandler = await eventHub.registerChaincodeEvent(CHAINCODE_ID,CHAINCODE_EVENT,

            // onEvent - receives Global~ChaincodeEvent
            // https://fabric-sdk-node.github.io/release-1.4/global.html#ChaincodeEvent
            (chaincodeEvent)=>{
                console.log(`\non chaincodeEvent: ${chaincodeEvent.chaincode_id}  ${chaincodeEvent.event_name}  
                ${new String(chaincodeEvent.payload)}`)
            },

            // onError
            ()=>{
                console.log('on Chaincode Event Error!!!')
            }
        )

        eventHub.connect(true, 
            // Connect callback
            ()=>{
                // console.log('chaincodeEvent  connectCallback')
            }
        )

    console.log('chaincodeEvenrHandler started with handler_id=',chaincodeHandler)
        
    let newAssetHandler = await eventHub.registerChaincodeEvent(CHAINCODE_ID,"newAsset", 
        (newAssetEvent) => {
            console.log(`\non newAssetEvent: ${newAssetEvent.chaincode_id} ${newAssetEvent.event_name}
             ${new String(newAssetEvent.payload)}`)

            var object = JSON.parse(newAssetEvent.payload)
            console.log(object.assetid)
            console.log(object.sharetime)
            console.log(object.accesstime)

            console.log('Event receipt time ' + Math.round((new Date()).getTime() / 1000))
        },
        ()=> {
            console.log(`on chaincode Event Error !!!`)
        }
    )

    eventHub.connect(true, 
        ()=> {

        }
    )

    console.log(`chaincodeEventHandler started with handler_id=`,newAssetHandler)

    let assetModifiedHandler = await eventHub.registerChaincodeEvent(CHAINCODE_ID,"assetModified", 
        (assetModifiedEvent) => {
            console.log(`\non assetModified: ${assetModifiedEvent.chaincode_id} ${assetModifiedEvent.event_name} 
            ${new String(assetModifiedEvent.payload)}`)

            var object = JSON.parse(assetModifiedEvent.payload)
            console.log(object.source)
            var d = object.dependents
            console.log(d.length)
        },
        ()=> {
            console.log(`on chaincode Event Error !!!`)
        }
    )

    eventHub.connect(true, 
        ()=> {

        }
    )

    console.log(`chaincodeEventHandler started with handler_id=`,assetModifiedHandler)
}

/**
 * Setup the Block listener
 */
// function    setupBlockListener(){

//     // Get the instance of the event hub
//     let eventHub = channel.getChannelEventHub(PEER_NAME);

//     // Register - returns an integer used for unregistering
//     // Save the integer as you would use it for unregistering
//     let blockHandler = eventHub.registerBlockEvent(

//         // onEvent - receives instance of Block
//         // https://fabric-sdk-node.github.io/release-1.4/global.html#Block
//         (block)=>{
//             console.log('\non Block Event: Number:',block.header.number)
//         },

//         // onError
//         ()=>{
//             console.log('on Block Event Error!!!')
//         }
//     )

//     // Options for the connection to Peer service
//     // https://fabric-sdk-node.github.io/release-1.4/global.html#RegistrationOpts
//     let  connectOptions = {
//         full_block: true   /** Filtered default=false */
//     }

//     eventHub.connect(connectOptions, 
//         // Connect callback
//         ()=>{
//             // console.log('block  connectCallback')
//         }
//     )

    // If uncommeneted - This will stop the listener
    // eventHub.unregisterBlockEvent(blockHandler)

//     console.log('blockHandler started with handler_id=',blockHandler)
// }


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
    // console.log("Created channel object.")

    return channel
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
    client.loadFromConfig(CLIENT_CONNECTION_PROFILE_PATH)

    // setup the client part
    // if (ORG_NAME == 'requirements.oem.com') {
    //     client.loadFromConfig(REQ_CLIENT_CONNECTION_PROFILE_PATH)
    // } else if (ORG_NAME == 'designgroup.oem.com') {
    //     client.loadFromConfig(DESGRP_CLIENT_CONNECTION_PROFILE_PATH)
    // } else {
    //     console.log("Invalid Org: ", ORG_NAME)
    //     process.exit(1)
    // }

    // Call the function for initializing the credentials store on file system
    await client.initCredentialStores()
        .then((done) => {
            // console.log("initCredentialStore(): ", done)
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
