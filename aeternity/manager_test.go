package aeternity

import (
	"encoding/hex"
	"github.com/aeternity/aepp-sdk-go/aeternity"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/log"
	"path/filepath"
	"testing"
)

func testNewWalletManager() *WalletManager {
	wm := NewWalletManager()

	//读取配置
	absFile := filepath.Join("conf", "AE.ini")
	//log.Debug("absFile:", absFile)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return nil
	}
	wm.LoadAssetsConfig(c)
	return wm
}

func init() {

}

func TestWalletManager_GetAccount(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.GetAccount("ak_2Ju1M5wyNHBVRuiL3PFT4T6AaRkfK1qYk4GgkDBi2uNSXxL9tT")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_GetAccountPendingTxCount(t *testing.T) {
	wm := testNewWalletManager()
	r, err := wm.GetAccountPendingTxCount("ak_qcqXt6ySgRPvBkNwEpNMvaKWzrhPZsoBHLvgg68qg9vRht62y")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}

func TestWalletManager_BroadcastTransaction(t *testing.T) {
	wm := testNewWalletManager()
	txHex, _ := hex.DecodeString("f85b0c01a1016eeba7851c2ddb4dd4d47588f1e1a204f88d6e0eed3026caf3b8ec3e432f9d89a1016e6490ba9ffa3ed276048e23c52f09a7622e02111124e9c770d1a6ac11a723c6872386f26fc100008612309ce540008301dfcb0180")
	signature, _ := hex.DecodeString("6d3987133430a65b834052e3ca22282ca5b7741ca46d8a5c295338a8b890e340c42aaffa1f887d2016b25e2bb0b652c310a02f03c3f3c11abe9e87cc9a0e9f08")
	txBytes, err := createSignedTransaction(txHex, [][]byte{signature})
	if err != nil {
		t.Errorf("createSignedTransaction failed, unexpected error: %v", err)
		return
	}
	signedEncodedTx := aeternity.Encode(aeternity.PrefixTransaction, txBytes)
	log.Infof("signedEncodedTx: %s", signedEncodedTx)
	txid, err := wm.BroadcastTransaction(hex.EncodeToString(txBytes))
	if err != nil {
		t.Errorf("BroadcastTransaction failed, unexpected error: %v", err)
		return
	}
	log.Infof("txid: %s", txid)
}