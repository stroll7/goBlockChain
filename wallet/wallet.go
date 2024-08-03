package wallet

import (
	"GoProject/utils"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
)

// 钱包结构体
type Wallet struct {
	privateKey        *ecdsa.PrivateKey //私钥
	publicKey         *ecdsa.PublicKey  //公钥
	blockChainAddress string            //节点/地址
}

// 创建钱包
func NewWallet() *Wallet {
	//1. 创建ECDSA私钥（32字节）公钥（64字节）
	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	w.privateKey = privateKey
	w.publicKey = &w.privateKey.PublicKey
	//2. 对32字节的公钥执行SHA-256哈希运算
	h2 := sha256.New()
	h2.Write(w.publicKey.X.Bytes())
	h2.Write(w.publicKey.Y.Bytes())
	digest2 := h2.Sum(nil) //为对象创建哈希值
	//3.对SHA-256（20字节）的结果执行RIPEMD-160哈希运算
	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)
	//4. 在 RIPEMD-160哈希前添加一个版本字节（例如，0x00 表示主网）。
	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3[:])
	//5. 对扩展的RIPEMD-160结果进行两次SHA-256哈希处理，以生成一个校验和。
	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)
	//6. 对上一个SHA-256哈希的结果执行SHA-256
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)
	//7. 如果第二个SHA-256哈希用于校验和，则取前4个字节
	chsum := digest6[:4]
	//8. 将7中的4个校验和字节添加到4中的扩展RIPEMD-160哈希的末尾（25个字节）
	dc8 := make([]byte, 25)
	copy(dc8[:21], vd4[:])   //复制 vd4 到 dc8 的前 21 个字节
	copy(dc8[21:], chsum[:]) //复制校验和到 dc8 的后 4 个字节
	//9. 将字节字符串的结果转换为base58编码
	address := base58.Encode(dc8)
	w.blockChainAddress = address
	return w
}

func (w *Wallet) Prinr() {
	fmt.Printf("privateKey:%x\n", w.privateKey)
	fmt.Printf("publicKey:%x\n", w.publicKey)
}

func (w *Wallet) PrivateKey() *ecdsa.PrivateKey {
	return w.privateKey
}

func (w *Wallet) PrivateKeyStr() string {
	return fmt.Sprintf("%x", w.privateKey.D.Bytes())
}

func (w *Wallet) PublicKey() *ecdsa.PublicKey {
	return w.publicKey
}

func (w *Wallet) PublicKeyStr() string {
	return fmt.Sprintf("%x%x", w.publicKey.X.Bytes(), w.publicKey.Y.Bytes())
}

func (w *Wallet) BlockChainAddress() string {
	return w.blockChainAddress
}

type Transaction struct {
	senderPrivateKey          *ecdsa.PrivateKey
	senderPublicKey           *ecdsa.PublicKey
	senderBlockChainAddress   string
	receiverBlockChainAddress string
	value                     float32
}

func NewTransaction(privateKey *ecdsa.PrivateKey, publicKey *ecdsa.PublicKey,
	senderAddr string, receiverAddr string, value float32) *Transaction {
	return &Transaction{privateKey, publicKey, senderAddr, receiverAddr, value}
}

func (t *Transaction) GenerateSignature() *utils.Signature {
	//序列哈
	m, _ := json.Marshal(t)
	//使用SHA-256对序列化后的交易数据进行哈希运算，得到交易的哈希值。
	h := sha256.Sum256([]byte(m))
	//使用椭圆曲线数字签名算法（ECDSA）和发送者的私钥 t.senderPrivateKey 对交易哈希值 h 进行签名。签名过程生成两个值 r 和 s。
	r, s, _ := ecdsa.Sign(rand.Reader, t.senderPrivateKey, h[:])
	return &utils.Signature{r, s}
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		SenderAddr   string
		ReceiverAddr string
		Value        float32
	}{
		SenderAddr:   t.senderBlockChainAddress,
		ReceiverAddr: t.receiverBlockChainAddress,
		Value:        t.value,
	})
}
