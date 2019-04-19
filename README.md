# aeternity-adapter

本项目适配了openwallet.AssetsAdapter接口，给应用提供了底层的区块链协议支持。

## 如何测试

openwtester包下的测试用例已经集成了openwallet钱包体系，创建conf文件，新建AE.ini文件，编辑如下内容：

```ini

# RPC api url
serverAPI = "http://127.0.0.1:10007"
# AE networkID, default(mainnet) networkID = "ae_mainnet",
networkID = "ae_mainnet"
# fix fees for transaction
fixFees = "0.00002"

```