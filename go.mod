module github.com/blocktree/aeternity-adapter

go 1.12

require (
	github.com/aeternity/aepp-sdk-go v1.0.2
	github.com/asdine/storm v2.1.2+incompatible
	github.com/astaxie/beego v1.11.1
	github.com/blocktree/go-owcdrivers v1.0.12
	github.com/blocktree/go-owcrypt v1.0.1
	github.com/blocktree/openwallet v1.5.5
	github.com/imroc/req v0.2.3
	github.com/randomshinichi/rlpae v0.0.0-20190813143754-207301e28aeb
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/tidwall/gjson v1.2.1
	github.com/tidwall/pretty v1.0.0 // indirect
)

replace github.com/aeternity/aepp-sdk-go v1.0.2 => github.com/aeternity/aepp-sdk-go/v4 v4.0.1

//replace github.com/blocktree/openwallet => ../../openwallet
