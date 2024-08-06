package main

import (
	"GoProject/block"
	"GoProject/utils"
	wallet "GoProject/wallet"
	"encoding/json"
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
		log.Printf("private_key %v", minersWallet.PrivateKeyStr())
		log.Printf("public_key %v", minersWallet.PublicKeyStr())
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

// 交易事务
func (bcs *BlockChainServer) Transactions(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	//获取交易事务
	case http.MethodGet:
		w.Header().Add("Content-Type", "application/json")
		bc := bcs.GetBlockChain()
		transactions := bc.TransactionPool()
		m, _ := json.Marshal(struct {
			Transactions []*block.Transaction `json:"transactions"`
			Length       int                  `json:"length"`
		}{
			Transactions: transactions,
			Length:       len(transactions),
		})
		io.WriteString(w, string(m[:]))
	//添加交易事务
	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var t block.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if !t.Validate() {
			log.Println("ERROR: missing field(s)")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		signature := utils.SignatureFromString(*t.Signature)
		bc := bcs.GetBlockChain()
		isCreated := bc.CreateTransaction(*t.SenderBlockChainAddress, *t.ReceiverBlockChainAddress, *t.Value, publicKey, signature)
		w.Header().Add("Content-Type", "application/json")
		var m []byte
		if !isCreated {
			w.WriteHeader(http.StatusBadRequest)
			m = utils.JsonStatus("fail")
		} else {
			w.WriteHeader(http.StatusOK)
			m = utils.JsonStatus("success")
		}
		io.WriteString(w, string(m))
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

// 开始挖矿
func (bcs *BlockChainServer) StartMine(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		bc := bcs.GetBlockChain()
		bc.StartMining()
		m := utils.JsonStatus("success")
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m))
	case http.MethodPost:
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

// 根据地址查看剩余虚拟币
func (bcs *BlockChainServer) Amount(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		blockchinAddress := req.URL.Query().Get("blockchin_address")
		amount := bcs.GetBlockChain().CalculateTotalAmount(blockchinAddress)
		ar := block.AmountResponse{amount}
		m, _ := ar.MarshalJSON()
		w.Header().Add("Content-Type", "application/json")
		io.WriteString(w, string(m[:]))

	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bcs *BlockChainServer) Consensus(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPut:
		bc := bcs.GetBlockChain()
		resolved := bc.ResolveConflicts()

		w.Header().Add("Content-Type", "application/json")
		if resolved {
			io.WriteString(w, string(utils.JsonStatus("success")))
		} else {
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
	default:
		log.Println("ERROR: Invalid HTTP Method")
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (bsc *BlockChainServer) Run() {
	bsc.GetBlockChain().Run()
	http.HandleFunc("/", bsc.GetChain)
	http.HandleFunc("/transactions", bsc.Transactions)
	http.HandleFunc("/mine/start", bsc.StartMine)
	http.HandleFunc("/amount", bsc.Amount)
	http.HandleFunc("/consensus", bsc.Consensus)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(bsc.Port())), nil))
}
