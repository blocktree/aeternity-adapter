module github.com/blocktree/aeternity-adapter

go 1.12

require (
	github.com/aeternity/aepp-sdk-go v0.0.0-20190409183752-7852cf075e5e
	github.com/asdine/storm v2.1.2+incompatible
	github.com/astaxie/beego v1.11.1
	github.com/blocktree/go-owcdrivers v1.0.3
	github.com/blocktree/go-owcrypt v1.0.1
	github.com/blocktree/openwallet v1.4.0
	github.com/imroc/req v0.2.3
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/tidwall/gjson v1.2.1
)

replace github.com/aeternity/aepp-sdk-go v0.0.0-20190409183752-7852cf075e5e => github.com/blocktree/aepp-sdk-go v0.0.0-20190418071916-ac99e1ac864a
