package aeternity

import (
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"testing"
)

func TestAEBlockScanner_GetCurrentBlockHeader(t *testing.T) {
	wm := testNewWalletManager()
	header, err := wm.GetBlockScanner().GetCurrentBlockHeader()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("header: %v", header)
}


func TestGetBlockHeight(t *testing.T) {
	wm := testNewWalletManager()
	height, err := wm.Api.APIGetHeight()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("height: %v", height)
}

func TestAEBlockScanner_GetBlockByHeight(t *testing.T) {
	wm := testNewWalletManager()
	block, err := wm.Blockscanner.GetBlockByHeight(67213)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("block: %v", block)
}

func TestAEBlockScanner_GetTransactionsByMicroBlockHash(t *testing.T) {
	wm := testNewWalletManager()
	txs, err := wm.Blockscanner.GetTransactionsByMicroBlockHash("mh_MxzxMdhrUAFm7cDU3JoxaB1i6wvNepZ15S3uQWgBqbbN4ZJhV")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("txs: %v", txs)
}

func TestGetCurrentGeneration(t *testing.T) {
	wm := testNewWalletManager()
	generation, err := wm.Api.Node.External.GetCurrentGeneration(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("generation: +%v", generation)
}

func TestAEBlockScanner_ExtractTransactionData(t *testing.T) {

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanTargetFunc := func(target openwallet.ScanTarget) (string, bool) {
		if target.Address == "	ak_qcqXt6ySgRPvBkNwEpNMvaKWzrhPZsoBHLvgg68qg9vRht62y" {
			return "sender", true
		} else if target.Address == "ak_mPXUBSsSCJgfu3yz2i2AiVTtLA2TzMyMJL5e6X7shM9Qa246t" {
			return "recipient", true
		}
		return "", false
	}

	wm := testNewWalletManager()
	txs, err := wm.Blockscanner.ExtractTransactionData("th_KJntgEFCoaQoaH3ycBJ58qoqvKa1Zp9KzcwDNVpP9NzYEKpbZ", scanTargetFunc)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("txs: %v", txs)
}

func TestAEBlockScanner_GetTopBlock(t *testing.T) {
	wm := testNewWalletManager()
	block, err := wm.Blockscanner.GetTopBlock()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("key block: %+v", block.KeyBlock)
	log.Infof("micro block: %+v", block.MicroBlock)
	log.Infof("height: %d", *block.MicroBlock.Height)
}