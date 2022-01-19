package launcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"strconv"
)

type GasPriceDecider interface {
	GasPriceInWei() decimal.Decimal
	EvaluateGasUsed(from string, to string, d string) *big.Int
}

type StaticGasPriceDecider struct {
	PriceInWei decimal.Decimal
}

func (s StaticGasPriceDecider) GasPriceInWei() decimal.Decimal {
	return s.PriceInWei
}

func (s StaticGasPriceDecider) EvaluateGasUsed(from string, to string, d string) *big.Int {
	return big.NewInt(-1)
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


func (s GasStationPriceDeciderWithFallback) EvaluateGasUsed(from string, to string, d string) *big.Int {
	url := os.Getenv("HSK_BLOCKCHAIN_RPC_URL")
	data := make(map[string]interface{})
	data["jsonrpc"] = "2.0"
	data["method"] = "eth_estimateGas"
	params := make([]map[string]interface{}, 1)
	params[0] = make(map[string]interface{})
	params[0]["from"] = from
	params[0]["to"] = to
	params[0]["data"] = d
	data["params"] =  params
	data["id"] = "1"
	jsonData, _ := json.Marshal(data)
	resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
	// resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
		return big.NewInt(-1)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return big.NewInt(-1)
	}

	gasStationResp := GasStationRespBody{}
	err = json.Unmarshal(body, &gasStationResp)
	if err != nil || len(gasStationResp.Result) <= 0 {
		fmt.Println(err)
		return big.NewInt(-1)
	}
	return Hex2Int(gasStationResp.Result, big.NewInt(-1))
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

func Hex2Int(val string, defaultValue *big.Int) *big.Int {
	n, err := strconv.ParseInt(val, 0, 64)
	if err != nil {
		fmt.Println(err)
		return defaultValue
	}
	return big.NewInt(n)
}

