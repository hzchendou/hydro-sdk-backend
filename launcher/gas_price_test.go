package launcher

import (
	"fmt"
	"github.com/shopspring/decimal"
	"os"
	"testing"
)

const rpcURL = "https://data-seed-prebsc-1-s1.binance.org:8545"

func TestGasPriceInWei(t *testing.T) {
	os.Setenv("HSK_BLOCKCHAIN_RPC_URL", rpcURL)
	fallbackGasPrice := decimal.New(6, 9) // 5Gwei
	priceDecider := NewGasStationGasPriceDecider(fallbackGasPrice)
	fmt.Println(priceDecider.GasPriceInWei())
}
