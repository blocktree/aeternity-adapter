package aeternity

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestAddressDecoder_PublicKeyToAddress(t *testing.T) {
	pub, _ := hex.DecodeString("4b5acf6b45652ee28cbe6cf2747b7971f3c00e9867d15e5b98155811ad66d4ea")
	decoder := AddressDecoder{}
	addr, err := decoder.PublicKeyToAddress(pub, false)
	if err != nil {
		t.Errorf("PublicKeyToAddress error: %v", err)
		return
	}
	fmt.Println(addr)
}