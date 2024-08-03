package utils

import (
	"fmt"
	"math/big"
)

// 签名结构体
type Signature struct {
	R, S *big.Int
}

func (s *Signature) String() string {
	return fmt.Sprintf("%x%x", s.R, s.S)
}
