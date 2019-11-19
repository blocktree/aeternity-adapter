package aeternity

import (
	"fmt"
	"github.com/aeternity/aepp-sdk-go/swagguard/node/models"
	"github.com/blocktree/openwallet/common"
	"github.com/blocktree/openwallet/openwallet"
	"math/big"
	"time"
)

const (
	//blockchainBucket = "blockchain" // blockchain dataset
	//periodOfTask      = 5 * time.Second // task interval
	maxExtractingSize = 10 // thread count
)

//AEBlockScanner AE block scanner
type AEBlockScanner struct {
	*openwallet.BlockScannerBase

	CurrentBlockHeight   uint64         //当前区块高度
	extractingCH         chan struct{}  //扫描工作令牌
	wm                   *WalletManager //钱包管理者
	RescanLastBlockCount uint64         //重扫上N个区块数量
}

//ExtractResult extract result
type ExtractTxResult struct {
	extractData map[string]*openwallet.TxExtractData
	TxID        string
}

//ExtractResult extract result
type ExtractResult struct {
	extractData  []*ExtractTxResult
	MicroBlockID string
	BlockHash    string
	BlockHeight  uint64
	BlockTime    int64
	Success      bool
}

//SaveResult result
type SaveResult struct {
	TxID        string
	BlockHeight uint64
	Success     bool
}

// NewEOSBlockScanner create a block scanner
func NewAEBlockScanner(wm *WalletManager) *AEBlockScanner {
	bs := AEBlockScanner{
		BlockScannerBase: openwallet.NewBlockScannerBase(),
	}

	bs.extractingCH = make(chan struct{}, maxExtractingSize)
	bs.wm = wm

	//AE区块链分为keyblock和microblock，keyblock记录记账权，不记录交易，microblock记录交易
	//所以顶部的keyblock后续出的区块是动态的，不能只扫描一次，要反复地扫描 顶部 - 3 个块，保证新增的microblock
	bs.RescanLastBlockCount = 3

	// set task
	bs.SetTask(bs.ScanBlockTask)

	return &bs
}

//GetBalanceByAddress 查询地址余额
func (bs *AEBlockScanner) GetBalanceByAddress(address ...string) ([]*openwallet.Balance, error) {

	addrBalanceArr := make([]*openwallet.Balance, 0)
	for _, a := range address {
		acc, err := bs.wm.GetAccount(a)
		if err == nil {
			balance := big.Int(acc.Balance)
			b := common.BigIntToDecimals(&balance, bs.wm.Decimal())
			obj := &openwallet.Balance{
				Symbol:           bs.wm.Symbol(),
				Address:          a,
				Balance:          b.String(),
				UnconfirmBalance: "0",
				ConfirmBalance:   "0",
			}

			addrBalanceArr = append(addrBalanceArr, obj)
			//return nil, err
		}

	}

	return addrBalanceArr, nil
}

//GetCurrentBlock 获取当前最新区块
func (bs *AEBlockScanner) GetCurrentBlock() (*Block, error) {

	generation, err := bs.wm.Api.External.GetCurrentGeneration(nil)
	if err != nil {
		return nil, err
	}

	block := NewBlock(generation.Payload)

	return block, nil
}

//GetBlockHeight 获取区块链高度
func (bs *AEBlockScanner) GetBlockHeight() (uint64, error) {

	height, err := bs.wm.Api.External.GetCurrentKeyBlockHeight(nil)
	if err != nil {
		return 0, err
	}
	return height.Payload.Height, nil
}

//GetCurrentBlockHeader 获取当前区块高度
func (bs *AEBlockScanner) GetCurrentBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		keyBlock *models.KeyBlock
		err      error
	)

	keyBlock, err = bs.wm.Api.GetCurrentKeyBlock()
	if err != nil {
		return nil, err
	}

	return &openwallet.BlockHeader{Height: *keyBlock.Height, Hash: *keyBlock.Hash}, nil
}

//GetTopBlock 获取顶部区块，可能是micro block 或 key block
func (bs *AEBlockScanner) GetTopBlock() (*models.KeyBlockOrMicroBlockHeader, error) {

	var (
		kb  *models.KeyBlockOrMicroBlockHeader
		err error
	)

	kb, err = bs.wm.Api.GetTopBlock()
	if err != nil {
		return nil, err
	}

	return kb, nil
}

//SetRescanBlockHeight 重置区块链扫描高度
func (bs *AEBlockScanner) SetRescanBlockHeight(height uint64) error {
	height = height - 1
	if height < 0 {
		return fmt.Errorf("block height to rescan must greater than 0.")
	}
	block, err := bs.GetBlockByHeight(height)
	if err != nil {
		return err
	}

	bs.SaveLocalBlockHead(height, block.Hash)

	return nil
}

func (bs *AEBlockScanner) GetBlockByHeight(height uint64) (*Block, error) {
	keyBlock, err := bs.wm.Api.GetGenerationByHeight(height)
	if err != nil {
		return nil, err
	}

	block := NewBlock(keyBlock)

	return block, nil
}

//GetScannedBlockHeader 获取当前扫描的区块头
func (bs *AEBlockScanner) GetScannedBlockHeader() (*openwallet.BlockHeader, error) {

	var (
		blockHeader *openwallet.BlockHeader
		blockHeight uint64 = 0
		hash        string
		err         error
	)

	blockHeight, hash, err = bs.GetLocalBlockHead()

	//如果本地没有记录，查询接口的高度
	if blockHeight == 0 {
		blockHeader, err = bs.GetCurrentBlockHeader()
		if err != nil {

			return nil, err
		}
		blockHeight = blockHeader.Height
		//就上一个区块链为当前区块
		blockHeight = blockHeight - 1

		block, err := bs.GetBlockByHeight(blockHeight)
		if err != nil {
			return nil, err
		}
		hash = block.Hash
	}

	return &openwallet.BlockHeader{Height: blockHeight, Hash: hash}, nil
}

//GetScannedBlockHeight 获取已扫区块高度
func (bs *AEBlockScanner) GetScannedBlockHeight() uint64 {
	localHeight, _, _ := bs.GetLocalBlockHead()
	return localHeight
}

//GetGlobalMaxBlockHeight 获取区块链全网最大高度
func (bs *AEBlockScanner) GetGlobalMaxBlockHeight() uint64 {

	generation, err := bs.wm.Api.External.GetCurrentGeneration(nil)
	if err != nil {
		return 0
	}

	return *generation.Payload.KeyBlock.Height
}

//GetTransactionsByMicroBlockHash
func (bs *AEBlockScanner) GetTransactionsByMicroBlockHash2(hash string) ([]*models.GenericSignedTx, error) {
	txs, err := bs.wm.Api.GetMicroBlockTransactionsByHash(hash)
	if err != nil {
		return nil, err
	}
	return txs.Transactions, nil
}

//GetTransactionsByMicroBlockHash
func (bs *AEBlockScanner) GetTransactionsByMicroBlockHash(hash string) ([]*models.GenericSignedTx, error) {

	if bs.wm.client == nil {
		return nil, fmt.Errorf("aeternity API is not inited")
	}

	path := fmt.Sprintf("/micro-blocks/hash/%s/transactions", hash)
	result, err := bs.wm.client.Call(path, "GET", nil)
	if err != nil {
		return nil, err
	}

	txs := result.Get("transactions")
	if !txs.IsArray() {
		return nil, nil
	}

	gtxArray := make([]*models.GenericSignedTx, 0)

	for _, tx := range txs.Array() {
		if tx.Get("tx.type").String() == "SpendTx" {
			gtx := &models.GenericSignedTx{}
			err := gtx.UnmarshalJSON([]byte(tx.Raw))
			if err != nil {
				return nil, err
			}
			gtxArray = append(gtxArray, gtx)
		}
	}


	return gtxArray, nil
}


//GetTransactionsByBlockHash
func (bs *AEBlockScanner) GetTransactionsByBlock(block *Block) ([]*models.GenericSignedTx, error) {

	transactions := make([]*models.GenericSignedTx, 0)
	for _, mb := range block.MicroBlocks {
		txs, err := bs.wm.Api.GetMicroBlockTransactionsByHash(string(mb))
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, txs.Transactions...)
	}
	return transactions, nil
}

//ScanBlockTask 扫描任务
func (bs *AEBlockScanner) ScanBlockTask() {

	//获取本地区块高度
	blockHeader, err := bs.GetScannedBlockHeader()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block height; unexpected error: %v", err)
		return
	}

	currentHeight := blockHeader.Height
	currentHash := blockHeader.Hash

	for {

		if !bs.Scanning {
			//区块扫描器已暂停，马上结束本次任务
			return
		}

		//获取最大高度
		maxHeight, err := bs.GetBlockHeight()
		if err != nil {
			//下一个高度找不到会报异常
			bs.wm.Log.Std.Info("block scanner can not get rpc-server block height; unexpected error: %v", err)
			break
		}

		//是否已到最新高度
		if currentHeight >= maxHeight {
			bs.wm.Log.Std.Info("block scanner has scanned full chain data. Current height: %d", maxHeight)
			break
		}

		//继续扫描下一个区块
		currentHeight = currentHeight + 1

		bs.wm.Log.Std.Info("block scanner scanning height: %d ...", currentHeight)

		//获取最大高度
		block, err := bs.GetBlockByHeight(currentHeight)
		if err != nil {
			//记录未扫区块
			unscanRecord := openwallet.NewUnscanRecord(currentHeight, "", err.Error(), bs.wm.Symbol())
			bs.SaveUnscanRecord(unscanRecord)
			bs.wm.Log.Std.Info("block height: %d extract failed.", currentHeight)
			return
		}

		hash := block.Hash

		isFork := false

		//判断hash是否上一区块的hash
		if currentHash != block.Previousblockhash {

			bs.wm.Log.Std.Info("block has been fork on height: %d.", currentHeight)
			bs.wm.Log.Std.Info("block height: %d local hash = %s ", currentHeight-1, currentHash)
			bs.wm.Log.Std.Info("block height: %d mainnet hash = %s ", currentHeight-1, block.Previousblockhash)

			bs.wm.Log.Std.Info("delete recharge records on block height: %d.", currentHeight-1)

			//查询本地分叉的区块
			forkBlock, _ := bs.GetLocalBlock(currentHeight - 1)

			//删除上一区块链的所有充值记录
			//bs.DeleteRechargesByHeight(currentHeight - 1)
			//删除上一区块链的未扫记录
			bs.DeleteUnscanRecord(currentHeight - 1)
			currentHeight = currentHeight - 2 //倒退2个区块重新扫描
			if currentHeight <= 0 {
				currentHeight = 1
			}

			localBlock, err := bs.GetLocalBlock(currentHeight)
			if err != nil {
				bs.wm.Log.Std.Warning("block scanner can not get local block; unexpected error: %v", err)

				//查找core钱包的RPC
				bs.wm.Log.Info("block scanner prev block height:", currentHeight)

				localBlock, err = bs.GetBlockByHeight(currentHeight)
				if err != nil {
					bs.wm.Log.Std.Error("block scanner can not get prev block; unexpected error: %v", err)
					break
				}

			}

			//重置当前区块的hash
			currentHash = localBlock.Hash

			bs.wm.Log.Std.Info("rescan block on height: %d, hash: %s .", currentHeight, currentHash)

			//重新记录一个新扫描起点
			bs.SaveLocalBlockHead(localBlock.Height, localBlock.Hash)

			isFork = true

			if forkBlock != nil {

				//通知分叉区块给观测者，异步处理
				bs.newBlockNotify(forkBlock, isFork)
			}

		} else {

			err = bs.BatchExtractTransaction(block)
			if err != nil {
				bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			}

			//重置当前区块的hash
			currentHash = hash

			//保存本地新高度
			bs.SaveLocalBlockHead(currentHeight, currentHash)
			bs.SaveLocalBlock(block)

			isFork = false

			//通知新区块给观测者，异步处理
			bs.newBlockNotify(block, isFork)
		}

	}

	//重扫前N个块，为保证记录找到
	for i := currentHeight - bs.RescanLastBlockCount; i < currentHeight; i++ {
		bs.scanBlock(i + 1)
	}

	//重扫失败区块
	bs.RescanFailedRecord()

}

//ScanBlock 扫描指定高度区块
func (bs *AEBlockScanner) ScanBlock(height uint64) error {

	block, err := bs.scanBlock(height)
	if err != nil {
		return err
	}

	//通知新区块给观测者，异步处理
	bs.newBlockNotify(block, false)

	return nil
}

func (bs *AEBlockScanner) scanBlock(height uint64) (*Block, error) {

	block, err := bs.GetBlockByHeight(height)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
		//记录未扫区块
		//unscanRecord := NewUnscanRecord(height, "", err.Error())
		//bs.SaveUnscanRecord(unscanRecord)
		bs.wm.Log.Std.Info("block height: %d extract failed.", height)
		return nil, err
	}

	bs.wm.Log.Std.Info("block scanner rescanning height: %d ...", block.Height)

	err = bs.BatchExtractTransaction(block)
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
		return nil, err
	}
	//通知新区块给观测者，异步处理
	//bs.newBlockNotify(block, false)

	return block, nil
}

//rescanFailedRecord 重扫失败记录
func (bs *AEBlockScanner) RescanFailedRecord() {

	var (
		blockMap = make(map[uint64][]models.EncodedHash)
	)

	list, err := bs.GetUnscanRecords()
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get rescan data; unexpected error: %v", err)
	}

	//组合成批处理
	for _, r := range list {

		if _, exist := blockMap[r.BlockHeight]; !exist {
			blockMap[r.BlockHeight] = make([]models.EncodedHash, 0)
		}

		//if len(r.MicroBlockID) > 0 {
		//	arr := blockMap[r.BlockHeight]
		//	arr = append(arr, models.EncodedHash(r.MicroBlockID))
		//
		//	blockMap[r.BlockHeight] = arr
		//}
	}

	for height, _ := range blockMap {

		if height == 0 {
			continue
		}

		bs.wm.Log.Std.Info("block scanner rescanning height: %d ...", height)

		block, err := bs.GetBlockByHeight(height)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not get new block data; unexpected error: %v", err)
			continue
		}

		err = bs.BatchExtractTransaction(block)
		if err != nil {
			bs.wm.Log.Std.Info("block scanner can not extractRechargeRecords; unexpected error: %v", err)
			continue
		}

		//删除未扫记录
		bs.DeleteUnscanRecord(height)
	}

}

//newBlockNotify 获得新区块后，通知给观测者
func (bs *AEBlockScanner) newBlockNotify(block *Block, isFork bool) {
	header := block.BlockHeader(bs.wm.Symbol())
	header.Fork = isFork
	bs.NewBlockNotify(header)
}

//BatchExtractTransaction 批量提取交易单
//bitcoin 1M的区块链可以容纳3000笔交易，批量多线程处理，速度更快
func (bs *AEBlockScanner) BatchExtractTransaction(block *Block) error {

	var (
		quit       = make(chan struct{})
		done       = 0 //完成标记
		failed     = 0
		shouldDone = len(block.MicroBlocks) //需要完成的总数
	)

	if len(block.MicroBlocks) == 0 {
		return nil
	}

	//生产通道
	producer := make(chan ExtractResult)
	defer close(producer)

	//消费通道
	worker := make(chan ExtractResult)
	defer close(worker)

	//保存工作
	saveWork := func(height uint64, result chan ExtractResult) {
		//回收创建的地址
		for gets := range result {

			if gets.Success {

				notifyErr := bs.newExtractDataNotify(height, gets.extractData)
				//saveErr := bs.SaveRechargeToWalletDB(height, gets.Recharges)
				if notifyErr != nil {
					failed++ //标记保存失败数
					bs.wm.Log.Std.Info("newExtractDataNotify unexpected error: %v", notifyErr)
				}

			} else {
				//记录未扫区块
				unscanRecord := openwallet.NewUnscanRecord(height, "", "", bs.wm.Symbol())
				bs.SaveUnscanRecord(unscanRecord)
				bs.wm.Log.Std.Info("block height: %d extract failed.", height)
				failed++ //标记保存失败数
			}
			//累计完成的线程数
			done++
			if done == shouldDone {
				//bs.wm.Log.Std.Info("done = %d, shouldDone = %d ", done, len(txs))
				close(quit) //关闭通道，等于给通道传入nil
			}
		}
	}

	//提取工作
	extractWork := func(eBlock *Block, eProducer chan ExtractResult) {
		for _, mid := range eBlock.MicroBlocks {
			bs.extractingCH <- struct{}{}
			//shouldDone++
			go func(mBlock *Block, mMid string, end chan struct{}, mProducer chan<- ExtractResult) {

				//导出提出的交易
				mProducer <- bs.ExtractMicroBlock(mBlock, mMid, bs.ScanTargetFunc)
				//释放
				<-end

			}(eBlock, mid, bs.extractingCH, eProducer)
		}
	}

	/*	开启导出的线程	*/

	//独立线程运行消费
	go saveWork(block.Height, worker)

	//独立线程运行生产
	go extractWork(block, producer)

	//以下使用生产消费模式
	bs.extractRuntime(producer, worker, quit)

	if failed > 0 {
		return fmt.Errorf("block scanner saveWork failed")
	} else {
		return nil
	}
}

//extractRuntime 提取运行时
func (bs *AEBlockScanner) extractRuntime(producer chan ExtractResult, worker chan ExtractResult, quit chan struct{}) {

	var (
		values = make([]ExtractResult, 0)
	)

	for {

		var activeWorker chan<- ExtractResult
		var activeValue ExtractResult

		//当数据队列有数据时，释放顶部，传输给消费者
		if len(values) > 0 {
			activeWorker = worker
			activeValue = values[0]

		}

		select {

		//生成者不断生成数据，插入到数据队列尾部
		case pa := <-producer:
			values = append(values, pa)
		case <-quit:
			//退出
			//bs.wm.Log.Std.Info("block scanner have been scanned!")
			return
		case activeWorker <- activeValue:
			//wm.Log.Std.Info("Get %d", len(activeValue))
			values = values[1:]
		}
	}

}

//ExtractMicroBlock
func (bs *AEBlockScanner) ExtractMicroBlock(block *Block, microBlockID string, scanTargetFunc openwallet.BlockScanTargetFunc) ExtractResult {

	var (
		result = ExtractResult{
			Success:      true,
			BlockHeight:  block.Height,
			MicroBlockID: string(microBlockID),
			extractData:  make([]*ExtractTxResult, 0),
		}
	)

	//查询micro block下的所有交易单
	txs, err := bs.GetTransactionsByMicroBlockHash(string(microBlockID))
	if err != nil {
		bs.wm.Log.Std.Info("block scanner can not get transaction data; unexpected error: %v", err)
		result.Success = false
		return result
	}

	for _, tx := range txs {
		txRes, txErr := bs.ExtractTransaction(block, result.MicroBlockID, tx, scanTargetFunc)
		if txErr != nil {
			bs.wm.Log.Std.Info("block scanner can not extract transaction data; unexpected error: %v", err)
			result.Success = false
			return result
		}

		result.extractData = append(result.extractData, txRes)
	}

	return result
}

//ExtractTransaction 提取交易单
func (bs *AEBlockScanner) ExtractTransaction(block *Block, microBlockID string, trx *models.GenericSignedTx, scanTargetFunc openwallet.BlockScanTargetFunc) (*ExtractTxResult, error) {

	var (
		txID   = *trx.Hash
		result = &ExtractTxResult{
			TxID:        txID,
			extractData: make(map[string]*openwallet.TxExtractData),
		}
		createAt = time.Now().Unix()
		trxType  = trx.Tx().Type()
		decimals = bs.wm.Decimal()
	)

	switch trxType {
	case "SpendTx":
		spendTxJSON, ok := trx.Tx().(*models.SpendTxJSON)
		if !ok {
			return nil, fmt.Errorf("the tx can not convert models.SpendTxJSON")
		}
		bigHeight := big.Int(trx.BlockHeight)
		bigAmount := big.Int(spendTxJSON.Amount)
		bigFee := big.Int(spendTxJSON.Fee)
		amount := common.BigIntToDecimals(&bigAmount, decimals).String()
		fees := common.BigIntToDecimals(&bigFee, decimals).String()
		from := *spendTxJSON.SenderID
		to := *spendTxJSON.RecipientID

		sourceKey, ok := scanTargetFunc(
			openwallet.ScanTarget{
				Address:          from,
				BalanceModelType: openwallet.BalanceModelTypeAddress,
			})
		if ok {
			input := openwallet.TxInput{}
			input.TxID = txID
			input.Address = *spendTxJSON.SenderID
			input.Amount = amount
			input.Coin = openwallet.Coin{
				Symbol:     bs.wm.Symbol(),
				IsContract: false,
			}
			input.Index = 0
			input.Sid = openwallet.GenTxInputSID(txID, bs.wm.Symbol(), "", uint64(0))
			//input.CreateAt = createAt
			input.BlockHeight = bigHeight.Uint64()
			//input.BlockHash = string(trx.BlockHash)
			input.BlockHash = block.Hash //TODO: 先记录keyblock的hash方便上层计算确认次数，以后做扩展
			ed := result.extractData[sourceKey]
			if ed == nil {
				ed = openwallet.NewBlockExtractData()
				result.extractData[sourceKey] = ed
			}

			ed.TxInputs = append(ed.TxInputs, &input)

			//手续费也作为一个输出
			tmp := *&input
			feeCharge := &tmp
			feeCharge.Amount = fees
			ed.TxInputs = append(ed.TxInputs, feeCharge)
		}

		sourceKey2, ok2 := scanTargetFunc(
			openwallet.ScanTarget{
				Address:          *spendTxJSON.RecipientID,
				BalanceModelType: openwallet.BalanceModelTypeAddress,
			})
		if ok2 {
			output := openwallet.TxOutPut{}
			output.TxID = txID
			output.Address = to
			output.Amount = amount
			output.Coin = openwallet.Coin{
				Symbol:     bs.wm.Symbol(),
				IsContract: false,
			}
			output.Index = 0
			output.Sid = openwallet.GenTxOutPutSID(txID, bs.wm.Symbol(), "", 0)
			output.CreateAt = createAt

			output.BlockHeight = bigHeight.Uint64()
			//output.BlockHash = string(trx.BlockHash)
			output.BlockHash = block.Hash //TODO: 先记录keyblock的hash方便上层计算确认次数，以后做扩展
			ed := result.extractData[sourceKey2]
			if ed == nil {
				ed = openwallet.NewBlockExtractData()
				result.extractData[sourceKey2] = ed
			}

			ed.TxOutputs = append(ed.TxOutputs, &output)
		}

		for _, extractData := range result.extractData {
			status := "1"
			reason := ""
			tx := &openwallet.Transaction{
				From:   []string{from + ":" + amount},
				To:     []string{to + ":" + amount},
				Amount: amount,
				Fees:   fees,
				Coin: openwallet.Coin{
					Symbol:     bs.wm.Symbol(),
					IsContract: false,
				},
				//BlockHash:   string(trx.BlockHash),
				BlockHash:   block.Hash, //TODO: 先记录keyblock的hash方便上层计算确认次数，以后做扩展
				BlockHeight: bigHeight.Uint64(),
				TxID:        txID,
				Decimal:     decimals,
				Status:      status,
				Reason:      reason,
				//SubmitTime:  int64(block.Time),
				ConfirmTime: int64(block.Time),
			}
			wxID := openwallet.GenTransactionWxID(tx)
			tx.WxID = wxID
			extractData.Transaction = tx
		}
	default:
		return result, nil
	}

	return result, nil

}

//newExtractDataNotify 发送通知
func (bs *AEBlockScanner) newExtractDataNotify(height uint64, extractTxResult []*ExtractTxResult) error {

	for o, _ := range bs.Observers {
		for _, txResult := range extractTxResult {
			for key, data := range txResult.extractData {
				err := o.BlockExtractDataNotify(key, data)
				if err != nil {
					bs.wm.Log.Error("BlockExtractDataNotify unexpected error:", err)
					//记录未扫区块
					unscanRecord := openwallet.NewUnscanRecord(height, "", "ExtractData Notify failed.", bs.wm.Symbol())
					err = bs.SaveUnscanRecord(unscanRecord)
					if err != nil {
						bs.wm.Log.Std.Error("block height: %d, save unscan record failed. unexpected error: %v", height, err.Error())
					}

				}
			}
		}
	}

	return nil
}

//ExtractTransactionData
func (bs *AEBlockScanner) ExtractTransactionData(txid string, scanAddressFunc openwallet.BlockScanTargetFunc) (map[string][]*openwallet.TxExtractData, error) {
	tx, err := bs.wm.Api.GetTransactionByHash(txid)
	if err != nil {
		return nil, err
	}
	bigHeight := big.Int(tx.BlockHeight)
	block, err := bs.GetBlockByHeight(bigHeight.Uint64())
	if err != nil {
		return nil, err
	}
	result, err := bs.ExtractTransaction(block, *tx.BlockHash, tx, scanAddressFunc)
	if err != nil {
		return nil, err
	}
	extData := make(map[string][]*openwallet.TxExtractData)
	for key, data := range result.extractData {
		txs := extData[key]
		if txs == nil {
			txs = make([]*openwallet.TxExtractData, 0)
		}
		txs = append(txs, data)
		extData[key] = txs
	}
	return extData, nil
}


//SupportBlockchainDAI 支持外部设置区块链数据访问接口
//@optional
func (bs *AEBlockScanner) SupportBlockchainDAI() bool {
	return true
}