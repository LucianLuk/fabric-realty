package service

import (
	"application/pkg/fabric"
	"encoding/json"
	"fmt"
	"time"
)

type CarDealerService struct{}

const CAR_DEALER_ORG = "org1" // 汽车经销商组织

// CreateCar 创建汽车信息
func (s *CarDealerService) CreateCar(id, model, vin, owner string) error {
	contract := fabric.GetContract(CAR_DEALER_ORG)
	now := time.Now().Format(time.RFC3339)
	// 注意：链码函数名也需要修改为 CreateCar
	_, err := contract.SubmitTransaction("CreateCar", id, model, vin, owner, now)
	if err != nil {
		return fmt.Errorf("创建汽车信息失败：%s", fabric.ExtractErrorMessage(err))
	}
	return nil
}

// QueryCar 查询汽车信息
func (s *CarDealerService) QueryCar(id string) (map[string]interface{}, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG)
	// 注意：链码函数名也需要修改为 QueryCar
	result, err := contract.EvaluateTransaction("QueryCar", id)
	if err != nil {
		return nil, fmt.Errorf("查询汽车信息失败：%s", fabric.ExtractErrorMessage(err))
	}

	var car map[string]interface{}
	if err := json.Unmarshal(result, &car); err != nil {
		return nil, fmt.Errorf("解析汽车数据失败：%v", err)
	}

	return car, nil
}

// QueryCarList 分页查询汽车列表
func (s *CarDealerService) QueryCarList(pageSize int32, bookmark string, status string) (map[string]interface{}, error) {
	contract := fabric.GetContract(CAR_DEALER_ORG)
	// 注意：链码函数名也需要修改为 QueryCarList
	result, err := contract.EvaluateTransaction("QueryCarList", fmt.Sprintf("%d", pageSize), bookmark, status)
	if err != nil {
		return nil, fmt.Errorf("查询汽车列表失败：%s", fabric.ExtractErrorMessage(err))
	}

	var queryResult map[string]interface{}
	if err := json.Unmarshal(result, &queryResult); err != nil {
		return nil, fmt.Errorf("解析查询结果失败：%v", err)
	}

	return queryResult, nil
}

// QueryBlockList 分页查询区块列表
func (s *CarDealerService) QueryBlockList(pageSize int, pageNum int) (*fabric.BlockQueryResult, error) {
	result, err := fabric.GetBlockListener().GetBlocksByOrg(CAR_DEALER_ORG, pageSize, pageNum)
	if err != nil {
		return nil, fmt.Errorf("查询区块列表失败：%v", err)
	}
	return result, nil
}
