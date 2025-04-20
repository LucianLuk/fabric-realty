package service

import (
	"application/pkg/fabric"
	"encoding/json"
	"fmt"
	"time"
)

type TradingPlatformService struct{}

const TRADE_ORG = "org3" // 交易平台组织

// CreateTransaction 生成交易
func (s *TradingPlatformService) CreateTransaction(txID, carID, seller, buyer string, price float64) error { // 修改 realEstateID 为 carID
	contract := fabric.GetContract(TRADE_ORG)
	now := time.Now().Format(time.RFC3339)
	// 注意：链码函数名 CreateTransaction 的参数也需要对应修改
	_, err := contract.SubmitTransaction("CreateTransaction", txID, carID, seller, buyer, fmt.Sprintf("%f", price), now)
	if err != nil {
		return fmt.Errorf("生成交易失败：%s", fabric.ExtractErrorMessage(err))
	}
	return nil
}

// QueryCar 查询汽车信息
func (s *TradingPlatformService) QueryCar(id string) (map[string]interface{}, error) { // 修改函数名和返回类型注释
	contract := fabric.GetContract(TRADE_ORG)
	// 注意：链码函数名也需要修改为 QueryCar
	result, err := contract.EvaluateTransaction("QueryCar", id)
	if err != nil {
		return nil, fmt.Errorf("查询汽车信息失败：%s", fabric.ExtractErrorMessage(err))
	}

	var car map[string]interface{} // 修改变量名
	if err := json.Unmarshal(result, &car); err != nil {
		return nil, fmt.Errorf("解析汽车数据失败：%v", err) // 修改错误信息
	}

	return car, nil // 修改返回值
}

// QueryTransaction 查询交易信息
func (s *TradingPlatformService) QueryTransaction(txID string) (map[string]interface{}, error) {
	contract := fabric.GetContract(TRADE_ORG)
	result, err := contract.EvaluateTransaction("QueryTransaction", txID)
	if err != nil {
		return nil, fmt.Errorf("查询交易信息失败：%s", fabric.ExtractErrorMessage(err))
	}

	var transaction map[string]interface{}
	if err := json.Unmarshal(result, &transaction); err != nil {
		return nil, fmt.Errorf("解析交易数据失败：%v", err)
	}

	return transaction, nil
}

// QueryTransactionList 分页查询交易列表
func (s *TradingPlatformService) QueryTransactionList(pageSize int32, bookmark string, status string) (map[string]interface{}, error) {
	contract := fabric.GetContract(TRADE_ORG)
	result, err := contract.EvaluateTransaction("QueryTransactionList", fmt.Sprintf("%d", pageSize), bookmark, status)
	if err != nil {
		return nil, fmt.Errorf("查询交易列表失败：%s", fabric.ExtractErrorMessage(err))
	}

	var queryResult map[string]interface{}
	if err := json.Unmarshal(result, &queryResult); err != nil {
		return nil, fmt.Errorf("解析查询结果失败：%v", err)
	}

	return queryResult, nil
}

// QueryBlockList 分页查询区块列表
func (s *TradingPlatformService) QueryBlockList(pageSize int, pageNum int) (*fabric.BlockQueryResult, error) {
	result, err := fabric.GetBlockListener().GetBlocksByOrg(TRADE_ORG, pageSize, pageNum)
	if err != nil {
		return nil, fmt.Errorf("查询区块列表失败：%v", err)
	}
	return result, nil
}
