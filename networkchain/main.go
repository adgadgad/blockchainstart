package main
import "https://github.com/adgadgad/blockchainstart/blob/main/networkchain/network.go"
import (
	"fmt" // just for printing something on the screen
	"network"
)

func main(args []string) {
  newblockchain := NewBlockchain() // Initialize the blockchain with the genesis block
  // create 5 blocks and add some transactions
  for i := 1; i <= 15; i++ { // use a for loop to add multiple blocks
    data := fmt.Sprintf("Transaction %d", i) // generate some data for each block
    newblockchain.AddBlock(data)             // add the block to the chain
  }
  // Now print all the blocks and their contents
  for i, block := range newblockchain.Blocks { // iterate on each block
    fmt.Printf("Block ID : %d \n", i)                                        // print the block ID
    fmt.Printf("Timestamp : %d \n", block.Timestamp+int64(i))                // print the timestamp of the block, to make them different, we just add a value i
    fmt.Printf("Hash of the block : %x\n", block.MyBlockHash)                // print the hash of the block
    fmt.Printf("Hash of the previous Block : %x\n", block.PreviousBlockHash) // print the hash of the previous block
    fmt.Printf("All the transactions : %s\n", block.AllData)                 // print the transactions
  } // our blockchain will be printed

  network.StartNode(args[0]) // start the node with the address
}
