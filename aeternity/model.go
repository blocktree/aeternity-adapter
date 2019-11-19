package aeternity

import (
	"github.com/aeternity/aepp-sdk-go/swagguard/node/models"
	"github.com/blocktree/openwallet/openwallet"
	"math/big"
)

type AddrBalance struct {
	Address      string
	Balance      *big.Int
	TokenBalance *big.Int
}

type txFeeInfo struct {
	GasUsed  *big.Int
	GasPrice *big.Int
	Fee      *big.Int
}

func (f *txFeeInfo) CalcFee() error {
	fee := new(big.Int)
	fee.Mul(f.GasUsed, f.GasPrice)
	f.Fee = fee
	return nil
}

type Block struct {
	Hash              string
	Confirmations     uint64
	Merkleroot        string
	MicroBlocks       []string
	Previousblockhash string
	Height            uint64 `storm:"id"`
	Version           uint64
	Time              uint64
	Fork              bool
}

func NewBlock(generation *models.Generation) *Block {
	obj := &Block{}
	obj.Height = *generation.KeyBlock.Height
	obj.Hash = *generation.KeyBlock.Hash
	obj.Previousblockhash = *generation.KeyBlock.PrevKeyHash
	obj.Time = *generation.KeyBlock.Time
	obj.MicroBlocks = generation.MicroBlocks

	return obj
}

//BlockHeader 区块链头
func (b *Block) BlockHeader(symbol string) *openwallet.BlockHeader {

	obj := openwallet.BlockHeader{}
	//解析json
	obj.Hash = b.Hash
	//obj.Confirmations = b.Confirmations
	//obj.Merkleroot = b.TransactionMerkleRoot
	obj.Previousblockhash = b.Previousblockhash
	obj.Height = b.Height
	obj.Version = uint64(b.Version)
	obj.Time = b.Time
	obj.Symbol = symbol

	return &obj
}

type MicroBlock struct {
	Hash              string `storm:"id"`
	Height            uint64
}

func NewMicroBlock(height uint64, hash string) *MicroBlock {
	obj := &MicroBlock{}
	obj.Height = height
	obj.Hash = hash
	return obj
}

