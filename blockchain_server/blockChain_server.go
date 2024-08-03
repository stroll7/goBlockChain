package main

import (
	"GoProject/block"
	"GoProject/wallet"
	"io"
	"log"
	"net/http"
	"strconv"
)

var cache map[string]*block.BlockChain = make(map[string]*block.BlockChain)

type BlockChainServer struct {
	port uint16
}

func NewBlockChainServer(port uint16) *BlockChainServer {
	return &BlockChainServer{port}
}

func (bcs *BlockChainServer) Port() uint16 {
	return bcs.port
}

// 获取区块链,区块链不存在就创建
func (bcs *BlockChainServer) GetBlockChain() *block.BlockChain {
	bc, ok := cache["blockchain"]
	if !ok {
		//创建当前节点钱包
		minersWallet := wallet.NewWallet()
		//使用当前钱包地址作为节点,加上端口创建区块链
		bc = block.NewBlockChain(minersWallet.BlockChainAddress(), bcs.Port())
		cache["blockchain"] = bc
		log.Printf("private_key %v", minersWallet.PrivateKey())
		log.Printf("public_key %v", minersWallet.PublicKey())
		log.Printf("blockchain_address %v", minersWallet.BlockChainAddress())
	}
	return bc
}

// 获取并序列化区块链
func (bcs *BlockChainServer) GetChain(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockChain()
		m, _ := bc.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func HelloWord(w http.ResponseWriter, req *http.Request) {
	io.WriteString(w, "hello block chain")
}
func (bsc *BlockChainServer) Run() {
	http.HandleFunc("/", bsc.GetChain)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bsc.Port())), nil))
}
