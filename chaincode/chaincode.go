package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// SmartContract 提供二手汽车交易的功能
type SmartContract struct {
	contractapi.Contract
}

// 文档类型常量（用于创建复合键）
const (
	VEHICLE     = "VE" // 车辆信息
	TRANSACTION = "TX" // 交易信息
)

// VehicleStatus 车辆状态
type VehicleStatus string

const (
	NORMAL         VehicleStatus = "NORMAL"         // 正常
	IN_TRANSACTION VehicleStatus = "IN_TRANSACTION" // 交易中
)

// TransactionStatus 交易状态
type TransactionStatus string

const (
	PENDING   TransactionStatus = "PENDING"   // 待付款
	COMPLETED TransactionStatus = "COMPLETED" // 已完成
)

// Vehicle 车辆信息
type Vehicle struct {
	ID           string        `json:"id"`           // 车辆ID (VIN)
	Model        string        `json:"model"`        // 车型
	Year         int           `json:"year"`         // 年份
	Brand        string        `json:"brand"`        // 品牌
	Mileage      float64       `json:"mileage"`      // 里程数
	Condition    string        `json:"condition"`    // 车况
	CurrentOwner string        `json:"currentOwner"` // 当前所有者
	Status       VehicleStatus `json:"status"`       // 状态
	CreateTime   time.Time     `json:"createTime"`   // 创建时间
	UpdateTime   time.Time     `json:"updateTime"`   // 更新时间
}

// Transaction 交易信息
type Transaction struct {
	ID         string            `json:"id"`         // 交易ID
	VehicleID  string            `json:"vehicleId"`  // 车辆ID
	Seller     string            `json:"seller"`     // 卖家
	Buyer      string            `json:"buyer"`      // 买家
	Price      float64           `json:"price"`      // 成交价格
	Status     TransactionStatus `json:"status"`     // 状态
	CreateTime time.Time         `json:"createTime"` // 创建时间
	UpdateTime time.Time         `json:"updateTime"` // 更新时间
}

// QueryResult 分页查询结果
type QueryResult struct {
	Records             []interface{} `json:"records"`             // 记录列表
	RecordsCount        int32         `json:"recordsCount"`        // 本次返回的记录数
	Bookmark            string        `json:"bookmark"`            // 书签，用于下一页查询
	FetchedRecordsCount int32         `json:"fetchedRecordsCount"` // 总共获取的记录数
}

// 组织 MSP ID 常量
const (
	VEHICLE_ORG_MSPID = "Org1MSP" // 车辆管理机构组织 MSP ID
	BANK_ORG_MSPID    = "Org2MSP" // 银行组织 MSP ID
	TRADE_ORG_MSPID   = "Org3MSP" // 交易平台组织 MSP ID
)

// 通用方法: 获取客户端身份信息
func (s *SmartContract) getClientIdentityMSPID(ctx contractapi.TransactionContextInterface) (string, error) {
	clientID, err := cid.New(ctx.GetStub())
	if err != nil {
		return "", fmt.Errorf("获取客户端身份信息失败：%v", err)
	}
	return clientID.GetMSPID()
}

// 通用方法：创建和获取复合键
func (s *SmartContract) getCompositeKey(ctx contractapi.TransactionContextInterface, objectType string, attributes []string) (string, error) {
	key, err := ctx.GetStub().CreateCompositeKey(objectType, attributes)
	if err != nil {
		return "", fmt.Errorf("创建复合键失败：%v", err)
	}
	return key, nil
}

// 通用方法：获取状态
func (s *SmartContract) getState(ctx contractapi.TransactionContextInterface, key string, value interface{}) error {
	bytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return fmt.Errorf("读取状态失败：%v", err)
	}
	if bytes == nil {
		return fmt.Errorf("键 %s 不存在", key)
	}

	err = json.Unmarshal(bytes, value)
	if err != nil {
		return fmt.Errorf("解析数据失败：%v", err)
	}
	return nil
}

// 通用方法：保存状态
func (s *SmartContract) putState(ctx contractapi.TransactionContextInterface, key string, value interface{}) error {
	bytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("序列化数据失败：%v", err)
	}

	err = ctx.GetStub().PutState(key, bytes)
	if err != nil {
		return fmt.Errorf("保存状态失败：%v", err)
	}
	return nil
}

// CreateVehicle 创建车辆信息（仅车辆管理机构组织可以调用）
func (s *SmartContract) CreateVehicle(ctx contractapi.TransactionContextInterface, id string, model string, year int, brand string, mileage float64, condition string, owner string, createTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是车辆管理机构组织的成员
	if clientMSPID != VEHICLE_ORG_MSPID {
		return fmt.Errorf("只有车辆管理机构组织成员才能创建车辆信息")
	}

	// 参数验证
	if len(id) == 0 {
		return fmt.Errorf("车辆ID不能为空")
	}
	if len(model) == 0 {
		return fmt.Errorf("车型不能为空")
	}
	if year <= 0 {
		return fmt.Errorf("年份必须大于0")
	}
	if len(brand) == 0 {
		return fmt.Errorf("品牌不能为空")
	}
	if mileage < 0 {
		return fmt.Errorf("里程数必须大于等于0")
	}
	if len(condition) == 0 {
		return fmt.Errorf("车况不能为空")
	}
	if len(owner) == 0 {
		return fmt.Errorf("所有者不能为空")
	}

	// 检查车辆是否已存在（检查所有可能的状态）
	for _, status := range []VehicleStatus{NORMAL, IN_TRANSACTION} {
		key, err := s.getCompositeKey(ctx, VEHICLE, []string{string(status), id})
		if err != nil {
			return fmt.Errorf("创建复合键失败：%v", err)
		}

		exists, err := ctx.GetStub().GetState(key)
		if err != nil {
			return fmt.Errorf("查询车辆信息失败：%v", err)
		}
		if exists != nil {
			return fmt.Errorf("车辆ID %s 已存在", id)
		}
	}

	// 创建车辆信息
	vehicle := Vehicle{
		ID:           id,
		Model:        model,
		Year:         year,
		Brand:        brand,
		Mileage:      mileage,
		Condition:    condition,
		CurrentOwner: owner,
		Status:       NORMAL,
		CreateTime:   createTime,
		UpdateTime:   createTime,
	}

	// 保存车辆信息（复合键：类型_状态_ID）
	key, err := s.getCompositeKey(ctx, VEHICLE, []string{string(NORMAL), id})
	if err != nil {
		return err
	}

	err = s.putState(ctx, key, vehicle)
	if err != nil {
		return err
	}

	return nil
}

// CreateTransaction 生成交易（仅交易平台组织可以调用）
func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, txID string, vehicleID string, seller string, buyer string, price float64, createTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是交易平台组织的成员
	if clientMSPID != TRADE_ORG_MSPID {
		return fmt.Errorf("只有交易平台组织成员才能生成交易")
	}

	// 参数验证
	if len(txID) == 0 {
		return fmt.Errorf("交易ID不能为空")
	}
	if len(vehicleID) == 0 {
		return fmt.Errorf("车辆ID不能为空")
	}
	if len(seller) == 0 {
		return fmt.Errorf("卖家不能为空")
	}
	if len(buyer) == 0 {
		return fmt.Errorf("买家不能为空")
	}
	if seller == buyer {
		return fmt.Errorf("买家和卖家不能是同一人")
	}
	if price <= 0 {
		return fmt.Errorf("价格必须大于0")
	}

	// 查询车辆信息
	vehicleKey, err := s.getCompositeKey(ctx, VEHICLE, []string{string(NORMAL), vehicleID})
	if err != nil {
		return err
	}

	var vehicle Vehicle
	err = s.getState(ctx, vehicleKey, &vehicle)
	if err != nil {
		return err
	}

	// 检查卖家是否是车辆所有者
	if vehicle.CurrentOwner != seller {
		return fmt.Errorf("卖家不是车辆所有者")
	}

	// 生成交易信息
	transaction := Transaction{
		ID:         txID,
		VehicleID:  vehicleID,
		Seller:     seller,
		Buyer:      buyer,
		Price:      price,
		Status:     PENDING,
		CreateTime: createTime,
		UpdateTime: createTime,
	}

	// 更新车辆状态
	vehicle.Status = IN_TRANSACTION
	vehicle.UpdateTime = createTime

	// 保存状态
	txKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(PENDING), txID})
	if err != nil {
		return err
	}

	// 删除旧的车辆记录
	err = ctx.GetStub().DelState(vehicleKey)
	if err != nil {
		return fmt.Errorf("删除旧的车辆记录失败：%v", err)
	}

	// 创建新的车辆记录（使用新状态）
	newVehicleKey, err := s.getCompositeKey(ctx, VEHICLE, []string{string(IN_TRANSACTION), vehicleID})
	if err != nil {
		return err
	}

	err = s.putState(ctx, txKey, transaction)
	if err != nil {
		return err
	}

	err = s.putState(ctx, newVehicleKey, vehicle)
	if err != nil {
		return err
	}

	return nil
}

// CompleteTransaction 完成交易（仅银行组织可以调用）
func (s *SmartContract) CompleteTransaction(ctx contractapi.TransactionContextInterface, txID string, updateTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是银行组织的成员
	if clientMSPID != BANK_ORG_MSPID {
		return fmt.Errorf("只有银行组织成员才能完成交易")
	}

	// 查询交易信息
	txKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(PENDING), txID})
	if err != nil {
		return err
	}

	var transaction Transaction
	err = s.getState(ctx, txKey, &transaction)
	if err != nil {
		return err
	}

	// 查询车辆信息
	vehicleKey, err := s.getCompositeKey(ctx, VEHICLE, []string{string(IN_TRANSACTION), transaction.VehicleID})
	if err != nil {
		return err
	}

	var vehicle Vehicle
	err = s.getState(ctx, vehicleKey, &vehicle)
	if err != nil {
		return err
	}

	// 更新状态
	vehicle.CurrentOwner = transaction.Buyer
	vehicle.Status = NORMAL
	vehicle.UpdateTime = updateTime

	transaction.Status = COMPLETED
	transaction.UpdateTime = updateTime

	// 删除旧记录
	err = ctx.GetStub().DelState(txKey)
	if err != nil {
		return fmt.Errorf("删除旧的交易记录失败：%v", err)
	}

	err = ctx.GetStub().DelState(vehicleKey)
	if err != nil {
		return fmt.Errorf("删除旧的车辆记录失败：%v", err)
	}

	// 创建新记录
	newTxKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(COMPLETED), txID})
	if err != nil {
		return err
	}

	newVehicleKey, err := s.getCompositeKey(ctx, VEHICLE, []string{string(NORMAL), transaction.VehicleID})
	if err != nil {
		return err
	}

	err = s.putState(ctx, newTxKey, transaction)
	if err != nil {
		return err
	}

	err = s.putState(ctx, newVehicleKey, vehicle)
	if err != nil {
		return err
	}

	return nil
}

// QueryVehicle 查询车辆信息
func (s *SmartContract) QueryVehicle(ctx contractapi.TransactionContextInterface, id string) (*Vehicle, error) {
	// 遍历所有可能的状态查询车辆
	for _, status := range []VehicleStatus{NORMAL, IN_TRANSACTION} {
		key, err := s.getCompositeKey(ctx, VEHICLE, []string{string(status), id})
		if err != nil {
			return nil, fmt.Errorf("创建复合键失败：%v", err)
		}

		bytes, err := ctx.GetStub().GetState(key)
		if err != nil {
			return nil, fmt.Errorf("查询车辆信息失败：%v", err)
		}
		if bytes != nil {
			var vehicle Vehicle
			err = json.Unmarshal(bytes, &vehicle)
			if err != nil {
				return nil, fmt.Errorf("解析车辆信息失败：%v", err)
			}
			return &vehicle, nil
		}
	}

	return nil, fmt.Errorf("车辆ID %s 不存在", id)
}

// QueryTransaction 查询交易信息
func (s *SmartContract) QueryTransaction(ctx contractapi.TransactionContextInterface, txID string) (*Transaction, error) {
	// 遍历所有可能的状态查询交易
	for _, status := range []TransactionStatus{PENDING, COMPLETED} {
		key, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(status), txID})
		if err != nil {
			return nil, fmt.Errorf("创建复合键失败：%v", err)
		}

		bytes, err := ctx.GetStub().GetState(key)
		if err != nil {
			return nil, fmt.Errorf("查询交易信息失败：%v", err)
		}
		if bytes != nil {
			var transaction Transaction
			err = json.Unmarshal(bytes, &transaction)
			if err != nil {
				return nil, fmt.Errorf("解析交易信息失败：%v", err)
			}
			return &transaction, nil
		}
	}

	return nil, fmt.Errorf("交易ID %s 不存在", txID)
}

// QueryVehicleList 分页查询车辆列表
func (s *SmartContract) QueryVehicleList(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string, status string) (*QueryResult, error) {
	var iterator shim.StateQueryIteratorInterface
	var metadata *peer.QueryResponseMetadata
	var err error

	if status != "" {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			VEHICLE,
			[]string{status},
			pageSize,
			bookmark,
		)
	} else {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			VEHICLE,
			[]string{},
			pageSize,
			bookmark,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("查询列表失败：%v", err)
	}
	defer iterator.Close()

	records := make([]interface{}, 0)
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一条记录失败：%v", err)
		}

		var vehicle Vehicle
		err = json.Unmarshal(queryResponse.Value, &vehicle)
		if err != nil {
			return nil, fmt.Errorf("解析车辆信息失败：%v", err)
		}

		records = append(records, vehicle)
	}

	return &QueryResult{
		Records:             records,
		RecordsCount:        int32(len(records)),
		Bookmark:            metadata.Bookmark,
		FetchedRecordsCount: metadata.FetchedRecordsCount,
	}, nil
}

// QueryTransactionList 分页查询交易列表
func (s *SmartContract) QueryTransactionList(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string, status string) (*QueryResult, error) {
	var iterator shim.StateQueryIteratorInterface
	var metadata *peer.QueryResponseMetadata
	var err error

	if status != "" {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			TRANSACTION,
			[]string{status},
			pageSize,
			bookmark,
		)
	} else {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			TRANSACTION,
			[]string{},
			pageSize,
			bookmark,
		)
	}

	if err != nil {
		return nil, fmt.Errorf("查询列表失败：%v", err)
	}
	defer iterator.Close()

	records := make([]interface{}, 0)
	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, fmt.Errorf("获取下一条记录失败：%v", err)
		}

		var transaction Transaction
		err = json.Unmarshal(queryResponse.Value, &transaction)
		if err != nil {
			return nil, fmt.Errorf("解析交易信息失败：%v", err)
		}

		records = append(records, transaction)
	}

	return &QueryResult{
		Records:             records,
		RecordsCount:        int32(len(records)),
		Bookmark:            metadata.Bookmark,
		FetchedRecordsCount: metadata.FetchedRecordsCount,
	}, nil
}

// Hello 用于验证
func (s *SmartContract) Hello(ctx contractapi.TransactionContextInterface) (string, error) {
	return "hello", nil
}

// InitLedger 初始化账本
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("InitLedger")
	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("创建智能合约失败：%v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("启动智能合约失败：%v", err)
	}
}
