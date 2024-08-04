package main

import (
	"GoProject/block"
	"GoProject/utils"
	wallet "GoProject/wallet"
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"path"
	"strconv"
)

const tempDir = "wallet_server/templates"

type WalletServer struct {
	port    uint16
	gateway string
}

func NewWalletServer(port uint16, gateway string) *WalletServer {
	return &WalletServer{port, gateway}
}

func (ws *WalletServer) GetPort() uint16 {
	return ws.port
}

func (ws *WalletServer) GetGateway() string {
	return ws.gateway
}

func (ws *WalletServer) Index(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		t, _ := template.ParseFiles(path.Join(tempDir, "index.html"))
		t.Execute(w, nil)
	default:
		log.Printf("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) Wallet(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		w.Header().Add("Content-Type", "application/json")
		myWallet := wallet.NewWallet()
		m, _ := myWallet.MarshalJSON()
		io.WriteString(w, string(m[:]))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

// 处理事务并向区块链服务器发送事务
func (ws *WalletServer) CreateTransaction(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		decoder := json.NewDecoder(req.Body)
		var t wallet.TransactionRequest
		err := decoder.Decode(&t)
		if err != nil {
			log.Printf("ERROR: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if !t.Validate() {
			log.Printf("ERROR: Invalid Transaction")
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		publicKey := utils.PublicKeyFromString(*t.SenderPublicKey)
		privateKey := utils.PrivateKeyFromString(*t.SenderPrivateKey, publicKey)
		value, err := strconv.ParseFloat(*t.Value, 32)
		if err != nil {
			log.Printf("ERROR: parse error")
			io.WriteString(w, string(utils.JsonStatus("fail")))
		}
		value32 := float32(value)
		w.Header().Add("Content-type", "application/json")
		io.WriteString(w, string(utils.JsonStatus("success")))
		transaction := wallet.NewTransaction(privateKey, publicKey,
			*t.SenderBlockChainAddress, *t.ReceiverBlockChainAddress, value32)
		signature := transaction.GenerateSignature()
		signatureStr := signature.String()

		bt := &block.TransactionRequest{
			t.SenderBlockChainAddress,
			t.ReceiverBlockChainAddress,
			t.SenderPublicKey,
			&value32, &signatureStr,
		}
		m, _ := json.Marshal(bt)
		buffer := bytes.NewBuffer(m)
		log.Println(ws.GetGateway())
		//向区块链服务器发送交易事务
		url := "http://" + ws.GetGateway() + "/transactions"
		resp, err := http.Post(url, "application/json", buffer)
		if err != nil {
			log.Printf("ERROR: failed to post transaction to gateway: %v", err)
			io.WriteString(w, string(utils.JsonStatus("fail")))
			return
		}
		if resp.StatusCode == 201 {
			io.WriteString(w, string(utils.JsonStatus("success")))
			return
		}
		io.WriteString(w, string(utils.JsonStatus("fail")))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("ERROR: Invalid HTTP Method")
	}
}

func (ws *WalletServer) Run() {
	http.HandleFunc("/", ws.Index)
	http.HandleFunc("/wallet", ws.Wallet)
	http.HandleFunc("/transaction", ws.CreateTransaction)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+strconv.Itoa(int(ws.GetPort())), nil))
}
