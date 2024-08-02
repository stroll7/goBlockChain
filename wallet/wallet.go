package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/base58"
	"golang.org/x/crypto/ripemd160"
	"math/big"
)

type Wallet struct {
	privateKey        *ecdsa.PrivateKey
	publicKey         *ecdsa.PublicKey
	blockChainAddress string
}

func NewWallet() *Wallet {
	//1. Creat ecdsa pr
	w := new(Wallet)
	privateKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	w.privateKey = privateKey
	w.publicKey = &w.privateKey.PublicKey
	h2 := sha256.New()
	h2.Write(w.publicKey.X.Bytes())
	h2.Write(w.publicKey.Y.Bytes())
	digest2 := h2.Sum(nil)
	h3 := ripemd160.New()
	h3.Write(digest2)
	digest3 := h3.Sum(nil)
	vd4 := make([]byte, 21)
	vd4[0] = 0x00
	copy(vd4[1:], digest3[:])
	h5 := sha256.New()
	h5.Write(vd4)
	digest5 := h5.Sum(nil)
	h6 := sha256.New()
	h6.Write(digest5)
	digest6 := h6.Sum(nil)
	chsum := digest6[:4]
	dc8 := make([]byte, 25)
	copy(dc8[20:], vd4[:])
	copy(dc8[21:], chsum[:])
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
	senderPrivateKey          *ecdsa.PublicKey
	senderPublicKey           *ecdsa.PublicKey
	senderBlockChainAddress   string
	receiverBlockChainAddress string
	value                     float32
}

func NewTransaction(privateKey *ecdsa.PublicKey, publicKey *ecdsa.PublicKey,
	senderAddr string, receiverAddr string, value float32) *Transaction {
	return &Transaction{privateKey, publicKey, senderAddr, receiverAddr, value}
}

func (t *Transaction) GenerateSignature() *Signature {
	m, _ := json.Marshal(t)
}

func (t *Transaction) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		senderAddr   string
		ReceiverAddr string
		value        float32
	}{
		senderAddr: t.senderBlockChainAddres
	})
}

type Signature struct {
	R, S *big.Int
}
