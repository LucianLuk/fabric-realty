package service

import (
	"application/pkg/fabric"
	"encoding/json"
	"fmt"
	"time"
)

type VehicleAgencyService struct{}

const VEHICLE_ORG = "org1" // 车辆管理机构组织

// CreateVehicle 创建车辆信息
func (s *VehicleAgencyService) CreateVehicle(id, model string, year int, brand string, mileage float64, condition string, owner string) error {
	contract := fabric.GetContract(VEHICLE_ORG)
	now := time.Now().Format(time.RFC3339)
	_, err := contract.SubmitTransaction("CreateVehicle", id, model, fmt.Sprintf("%d", year), brand, fmt.Sprintf("%f", mileage), condition, owner, now)
	if err != nil {
		return fmt.Errorf("创建车辆信息失败：%s", fabric.ExtractErrorMessage(err))
	}
	return nil
}

// QueryVehicle 查询车辆信息
func (s *VehicleAgencyService) QueryVehicle(id string) (map[string]interface{}, error) {
	contract := fabric.GetContract(VEHICLE_ORG)
	result, err := contract.EvaluateTransaction("QueryVehicle", id)
	if err != nil {
		return nil, fmt.Errorf("查询车辆信息失败：%s", fabric.ExtractErrorMessage(err))
	}

	var vehicle map[string]interface{}
	if err := json.Unmarshal(result, &vehicle); err != nil {
		return nil, fmt.Errorf("解析车辆数据失败：%v", err)
	}

	return vehicle, nil
}

// QueryVehicleList 分页查询车辆列表
func (s *VehicleAgencyService) QueryVehicleList(pageSize int32, bookmark string, status string) (map[string]interface{}, error) {
	contract := fabric.GetContract(VEHICLE_ORG)
	result, err := contract.EvaluateTransaction("QueryVehicleList", fmt.Sprintf("%d", pageSize), bookmark, status)
	if err != nil {
		return nil, fmt.Errorf("查询车辆列表失败：%s", fabric.ExtractErrorMessage(err))
	}

	var queryResult map[string]interface{}
	if err := json.Unmarshal(result, &queryResult); err != nil {
		return nil, fmt.Errorf("解析查询结果失败：%v", err)
	}

	return queryResult, nil
}

// QueryBlockList 分页查询区块列表
func (s *VehicleAgencyService) QueryBlockList(pageSize int, pageNum int) (*fabric.BlockQueryResult, error) {
	result, err := fabric.GetBlockListener().GetBlocksByOrg(VEHICLE_ORG, pageSize, pageNum)
	if err != nil {
		return nil, fmt.Errorf("查询区块列表失败：%v", err)
	}
	return result, nil
}
