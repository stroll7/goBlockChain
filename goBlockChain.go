package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type Block struct {
	nonce        int
	previousHash string
	timestamp    int64
	transactions []string
}

func newBlock(nonce int, previousHash string) *Block {
	b := new(Block)
	b.nonce = nonce
	b.previousHash = previousHash
	b.timestamp = time.Now().Unix()
	return b
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	fmt.Println(string(m))
	return sha256.Sum256([]byte(m))
	/*b.transactions = append(b.transactions, string(hash[:]))*/
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Nonce        int
		PreviousHash string
		Timestamp    int64
		Transactions []string
	}{
		Nonce:        b.nonce,
		PreviousHash: b.previousHash,
		Timestamp:    b.timestamp,
		Transactions: b.transactions,
	})
}

func (b *Block) Print() {
	fmt.Println("nonce:", b.nonce)
	fmt.Println("previousHash:", b.previousHash)
	fmt.Println("timestamp:", b.timestamp)
	fmt.Println("transactions:", b.transactions)
}

type BlockChain struct {
	transactionPool []string
	chain           []*Block
}

func NewBlockChain() *BlockChain {
	bc := new(BlockChain)
	bc.CreateBlock(0, "Init hash")
	return bc
}

func (bc *BlockChain) CreateBlock(nonce int, previousHash string) *Block {
	b := newBlock(nonce, previousHash)
	m, _ := json.Marshal(b)
	sum256 := sha256.Sum256([]byte(m))
	bc.transactionPool = append(bc.transactionPool, string(sum256[:]))
	bc.chain = append(bc.chain, b)
	return b
}

func (bc *BlockChain) Print() {
	for i, b := range bc.chain {
		fmt.Println("%s Chin %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		b.Print()
	}
	fmt.Println("%s\n", strings.Repeat("*", 50))
}

func init() {
	log.SetPrefix("BlockChain: ")
}

func main() {
	/*blockChain := NewBlockChain()
	blockChain.Print()
	blockChain.CreateBlock(5, "hash 1")
	blockChain.Print()
	blockChain.CreateBlock(2, "hash 2")
	blockChain.Print()*/
	block := newBlock(1, "232")
	fmt.Printf("%x\n", block.Hash())
}
