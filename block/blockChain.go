package block

import (
	utils "GoProject/utils"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

//使用共识算法计算挖矿时间

// 挖矿难度
const (
	MINING_DIFFICULTY = 3
	MINING_SENDER     = "THE BLOCKCHAIN"
	MINING_REWARD     = 1.0
	MINING_TIMER_SEC  = 20

	//区块链端口开始反胃
	BLOCKCHAIN_PORT_RANGE_START = 5000
	BLOCKCHAIN_PORT_RANGE_END   = 5003
	//使用IP范围  0~1：只使用一个
	NEIGHBOR_IP_RANGE_START = 0
	NEIGHBOR_IP_RANGE_END   = 1
	//区块链同步时间
	BLOCKCHAIN_NEIGHBOR_SYNC_SEC = 20
)

// 定义一个区块对象
type Block struct {
	timestamp    int64
	nonce        int
	previousHash [32]byte
	transactions []*Transaction
}

func (b *Block) PreviousHash() [32]byte {
	return b.previousHash
}

func (b *Block) Nonce() int {
	return b.nonce
}

func (b *Block) Transactions() []*Transaction {
	return b.transactions
}

func newBlock(nonce int, previousHash [32]byte, transactions []*Transaction) *Block {
	b := new(Block)
	b.timestamp = time.Now().UnixNano()
	b.nonce = nonce
	b.previousHash = previousHash
	b.transactions = transactions
	return b
}

func (b *Block) Hash() [32]byte {
	m, _ := json.Marshal(b)
	return sha256.Sum256([]byte(m))
}

// 重写序列化方法(开头不能是小写)
func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Timestamp    int64          `json:"timestamp"`
		Nonce        int            `json:"nonce"`
		PreviousHash string         `json:"previous_hash"`
		Transactions []*Transaction `json:"transactions"`
	}{
		Timestamp:    b.timestamp,
		Nonce:        b.nonce,
		PreviousHash: fmt.Sprintf("%x", b.previousHash),
		Transactions: b.transactions,
	})
}

// 反序列化
func (b *Block) UnmarshalJSON(data []byte) error {
	var previousHash string
	v := &struct {
		Timestamp    *int64          `json:"timestamp"`
		Nonce        *int            `json:"nonce"`
		PreviousHash *string         `json:"previous_hash"`
		Transactions *[]*Transaction `json:"transactions"`
	}{
		Timestamp:    &b.timestamp,
		Nonce:        &b.nonce,
		PreviousHash: &previousHash,
		Transactions: &b.transactions,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	ph, _ := hex.DecodeString(*v.PreviousHash)
	copy(b.previousHash[:], ph[:32])
	return nil
}

func (b *Block) Print() {
	fmt.Printf("timestamp:           %d\n", b.timestamp)
	fmt.Printf("nonce:               %d\n", b.nonce)
	fmt.Printf("previous_hash:       %x\n", b.previousHash)
	for _, t := range b.transactions {
		t.Print()
	}
}

type BlockChain struct {
	transactionPool   []*Transaction
	chain             []*Block   //当前区块链区块
	blockChainAddress string     //区块链节点地址
	port              uint16     //当前节点监听端口号
	mux               sync.Mutex //互斥锁

	neighbors    []string   //附近节点列表
	muxNeighbors sync.Mutex //节点同步锁
}

// 创建区块链同时创建第一个区块
func NewBlockChain(blockChainAddress string, port uint16) *BlockChain {
	b := &Block{}
	bc := new(BlockChain)
	bc.blockChainAddress = blockChainAddress
	//nonce为0,使用空区块的hash,创建第一个区块
	bc.CreateBlock(0, b.Hash())
	bc.port = port
	return bc
}

func (bc *BlockChain) Chain() []*Block {
	return bc.chain
}

func (bc *BlockChain) Run() {
	bc.StartSyncNeighbors()
	bc.ResolveConflicts()
}

func (bc *BlockChain) SetNeighbors() {
	bc.neighbors = utils.FindNeighbors(
		utils.GetHost(), bc.port,
		NEIGHBOR_IP_RANGE_START, NEIGHBOR_IP_RANGE_END,
		BLOCKCHAIN_PORT_RANGE_START, BLOCKCHAIN_PORT_RANGE_END)
	log.Printf("%v", bc.neighbors)
}

func (bc *BlockChain) SyncNeighbors() {
	//上锁,避免多次同步
	bc.muxNeighbors.Lock()
	defer bc.muxNeighbors.Unlock()
	bc.SetNeighbors()
}

func (bc *BlockChain) StartSyncNeighbors() {
	bc.SyncNeighbors()
	//每20秒执行一次同步方法，同步区块链节点
	_ = time.AfterFunc(time.Second*BLOCKCHAIN_NEIGHBOR_SYNC_SEC, bc.StartSyncNeighbors)
}

func (bc *BlockChain) TransactionPool() []*Transaction {
	return bc.transactionPool
}

func (bc *BlockChain) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Blocks []*Block `json:"block"`
	}{
		Blocks: bc.chain,
	})
}

func (bc *BlockChain) UnmarshalJSON(data []byte) error {
	v := &struct {
		Blocks *[]*Block `json:"block"`
	}{
		Blocks: &bc.chain,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

func (bc *BlockChain) CreateBlock(nonce int, previousHash [32]byte) *Block {
	b := newBlock(nonce, previousHash, bc.transactionPool)
	bc.chain = append(bc.chain, b)
	bc.transactionPool = []*Transaction{}
	return b
}

func (bc *BlockChain) LastBlock() *Block {
	//返回最后一个区块
	return bc.chain[len(bc.chain)-1]
}

func (bc *BlockChain) Print() {
	for i, block := range bc.chain {
		fmt.Printf("%s Chain %d %s\n", strings.Repeat("=", 25), i, strings.Repeat("=", 25))
		block.Print()
	}
	fmt.Printf("%s\n", strings.Repeat("*", 25))
}

// 复制事务
func (bc *BlockChain) CopyTransactionPool() []*Transaction {
	transactions := make([]*Transaction, 0)
	for _, t := range bc.transactionPool {
		transactions = append(transactions,
			NewTransaction(t.senderBlockchainAddress,
				t.recipientBlockchainAddress,
				t.value))
	}
	return transactions
}

// 检验找的哈希是否满足工作量证明的要求
func (bc *BlockChain) ValidProof(nonce int, previousHash [32]byte, transactions []*Transaction, difficulty int) bool {
	//比较新区块哈希值的基准(前面是几个0,控制挖矿难度)
	//0越少,找到有效哈希值所需的计算工作越少，挖矿相对容易
	zeros := strings.Repeat("0", difficulty)
	guessBlock := Block{0, nonce, previousHash, transactions}
	//获得区块的哈希值
	guessHashStr := fmt.Sprintf("%x", guessBlock.Hash())
	//fmt.Println(guessHashStr)
	return guessHashStr[:difficulty] == zeros
}

func (bc *BlockChain) ProofOfWork() int {
	//上个区块的事务
	transactions := bc.CopyTransactionPool()
	//上个区块的hash
	previousHash := bc.LastBlock().Hash()
	nonce := 0
	for !bc.ValidProof(nonce, previousHash, transactions, MINING_DIFFICULTY) {
		nonce += 1
	}
	return nonce
}

func (bc *BlockChain) Mining() bool {
	bc.mux.Lock()         //加锁
	defer bc.mux.Unlock() //执行完成解锁
	//有交易产生时才能挖矿
	if len(bc.transactionPool) == 0 {
		return false
	}

	bc.AddTransaction(MINING_SENDER, bc.blockChainAddress, MINING_REWARD, nil, nil)
	nonce := bc.ProofOfWork()
	previousHash := bc.LastBlock().Hash()
	bc.CreateBlock(nonce, previousHash)
	log.Println("action=mining, status=success")

	for _, n := range bc.neighbors {
		endpoint := fmt.Sprintf("http://%s/consensus", n)
		client := &http.Client{}
		req, _ := http.NewRequest("PUT", endpoint, nil)
		resp, _ := client.Do(req)
		log.Printf("%v", resp)
	}
	return true
}

// 开始挖矿
func (bc *BlockChain) StartMining() {
	bc.Mining()
	_ = time.AfterFunc(time.Second*MINING_TIMER_SEC, bc.StartMining)
}

// 根据区块链地址获取虚拟币数量
func (bc *BlockChain) CalculateTotalAmount(blockChainAddress string) float32 {
	var totalAmount float32 = 0.0
	for _, c := range bc.chain {
		transactions := c.transactions
		for _, t := range transactions {
			value := t.value
			if t.recipientBlockchainAddress == blockChainAddress {
				totalAmount += value
			}
			if t.senderBlockchainAddress == blockChainAddress {
				totalAmount -= value
			}
		}
	}
	return totalAmount
}

// 验证区块链有效性
func (bc *BlockChain) ValidChain(chain []*Block) bool {
	//获取初始区块
	preBlock := chain[0]
	//从第二个区块开始遍历区块链
	currentIndex := 1
	for currentIndex < len(chain) {
		b := chain[currentIndex]
		//检查与前一个区块的哈希值相匹配
		if b.previousHash != preBlock.Hash() {
			return false
		}
		//验证工作量证明
		if !bc.ValidProof(b.nonce, b.previousHash, b.transactions, MINING_DIFFICULTY) {
			return false
		}
		//替换区块,继续验证下一个区块
		preBlock = b
		currentIndex += 1
	}
	return true
}

func (bc *BlockChain) ResolveConflicts() bool {
	var longestChain []*Block = nil
	maxLeng := len(bc.chain)
	//遍历区块链节点
	for _, n := range bc.neighbors {
		//对每个邻居节点发起 HTTP GET 请求，获取它们的区块链。
		endpoint := fmt.Sprintf("http://%s/chain", n)
		resp, _ := http.Get(endpoint)
		if resp.StatusCode == http.StatusOK {
			//解析响应体中的 BlockChain 结构
			var bcResp BlockChain
			decoder := json.NewDecoder(resp.Body)
			_ = decoder.Decode(&bcResp)
			chain := bcResp.chain
			//判断获取的链是否比当前的链更长,更长则验证其有消息
			if len(chain) > maxLeng && bc.ValidChain(chain) {
				//更新链长度
				maxLeng = len(bc.chain)
				//替换当前节点的链
				longestChain = chain
			}
		}
	}
	//longestChain不为空，说明替换成功
	if longestChain != nil {
		bc.chain = longestChain
		log.Printf("Resolve conflicts replaced")
		return true
	}
	log.Printf("Resolve conflicts not replaced")
	return false
}

// 定义一个事务对象
type Transaction struct {
	senderBlockchainAddress    string
	recipientBlockchainAddress string
	value                      float32
}

func NewTransaction(sender string, recipient string, value float32) *Transaction {
	return &Transaction{sender, recipient, value}
}

func (t *Transaction) Print() {
	fmt.Printf("%s\n", strings.Repeat("-", 50))
	fmt.Printf("sender_blockchain_address: %s\n", t.senderBlockchainAddress)
	fmt.Printf("recipient_blockchain_address: %s\n", t.recipientBlockchainAddress)
	fmt.Printf("value: %.1f\n", t.value)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Sender    string  `json:"sender_blockchain_address"`
		Recipient string  `json:"recipient_blockchain_address"`
		Value     float32 `json:"value"`
	}{
		Sender:    t.senderBlockchainAddress,
		Recipient: t.recipientBlockchainAddress,
		Value:     t.value,
	})
}
func (bc *BlockChain) CreateTransaction(sender string, recipient string, value float32,
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature) bool {
	isTransaction := bc.AddTransaction(sender, recipient, value, senderPublicKey, s)
	return isTransaction
}

func (bc *BlockChain) AddTransaction(sender string, recipient string, value float32, senderPublicKey *ecdsa.PublicKey,
	s *utils.Signature) bool {
	t := NewTransaction(sender, recipient, value)

	if sender == MINING_SENDER {
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	}
	if bc.VerifyTransactionSignature(senderPublicKey, s, t) {
		if bc.CalculateTotalAmount(sender) < value {
			log.Println("Error: Not enough balance in a wallet")
			return false
		}
		bc.transactionPool = append(bc.transactionPool, t)
		return true
	} else {
		log.Println("ERROR: Verify Transaction")
	}
	return false
}

func (bc *BlockChain) VerifyTransactionSignature(
	senderPublicKey *ecdsa.PublicKey, s *utils.Signature, t *Transaction) bool {
	m, _ := json.Marshal(t)
	h := sha256.Sum256([]byte(m))
	return ecdsa.Verify(senderPublicKey, h[:], s.R, s.S)
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	v := &struct {
		Sender    *string  `json:"sender_blockchain_address"`
		Recipient *string  `json:"recipient_blockchain_address"`
		Value     *float32 `json:"value"`
	}{
		Sender:    &t.senderBlockchainAddress,
		Recipient: &t.recipientBlockchainAddress,
		Value:     &t.value,
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return nil
}

type TransactionRequest struct {
	SenderBlockChainAddress   *string  `json:"sender_blockchain_address"`
	ReceiverBlockChainAddress *string  `json:"receiver_blockchain_address"`
	SenderPublicKey           *string  `json:"sender_public_key"`
	Value                     *float32 `json:"value"`
	Signature                 *string  `json:"signature"`
}

func (tr *TransactionRequest) Validate() bool {
	if tr.SenderBlockChainAddress == nil ||
		tr.ReceiverBlockChainAddress == nil ||
		tr.SenderPublicKey == nil ||
		tr.Value == nil ||
		tr.Signature == nil {
		return false
	}
	return true
}

type AmountResponse struct {
	Amount float32 `json:"amount"`
}

func (ar *AmountResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Amount float32 `json:"amount"`
	}{
		Amount: ar.Amount,
	})
}

/**
---------------------------------总结---------------------------------------
区块：每一个区块除了存储各种交易信息外，还存储上一个区块的hash,
区块链：将这些区块连接起来的链(定义两个数组,分别存每个区块的hash和每个区块的对象,hash的下标对应相应的区块)
*/
