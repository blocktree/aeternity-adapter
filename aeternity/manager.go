package aeternity

import (
	"encoding/hex"
	"fmt"
	"github.com/aeternity/aepp-sdk-go/aeternity"
	"github.com/aeternity/aepp-sdk-go/swagguard/node/client/external"
	"github.com/aeternity/aepp-sdk-go/swagguard/node/models"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	rlp "github.com/randomshinichi/rlpae"
)

type WalletManager struct {
	openwallet.AssetsAdapterBase

	Api             *aeternity.Node                 // 节点客户端
	Config          *WalletConfig                   // 节点配置
	Decoder         openwallet.AddressDecoder       //地址编码器
	TxDecoder       openwallet.TransactionDecoder   //交易单编码器
	Log             *log.OWLogger                   //日志工具
	ContractDecoder openwallet.SmartContractDecoder //智能合约解析器
	Blockscanner    *AEBlockScanner                 //区块扫描器
	client          *Client                         //本地封装的http client
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol)
	wm.Blockscanner = NewAEBlockScanner(&wm)
	wm.Decoder = NewAddressDecoder(&wm)
	wm.TxDecoder = NewTransactionDecoder(&wm)
	wm.Log = log.NewOWLogger(wm.Symbol())
	//wm.ContractDecoder = NewContractDecoder(&wm)
	return &wm
}

//GetAccount
func (wm *WalletManager) GetAccount(address string) (*models.Account, error) {

	if wm.Api == nil {
		return nil, fmt.Errorf("aeternity API is not inited")
	}
	return wm.Api.GetAccount(address)
}

//GetAccountPendingTxCount
func (wm *WalletManager) GetAccountPendingTxCount(address string) (uint64, error) {

	//GetPendingAccountTransactionsByPubkey有bug
	p := external.NewGetPendingAccountTransactionsByPubkeyParams().WithPubkey(address)
	result, err := wm.Api.External.GetPendingAccountTransactionsByPubkey(p)
	if err != nil {
		return 0, err
	}

	count := len(result.Payload.Transactions)
	return uint64(count), nil

	//if wm.client == nil {
	//	return 0, fmt.Errorf("aeternity API is not inited")
	//}
	//
	//path := fmt.Sprintf("/accounts/%s/transactions/pending", address)
	//result, err := wm.client.Call(path, "GET", nil)
	//if err != nil {
	//	return 0, err
	//}

	//txs := result.Get("transactions")
	//if !txs.IsArray() {
	//	return 0, nil
	//}
	//
	//return uint64(len(txs.Array())), nil
}

// BroadcastTransaction recalculates the transaction hash and sends the transaction to the node.
func (wm *WalletManager) BroadcastTransaction(txHex string) (string, error) {
	txBytes, err := hex.DecodeString(txHex)
	if err != nil {
		return "", fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}
	signedEncodedTx := aeternity.Encode(aeternity.PrefixTransaction, txBytes)
	wm.Log.Debugf("signedEncodedTx: %s", signedEncodedTx)
	// calculate the hash of the decoded txRLP
	//rlpTxHashRaw := owcrypt.Hash(txBytes, 32, owcrypt.HASH_ALG_BLAKE2B)
	//// base58/64 encode the hash with the th_ prefix
	//signedEncodedTxHash := aeternity.Encode(aeternity.PrefixTransactionHash, rlpTxHashRaw)

	// send it to the network
	return postTransaction(wm.Api, signedEncodedTx)
}

// SignEncodeTx sign and encode a transaction
func SignEncodeTx(txRaw, sigRaw []byte) (string, error) {
	// encode the message using rlp
	rlpTxRaw, err := createSignedTransaction(txRaw, [][]byte{sigRaw})
	if err != nil {
		return "", err
	}
	// encode the rlp message with the prefix
	signedEncodedTx := aeternity.Encode(aeternity.PrefixTransaction, rlpTxRaw)
	return signedEncodedTx, err
}

func createSignedTransaction(txRaw []byte, signatures [][]byte) (rlpRawMsg []byte, err error) {
	// encode the message using rlp
	rlpRawMsg, err = buildRLPMessage(
		aeternity.ObjectTagSignedTransaction,
		1,
		signatures,
		txRaw,
	)
	return
}

func buildRLPMessage(tag uint, version uint, fields ...interface{}) (rlpRawMsg []byte, err error) {
	// create a message of the transaction and signature
	data := []interface{}{tag, version}
	data = append(data, fields...)
	// fmt.Printf("TX %#v\n\n", data)
	// encode the message using rlp
	rlpRawMsg, err = rlp.EncodeToBytes(data)
	// fmt.Printf("ENCODED %#v\n\n", data)
	return
}

// postTransaction post a transaction to the chain
func postTransaction(node *aeternity.Node, signedEncodedTx string) (string, error) {
	p := external.NewPostTransactionParams().WithBody(&models.Tx{
		Tx: &signedEncodedTx,
	})
	r, err := node.External.PostTransaction(p)
	if err != nil {
		return "", err
	}
	//if r.Payload.TxHash != models.EncodedHash(signedEncodedTxHash) {
	//	err = fmt.Errorf("Transaction hash mismatch, expected %s got %s", signedEncodedTxHash, r.Payload.TxHash)
	//}
	return string(*r.Payload.TxHash), nil
}
