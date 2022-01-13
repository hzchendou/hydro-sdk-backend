package launcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type GasPriceDecider interface {
	GasPriceInWei() decimal.Decimal
}

type StaticGasPriceDecider struct {
	PriceInWei decimal.Decimal
}

func (s StaticGasPriceDecider) GasPriceInWei() decimal.Decimal {
	return s.PriceInWei
}

type GasStationPriceDeciderWithFallback struct {
	FallbackGasPriceInWei decimal.Decimal
}

func (s GasStationPriceDeciderWithFallback) GasPriceInWei() decimal.Decimal {
	url := os.Getenv("HSK_BLOCKCHAIN_RPC_URL")
	data := make(map[string]interface{})
	//"jsonrpc":"2.0","method":"eth_gasPrice","id":1
    data["jsonrpc"] = "2.0"
    data["method"] = "eth_gasPrice"
    data["id"] = "1"
    jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	// resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return s.FallbackGasPriceInWei
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return s.FallbackGasPriceInWei
	}

	gasStationResp := GasStationRespBody{}
	err = json.Unmarshal(body, &gasStationResp)
	if err != nil || len(gasStationResp.Result) <= 0 {
		fmt.Println(err)
		return s.FallbackGasPriceInWei
	}
	return Hex2Dec(gasStationResp.Result, s.FallbackGasPriceInWei)
}

type GasStationRespBody struct {
	Result  string `json:"result"`
}

func NewStaticGasPriceDecider(gasPrice decimal.Decimal) GasPriceDecider {
	return StaticGasPriceDecider{
		PriceInWei: gasPrice,
	}
}

func NewGasStationGasPriceDecider(fallbackGasPrice decimal.Decimal) GasPriceDecider {
	return GasStationPriceDeciderWithFallback{
		FallbackGasPriceInWei: fallbackGasPrice,
	}
}

func Hex2Dec(val string, defaultValue decimal.Decimal) decimal.Decimal {
	n, err := strconv.ParseInt(val, 0, 64)
	if err != nil {
		fmt.Println(err)
		return defaultValue
	}
	return decimal.New(n, 0)
}

