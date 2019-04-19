package aeternity

import (
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
	r, err := wm.GetAccountPendingTxCount("ak_rozWtRmHh91aEu1Qo46wGSHtJfaGtbgRgPEezZvTmHtRu1fqe")
	if err != nil {
		log.Errorf("unexpected error: %v", err)
		return
	}
	log.Infof("%+v", r)
}