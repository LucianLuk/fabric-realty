package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings" // Added for string manipulation
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
)

// SmartContract 提供二手车交易的功能 (修改注释)
type SmartContract struct {
	contractapi.Contract
}

// 文档类型常量（用于创建复合键）
const (
	CAR         = "CAR"  // 汽车信息 (修改常量)
	TRANSACTION = "TX"   // 交易信息
	CERTIFICATE = "CERT" // 证书信息 (新增)
)

// CertificateStatus 证书状态 (新增, MVP 暂未使用)
// type CertificateStatus string
// const (
// 	CERT_ACTIVE  CertificateStatus = "ACTIVE"
// 	CERT_REVOKED CertificateStatus = "REVOKED"
// )

// CarStatus 汽车状态 (修改类型名)
type CarStatus string

const (
	AVAILABLE      CarStatus = "AVAILABLE"      // 待售 (修改状态)
	IN_TRANSACTION CarStatus = "IN_TRANSACTION" // 交易中
	SOLD           CarStatus = "SOLD"           // 已售 (新增状态)
)

// TransactionStatus 交易状态
type TransactionStatus string

const (
	PENDING   TransactionStatus = "PENDING"   // 待付款
	COMPLETED TransactionStatus = "COMPLETED" // 已完成
	// 可以考虑添加 CANCELLED 状态，但当前逻辑未包含
)

// Car 汽车信息 (修改结构体名和字段)
type Car struct {
	ID           string    `json:"id"`           // 汽车ID (例如车牌号)
	Model        string    `json:"model"`        // 车型
	VIN          string    `json:"vin"`          // 车辆识别代号
	CurrentOwner string    `json:"currentOwner"` // 当前所有者
	Status       CarStatus `json:"status"`       // 状态
	CreateTime   time.Time `json:"createTime"`   // 创建时间
	UpdateTime   time.Time `json:"updateTime"`   // 更新时间
}

// Transaction 交易信息 (修改字段)
type Transaction struct {
	ID         string            `json:"id"`         // 交易ID
	CarID      string            `json:"carId"`      // 汽车ID (修改字段名)
	Seller     string            `json:"seller"`     // 卖家
	Buyer      string            `json:"buyer"`      // 买家
	Price      float64           `json:"price"`      // 成交价格
	Status     TransactionStatus `json:"status"`     // 状态
	CreateTime time.Time         `json:"createTime"` // 创建时间
	UpdateTime time.Time         `json:"updateTime"` // 更新时间
}

// Certificate 证书信息 (新增 MVP 结构)
type Certificate struct {
	CertID       string    `json:"certId"`       // 证书唯一ID
	CarID        string    `json:"carId"`        // 关联的汽车ID
	CertType     string    `json:"certType"`     // 证书类型 (e.g., "REGISTRATION", "OTHER")
	FileHash     string    `json:"fileHash"`     // 文件SHA256哈希
	FileLocation string    `json:"fileLocation"` // 本地文件路径
	UploadTime   time.Time `json:"uploadTime"`   // 上传时间
}

// QueryResult 分页查询结果
type QueryResult struct {
	Records             []interface{} `json:"records"`             // 记录列表
	RecordsCount        int32         `json:"recordsCount"`        // 本次返回的记录数
	Bookmark            string        `json:"bookmark"`            // 书签，用于下一页查询
	FetchedRecordsCount int32         `json:"fetchedRecordsCount"` // 总共获取的记录数 (注意: Fabric v2.x 中此字段可能不准确)
}

// 组织 MSP ID 常量 (修改常量名)
const (
	CAR_DEALER_ORG_MSPID = "Org1MSP" // 汽车经销商组织 MSP ID
	BANK_ORG_MSPID       = "Org2MSP" // 银行组织 MSP ID
	TRADE_ORG_MSPID      = "Org3MSP" // 交易平台组织 MSP ID
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

// CreateCar 创建汽车信息（仅汽车经销商组织可以调用）(修改函数名和逻辑)
func (s *SmartContract) CreateCar(ctx contractapi.TransactionContextInterface, id string, model string, vin string, owner string, createTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是汽车经销商组织的成员 (修改常量)
	if clientMSPID != CAR_DEALER_ORG_MSPID {
		return fmt.Errorf("只有汽车经销商组织成员才能创建汽车信息") // 修改错误信息
	}

	// 参数验证 (修改验证字段)
	if len(id) == 0 {
		return fmt.Errorf("汽车ID不能为空")
	}
	if len(model) == 0 {
		return fmt.Errorf("车型不能为空")
	}
	if len(vin) != 17 { // VIN 通常是17位
		return fmt.Errorf("VIN必须是17位")
	}
	if len(owner) == 0 {
		return fmt.Errorf("所有者不能为空")
	}

	// 检查汽车是否已存在（检查所有可能的状态）(修改常量和状态)
	for _, status := range []CarStatus{AVAILABLE, IN_TRANSACTION, SOLD} {
		key, err := s.getCompositeKey(ctx, CAR, []string{string(status), id}) // 修改常量
		if err != nil {
			return fmt.Errorf("创建复合键失败：%v", err)
		}

		exists, err := ctx.GetStub().GetState(key)
		if err != nil {
			return fmt.Errorf("查询汽车信息失败：%v", err) // 修改错误信息
		}
		if exists != nil {
			return fmt.Errorf("汽车ID %s 已存在", id) // 修改错误信息
		}
	}

	// 创建汽车信息 (修改结构体和字段)
	car := Car{
		ID:           id,
		Model:        model,
		VIN:          vin,
		CurrentOwner: owner,
		Status:       AVAILABLE, // 初始状态为待售
		CreateTime:   createTime,
		UpdateTime:   createTime,
	}

	// 保存汽车信息（复合键：类型_状态_ID）(修改常量和状态)
	key, err := s.getCompositeKey(ctx, CAR, []string{string(AVAILABLE), id})
	if err != nil {
		return err
	}

	err = s.putState(ctx, key, car)
	if err != nil {
		return err
	}

	return nil
}

// CreateTransaction 生成交易（仅交易平台组织可以调用）(修改逻辑)
func (s *SmartContract) CreateTransaction(ctx contractapi.TransactionContextInterface, txID string, carID string, seller string, buyer string, price float64, createTime time.Time) error {
	// 检查调用者身份
	clientMSPID, err := s.getClientIdentityMSPID(ctx)
	if err != nil {
		return fmt.Errorf("获取调用者身份失败：%v", err)
	}

	// 验证是否是交易平台组织的成员
	if clientMSPID != TRADE_ORG_MSPID {
		return fmt.Errorf("只有交易平台组织成员才能生成交易")
	}

	// 参数验证 (修改字段名)
	if len(txID) == 0 {
		return fmt.Errorf("交易ID不能为空")
	}
	if len(carID) == 0 {
		return fmt.Errorf("汽车ID不能为空") // 修改错误信息
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

	// 查询汽车信息 (修改常量、状态和变量)
	carKey, err := s.getCompositeKey(ctx, CAR, []string{string(AVAILABLE), carID})
	if err != nil {
		return err
	}

	var car Car
	err = s.getState(ctx, carKey, &car)
	if err != nil {
		// 如果在 AVAILABLE 状态找不到，尝试在 IN_TRANSACTION 状态找 (防止重复创建交易)
		inTransactionKey, keyErr := s.getCompositeKey(ctx, CAR, []string{string(IN_TRANSACTION), carID})
		if keyErr == nil {
			existsBytes, getErr := ctx.GetStub().GetState(inTransactionKey)
			if getErr == nil && existsBytes != nil {
				return fmt.Errorf("汽车 %s 正在交易中，无法创建新交易", carID)
			}
		}
		// 如果在 SOLD 状态找到
		soldKey, keyErr := s.getCompositeKey(ctx, CAR, []string{string(SOLD), carID})
		if keyErr == nil {
			existsBytes, getErr := ctx.GetStub().GetState(soldKey)
			if getErr == nil && existsBytes != nil {
				return fmt.Errorf("汽车 %s 已售出，无法创建新交易", carID)
			}
		}
		return fmt.Errorf("查询汽车信息失败或汽车非待售状态：%v", err) // 修改错误信息
	}

	// 检查卖家是否是汽车所有者 (修改变量)
	if car.CurrentOwner != seller {
		return fmt.Errorf("卖家不是汽车所有者") // 修改错误信息
	}

	// 生成交易信息 (修改字段名)
	transaction := Transaction{
		ID:         txID,
		CarID:      carID,
		Seller:     seller,
		Buyer:      buyer,
		Price:      price,
		Status:     PENDING,
		CreateTime: createTime,
		UpdateTime: createTime,
	}

	// 更新汽车状态 (修改变量和状态)
	car.Status = IN_TRANSACTION
	car.UpdateTime = createTime

	// 保存状态 (修改常量和变量)
	txKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(PENDING), txID})
	if err != nil {
		return err
	}

	// 删除旧的汽车记录 (修改变量)
	err = ctx.GetStub().DelState(carKey)
	if err != nil {
		return fmt.Errorf("删除旧的汽车记录失败：%v", err) // 修改错误信息
	}

	// 创建新的汽车记录（使用新状态）(修改常量、状态和变量)
	newCarKey, err := s.getCompositeKey(ctx, CAR, []string{string(IN_TRANSACTION), carID})
	if err != nil {
		return err
	}

	err = s.putState(ctx, txKey, transaction)
	if err != nil {
		return err
	}

	err = s.putState(ctx, newCarKey, car)
	if err != nil {
		return err
	}

	return nil
}

// CompleteTransaction 完成交易（仅银行组织可以调用）(修改逻辑)
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

	// 查询汽车信息 (修改常量、状态和变量)
	carKey, err := s.getCompositeKey(ctx, CAR, []string{string(IN_TRANSACTION), transaction.CarID}) // 使用 CarID
	if err != nil {
		return err
	}

	var car Car
	err = s.getState(ctx, carKey, &car)
	if err != nil {
		return err
	}

	// 更新状态 (修改变量和状态)
	car.CurrentOwner = transaction.Buyer
	car.Status = SOLD // 交易完成后状态变为 SOLD
	car.UpdateTime = updateTime

	transaction.Status = COMPLETED
	transaction.UpdateTime = updateTime

	// 删除旧记录 (修改变量)
	err = ctx.GetStub().DelState(txKey)
	if err != nil {
		return fmt.Errorf("删除旧的交易记录失败：%v", err)
	}

	err = ctx.GetStub().DelState(carKey)
	if err != nil {
		return fmt.Errorf("删除旧的汽车记录失败：%v", err) // 修改错误信息
	}

	// 创建新记录 (修改常量、状态和变量)
	newTxKey, err := s.getCompositeKey(ctx, TRANSACTION, []string{string(COMPLETED), txID})
	if err != nil {
		return err
	}

	newCarKey, err := s.getCompositeKey(ctx, CAR, []string{string(SOLD), transaction.CarID}) // 新状态为 SOLD
	if err != nil {
		return err
	}

	err = s.putState(ctx, newTxKey, transaction)
	if err != nil {
		return err
	}

	err = s.putState(ctx, newCarKey, car)
	if err != nil {
		return err
	}

	return nil
}

// QueryCar 查询汽车信息 (修改函数名和逻辑)
func (s *SmartContract) QueryCar(ctx contractapi.TransactionContextInterface, id string) (*Car, error) {
	// 遍历所有可能的状态查询汽车 (修改常量、状态和变量)
	for _, status := range []CarStatus{AVAILABLE, IN_TRANSACTION, SOLD} {
		key, err := s.getCompositeKey(ctx, CAR, []string{string(status), id})
		if err != nil {
			return nil, fmt.Errorf("创建复合键失败：%v", err)
		}

		bytes, err := ctx.GetStub().GetState(key)
		if err != nil {
			return nil, fmt.Errorf("查询汽车信息失败：%v", err) // 修改错误信息
		}
		if bytes != nil {
			var car Car
			err = json.Unmarshal(bytes, &car)
			if err != nil {
				return nil, fmt.Errorf("解析汽车信息失败：%v", err) // 修改错误信息
			}
			return &car, nil
		}
	}

	return nil, fmt.Errorf("汽车ID %s 不存在", id) // 修改错误信息
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

// QueryCarList 分页查询汽车列表 (修改函数名和逻辑)
func (s *SmartContract) QueryCarList(ctx contractapi.TransactionContextInterface, pageSize int32, bookmark string, status string) (*QueryResult, error) {
	var iterator shim.StateQueryIteratorInterface
	var metadata *peer.QueryResponseMetadata
	var err error

	// 验证 status 是否是有效的 CarStatus
	isValidStatus := false
	if status != "" {
		for _, validStatus := range []CarStatus{AVAILABLE, IN_TRANSACTION, SOLD} {
			if CarStatus(status) == validStatus {
				isValidStatus = true
				break
			}
		}
		if !isValidStatus {
			return nil, fmt.Errorf("无效的汽车状态: %s", status)
		}
	}

	// 根据 status 查询 (修改常量)
	if status != "" {
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			CAR,
			[]string{status},
			pageSize,
			bookmark,
		)
	} else {
		// 查询所有状态
		iterator, metadata, err = ctx.GetStub().GetStateByPartialCompositeKeyWithPagination(
			CAR,
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

		var car Car // 修改变量类型
		err = json.Unmarshal(queryResponse.Value, &car)
		if err != nil {
			return nil, fmt.Errorf("解析汽车信息失败：%v", err) // 修改错误信息
		}

		records = append(records, car)
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

	// 验证 status 是否是有效的 TransactionStatus
	isValidStatus := false
	if status != "" {
		for _, validStatus := range []TransactionStatus{PENDING, COMPLETED} {
			if TransactionStatus(status) == validStatus {
				isValidStatus = true
				break
			}
		}
		if !isValidStatus {
			return nil, fmt.Errorf("无效的交易状态: %s", status)
		}
	}

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

// --- 新增证书相关函数 ---

// AddCertificate 添加证书信息 (MVP)
func (s *SmartContract) AddCertificate(ctx contractapi.TransactionContextInterface, certJsonString string) error {
	var cert Certificate
	err := json.Unmarshal([]byte(certJsonString), &cert)
	if err != nil {
		return fmt.Errorf("解析证书 JSON 失败: %v", err)
	}

	// 参数验证 (基础)
	if len(cert.CertID) == 0 {
		return fmt.Errorf("证书ID不能为空")
	}
	if len(cert.CarID) == 0 {
		return fmt.Errorf("关联的汽车ID不能为空")
	}
	if len(cert.FileHash) == 0 {
		return fmt.Errorf("文件哈希不能为空")
	}
	if len(cert.FileLocation) == 0 {
		return fmt.Errorf("文件位置不能为空")
	}
	// 可以在此添加更多验证，例如 CertType 是否在允许列表内

	// 新增验证: 确保 FileLocation 符合预期的格式。
	// 1. FileLocation 不应是绝对路径 (简单检查常见的绝对路径指示符)
	if strings.HasPrefix(cert.FileLocation, "/") || strings.Contains(cert.FileLocation, ":\\") || strings.Contains(cert.FileLocation, ":/") {
		return fmt.Errorf("文件位置 (FileLocation) '%s' 不应是绝对路径，应为相对于 '%s' 的路径，例如 'CAR_ID/filename.ext'", cert.FileLocation, "application/server/data/certificates/")
	}
	// 2. FileLocation 应该以 CarID 开头，后跟一个路径分隔符 '/'
	//    这与用户期望在 "data/certificates里面的对应车辆的文件夹" 中找到文件一致
	expectedPrefix := cert.CarID + "/"
	if !strings.HasPrefix(cert.FileLocation, expectedPrefix) {
		return fmt.Errorf("文件位置 (FileLocation) '%s' 必须以车辆ID '%s/' 开头 (例如: '%sfilename.ext')", cert.FileLocation, cert.CarID, expectedPrefix)
	}
	// 3. FileLocation CarID/ 之后必须有文件名 (文件名不能为空)
	if len(cert.FileLocation) <= len(expectedPrefix) {
		return fmt.Errorf("文件位置 (FileLocation) '%s' 在车辆ID '%s/' 之后必须包含有效的文件名", cert.FileLocation, cert.CarID)
	}
	// 4. 文件名部分不应包含额外的路径分隔符 (即文件应直接在 CarID 目录下)
	remainingPath := cert.FileLocation[len(expectedPrefix):]
	if strings.Contains(remainingPath, "/") {
		return fmt.Errorf("文件位置 (FileLocation) '%s' 在车辆ID '%s/' 之后不应包含额外的子目录路径，应直接是文件名", cert.FileLocation, cert.CarID)
	}


	// 检查证书是否已存在
	certKey, err := s.getCompositeKey(ctx, CERTIFICATE, []string{cert.CertID})
	if err != nil {
		return fmt.Errorf("创建证书复合键失败: %v", err)
	}
	existsBytes, err := ctx.GetStub().GetState(certKey)
	if err != nil {
		return fmt.Errorf("查询证书是否存在时出错: %v", err)
	}
	if existsBytes != nil {
		return fmt.Errorf("证书ID %s 已存在", cert.CertID)
	}

	// 保存证书
	err = s.putState(ctx, certKey, cert)
	if err != nil {
		return fmt.Errorf("保存证书失败: %v", err)
	}

	return nil
}

// GetAllCertificates 获取所有证书记录 (MVP - 后端过滤)
func (s *SmartContract) GetAllCertificates(ctx contractapi.TransactionContextInterface) ([]*Certificate, error) {
	// 使用范围查询获取所有以 CERTIFICATE 开头的键
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(CERTIFICATE, []string{})
	if err != nil {
		return nil, fmt.Errorf("获取证书列表失败: %v", err)
	}
	defer resultsIterator.Close()

	var certificates []*Certificate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("读取证书记录失败: %v", err)
		}

		var cert Certificate
		err = json.Unmarshal(queryResponse.Value, &cert)
		if err != nil {
			// 记录错误但继续处理其他记录
			log.Printf("解析证书失败 (Key: %s): %v", queryResponse.Key, err)
			continue
		}
		certificates = append(certificates, &cert)
	}

	return certificates, nil
}

// GetCertificate returns the certificate stored in the world state with the given ID. (新增)
func (s *SmartContract) GetCertificate(ctx contractapi.TransactionContextInterface, certId string) (*Certificate, error) {
	if len(certId) == 0 {
		return nil, fmt.Errorf("证书ID不能为空")
	}
	certKey, err := s.getCompositeKey(ctx, CERTIFICATE, []string{certId})
	if err != nil {
		return nil, fmt.Errorf("创建证书复合键失败: %v", err)
	}

	certJSON, err := ctx.GetStub().GetState(certKey)
	if err != nil {
		return nil, fmt.Errorf("读取证书状态失败: %v", err)
	}
	if certJSON == nil {
		return nil, fmt.Errorf("证书ID %s 不存在", certId)
	}

	var cert Certificate
	err = json.Unmarshal(certJSON, &cert)
	if err != nil {
		return nil, fmt.Errorf("解析证书JSON失败: %v", err)
	}

	return &cert, nil
}

// Hello 用于验证
func (s *SmartContract) Hello(ctx contractapi.TransactionContextInterface) (string, error) {
	return "hello", nil
}

// InitLedger 初始化账本
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	log.Println("InitLedger for Car Trading Chaincode") // 修改日志信息
	// 可以在这里添加一些初始的汽车数据用于测试
	// 例如:
	// cars := []Car{
	// 	{ID: "CAR001", Model: "特斯拉 Model 3", VIN: "VIN123456789ABCDE", CurrentOwner: "Alice", Status: AVAILABLE, CreateTime: time.Now(), UpdateTime: time.Now()},
	// 	{ID: "CAR002", Model: "比亚迪 汉", VIN: "VINFGHIJKLMNOPQRS", CurrentOwner: "Bob", Status: AVAILABLE, CreateTime: time.Now(), UpdateTime: time.Now()},
	// }
	// for _, car := range cars {
	// 	key, err := s.getCompositeKey(ctx, CAR, []string{string(car.Status), car.ID})
	// 	if err != nil {
	// 		return fmt.Errorf("创建汽车 %s 的复合键失败: %v", car.ID, err)
	// 	}
	// 	err = s.putState(ctx, key, car)
	// 	if err != nil {
	// 		return fmt.Errorf("保存汽车 %s 失败: %v", car.ID, err)
	// 	}
	// }
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
