package network

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"net"
)

// Define some constants for the network protocol
const (
  protocol      = "tcp" // the network protocol to use
  nodeVersion   = 1     // the version of the node software
  commandLength = 12    // the fixed length of the command field in a message
)

// Define some commands for the network protocol
const (
  cmdVersion    = "version"    // a command to send version and blockchain height
  cmdGetBlocks  = "getblocks"  // a command to request blocks from a node
  cmdInv        = "inv"        // a command to send an inventory of blocks or transactions
  cmdGetData    = "getdata"    // a command to request a specific block or transaction
  cmdBlock      = "block"      // a command to send a block
  cmdTx         = "tx"         // a command to send a transaction
  cmdAddr       = "addr"       // a command to send a list of known nodes
  cmdGetAddr    = "getaddr"    // a command to request a list of known nodes
  cmdPing       = "ping"       // a command to check the connectivity of a node
  cmdPong       = "pong"       // a command to respond to a ping
)

// Define a struct for a message
type Message struct {
  Command []byte // the command name
  Payload []byte // the command data
}

// Define a struct for a version command
type Version struct {
  Version    int    // the node version
  BestHeight int    // the blockchain height
  AddrFrom   string // the address of the sender
}

// Define a struct for an inventory command
type Inv struct {
  AddrFrom string   // the address of the sender
  Type     string   // the type of the inventory (block or tx)
  Items    [][]byte // the hashes of the items
}

// Define a struct for a getdata command
type GetData struct {
  AddrFrom string // the address of the sender
  Type     string // the type of the data (block or tx)
  ID       []byte // the hash of the data
}

// Define a struct for a block command
type Block struct {
  AddrFrom string // the address of the sender
  Block    []byte // the serialized block
}

// Define a struct for a transaction command
type Tx struct {
  AddrFrom    string    // the address of the sender
  Transaction []byte    // the serialized transaction
}

// Define a struct for an address command
type Addr struct {
  AddrList []string // the list of known node addresses
}

// Define a struct for a ping command
type Ping struct {
  Nonce int64 // a random number to identify the ping
}

// Define a struct for a pong command
type Pong struct {
  Nonce int64 // the same number as the ping
}

// Define a global variable for the node address
var nodeAddress string

// Define a global variable for the known nodes
var knownNodes = []string{"localhost:3000"} // a list of node addresses, starting with the first node
// Define a function to start a node
func StartNode(address string) {
  nodeAddress = address // set the node address
  ln, err := net.Listen(protocol, address) // create a listener for the node
  if err != nil {
    log.Panic(err) // handle any errors
  }
  defer ln.Close() // close the listener when done
  bc := NewBlockchain() // create a new blockchain for the node
  if address != knownNodes[0] { // if the node is not the first node
    sendVersion(knownNodes[0], bc) // send the version and height to the first node
  }
  for { // loop forever
    conn, err := ln.Accept() // accept incoming connections
    if err != nil {
      log.Panic(err) // handle any errors
    }
    go handleConnection(conn, bc) // handle the connection in a separate goroutine
  }
}

// Define a function to handle a connection
func handleConnection(conn net.Conn, bc *Blockchain) {
  defer conn.Close() // close the connection when done
  request := make([]byte, commandLength) // create a buffer for the request
  _, err := conn.Read(request) // read the request from the connection
  if err != nil {
    log.Panic(err) // handle any errors
  }
  command := bytesToCommand(request) // convert the request to a command
  switch command { // switch on the command
  case cmdVersion: // if the command is version
    handleVersion(request, bc) // handle the version command
  case cmdGetBlocks: // if the command is getblocks
    handleGetBlocks(request, bc) // handle the getblocks command
  case cmdInv: // if the command is inv
    handleInv(request, bc) // handle the inv command
  case cmdGetData: // if the command is getdata
    handleGetData(request, bc) // handle the getdata command
  case cmdBlock: // if the command is block
    handleBlock(request, bc) // handle the block command
  case cmdTx: // if the command is tx
    handleTx(request, bc) // handle the tx command
  case cmdAddr: // if the command is addr
    handleAddr(request, bc) // handle the addr command
  case cmdGetAddr: // if the command is getaddr
    handleGetAddr(request, bc) // handle the getaddr command
  case cmdPing: // if the command is ping
    handlePing(request, bc) // handle the ping command
  case cmdPong: // if the command is pong
    handlePong(request, bc) // handle the pong command
  default: // if the command is unknown
    fmt.Println("Unknown command") // print a message
  }
}

// Define a function to convert bytes to a command
func bytesToCommand(data []byte) string {
  var command []byte // create a buffer for the command
  for _, b := range data { // iterate over the data
    if b != 0x0 { // if the byte is not zero
      command = append(command, b) // append it to the command
    }
  }
  return fmt.Sprintf("%s", command) // return the command as a string
}

// Define a function to convert a command to bytes
func commandToBytes(command string) []byte {
  var data [commandLength]byte // create a buffer for the data
  for i, c := range command { // iterate over the command
    data[i] = byte(c) // set the byte at the index
  }
  return data[:] // return the data as a slice
}

// Define a function to send a message to a node
func sendData(address string, data []byte) {
  conn, err := net.Dial(protocol, address) // create a connection to the node
  if err != nil {
    fmt.Printf("%s is not available\n", address) // print a message if the node is not available
    return
  }
  defer conn.Close() // close the connection when done
  _, err = conn.Write(data) // write the data to the connection
  if err != nil {
    log.Panic(err) // handle any errors
  }
}

// Define a function to send a version command to a node
func sendVersion(address string, bc *Blockchain) {
  bestHeight := bc.GetBestHeight() // get the best height of the blockchain
  payload := gobEncode(Version{nodeVersion, bestHeight, nodeAddress}) // encode the version struct into a payload
  message := append(commandToBytes(cmdVersion), payload...) // append the command and the payload
  sendData(address, message) // send the message to the node
}

// Define a function to handle a version command from a node
func handleVersion(request []byte, bc *Blockchain) {
  var payload Version // create a buffer for the payload
  gobDecode(request[commandLength:], &payload) // decode the request into the payload
  peerVersion := payload.Version // get the peer version
  peerBestHeight := payload.BestHeight // get the peer best height
  peerAddress := payload.AddrFrom // get the peer address
  fmt.Printf("Received version %d and best height %d from %s\n", peerVersion, peerBestHeight, peerAddress) // print a message
  if peerVersion < nodeVersion { // if the peer version is lower than the node version
    sendVersion(peerAddress, bc) // send the node version and height to the peer
  } else if peerVersion > nodeVersion { // if the peer version is higher than the node version
    fmt.Println("Please update your node software") // print a message
  }
  if peerBestHeight > bc.GetBestHeight() { // if the peer best height is higher than the node best height
    sendGetBlocks(peerAddress) // send a getblocks command to the peer
  }
  if !nodeIsKnown(peerAddress) { // if the peer address is not known
    knownNodes = append(knownNodes, peerAddress) // add it to the known nodes
  }
}

// Define a function to send a transaction command to a node
func sendTx(address string, tx *Transaction) {
  payload := gobEncode(Tx{nodeAddress, tx.Serialize()}) // encode the tx struct into a payload
  message := append(commandToBytes(cmdTx), payload...) // append the command and the payload
  sendData(address, message) // send the message to the node
}

// Define a function to handle a transaction command from a node
func handleTx(request []byte, bc *Blockchain) {
  var payload Tx // create a buffer for the payload
  gobDecode(request[commandLength:], &payload) // decode the request into the payload
  peerAddress := payload.AddrFrom // get the peer address
  txData := payload.Transaction // get the transaction data
  tx := DeserializeTransaction(txData) // deserialize the transaction
  fmt.Println("Received a new transaction") // print a message
  bc.AddTxToMempool(tx) // add the transaction to the mempool
  fmt.Printf("Added transaction %x\n", tx.ID) // print a message
  if nodeAddress == knownNodes[0] { // if the node is the first node
    for _, node := range knownNodes { // iterate over the known nodes
      if node != nodeAddress && node != peerAddress { // if the node is not the sender or the receiver
        sendInv(node, "tx", [][]byte{tx.ID}) // send an inv command with the transaction hash to the node
      }
    }
  } else { // if the node is not the first node
    if len(bc.Mempool) >= 2 && len(bc.Mempool)%2 == 0 { // if the mempool has enough transactions to mine a new block
      MineBlock(bc) // mine a new block
    }
  }
}

// Define a function to send an address command to a node
func sendAddr(address string) {
  payload := gobEncode(Addr{knownNodes}) // encode the addr struct into a payload
  message := append(commandToBytes(cmdAddr), payload...) // append the command and the payload
  sendData(address, message) // send the message to the node
}

// Define a function to handle an address command from a node
func handleAddr(request []byte, bc *Blockchain) {
  var payload Addr // create a buffer for the payload
  gobDecode(request[commandLength:], &payload) // decode the request into the payload
  peerAddressList := payload.AddrList // get the peer address list
  for _, address := range peerAddressList { // iterate over the addresses
    if !nodeIsKnown(address) { // if the address is not known
      knownNodes = append(knownNodes, address) // add it to the known nodes
    }
  }
}

// Define a function to send a getaddr command to a node
func sendGetAddr(address string) {
  payload := gobEncode(GetAddr{nodeAddress}) // encode the getaddr struct into a payload
  message := append(commandToBytes(cmdGetAddr), payload...) // append the command and the payload
  sendData(address, message) // send the message to the node
}

// Define a function to handle a getaddr command from a node
func handleGetAddr(request []byte, bc *Blockchain) {
  var payload GetAddr // create a buffer for the payload
  gobDecode(request[commandLength:], &payload) // decode the request into the payload
  peerAddress := payload.AddrFrom // get the peer address
  sendAddr(peerAddress) // send an addr command with the known nodes to the peer
}

// Define a function to send a ping command to a node
func sendPing(address string, nonce int64) {
  payload := gobEncode(Ping{nonce}) // encode the ping struct into a payload
  message := append(commandToBytes(cmdPing), payload...) // append the command and the payload
  sendData(address, message) // send the message to the node
}

// Define a function to handle a ping command from a node
func handlePing(request []byte, bc *Blockchain) {
  var payload Ping // create a buffer for the payload
  gobDecode(request[commandLength:], &payload) // decode the request into the payload
  peerAddress := payload.AddrFrom // get the peer address
  peerNonce := payload.Nonce // get the peer nonce
  sendPong(peerAddress, peerNonce) // send a pong command with the same nonce to the peer
}

// Define a function to send a pong command to a node
func sendPong(address string, nonce int64) {
  payload := gobEncode(Pong{nonce}) // encode the pong struct into a payload
  message := append(commandToBytes(cmdPong), payload...) // append the command and the payload
  sendData(address, message) // send the message to the node
}

// Define a function to handle a pong command from a node
func handlePong(request []byte, bc *Blockchain) {
  var payload Pong // create a buffer for the payload
  gobDecode(request[commandLength:], &payload) // decode the request into the payload
  peerAddress := payload.AddrFrom // get the peer address
  peerNonce := payload.Nonce // get the peer nonce
  fmt.Printf("Received pong %d from %s\n", peerNonce, peerAddress) // print a message
}

// Define a function to check if a node is known
func nodeIsKnown(address string) bool {
  for _, node := range knownNodes { // iterate over the known nodes
    if node == address { // if the node matches the address
      return true // return true
    }
  }
  return false // return false
}

// Define a function to encode a struct into a byte slice
func gobEncode(data interface{}) []byte {
  var buffer bytes.Buffer // create a buffer
  encoder := gob.NewEncoder(&buffer) // create a new encoder
  err := encoder.Encode(data) // encode the data into the buffer
  if err != nil {
    log.Panic(err) // handle any errors
  }
  return buffer.Bytes() // return the buffer as a byte slice
}

// Define a function to decode a byte slice into a struct
func gobDecode(data []byte, target interface{}) {
  reader := bytes.NewReader(data) // create a reader from the data
  decoder := gob.NewDecoder(reader) // create a new decoder
  err := decoder.Decode(target) // decode the data into the target
  if err != nil {
    log.Panic(err) // handle any errors
  }
}
