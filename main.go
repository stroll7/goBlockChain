package main

import (
	"GoProject/wallet"
	"fmt"
	"log"
)

func init() {
	log.SetPrefix("BlockChain: ")
}

type Wallet struct {
}

func main() {
	//创建钱包
	w := wallet.NewWallet()
	fmt.Println("私钥：", w.PrivateKey())
	fmt.Println("私钥：", w.PublicKey())

	//加密
	fmt.Println("加密私钥：", w.PrivateKeyStr())
	fmt.Println("加密私钥：", w.PublicKeyStr())

	fmt.Println("缩减公钥", w.BlockChainAddress())

	/*//初始化区块链
	myBlockChainAddress := "my_blockChain_address"
	blockChain := NewBlockChain(myBlockChainAddress)
	//打印
	blockChain.Print()

	//为初始的区块添加一条事务
	blockChain.AddTransaction("A", "B", 1.0)
	blockChain.Mining()
	blockChain.Print()

	blockChain.AddTransaction("C", "D", 2.0)
	blockChain.AddTransaction("E", "F", 3.0)
	blockChain.Mining()
	blockChain.Print()

	fmt.Printf("my%.1f\n", blockChain.CalculateTotalAmount("my_blockChain_address"))
	fmt.Printf("C %.1f\n", blockChain.CalculateTotalAmount("C"))
	fmt.Printf("D %.1f\n", blockChain.CalculateTotalAmount("D"))*/
}
