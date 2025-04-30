package main

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/v2/shim"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
	"github.com/hyperledger/fabric-protos-go-apiv2/peer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 模拟 TransactionContext
type MockTransactionContext struct {
	contractapi.TransactionContext
	MockStub *MockChaincodeStub
}

// 获取 stub
func (m *MockTransactionContext) GetStub() shim.ChaincodeStubInterface {
	return m.MockStub
}

// 模拟 ChaincodeStub
type MockChaincodeStub struct {
	mock.Mock
	shim.ChaincodeStubInterface
}

// 模拟 GetState 方法
func (m *MockChaincodeStub) GetState(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

// 模拟 PutState 方法
func (m *MockChaincodeStub) PutState(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

// 模拟 DelState 方法
func (m *MockChaincodeStub) DelState(key string) error {
	args := m.Called(key)
	return args.Error(0)
}

// 模拟 CreateCompositeKey 方法
func (m *MockChaincodeStub) CreateCompositeKey(objectType string, attributes []string) (string, error) {
	args := m.Called(objectType, attributes)
	return args.String(0), args.Error(1)
}

// 模拟 GetStateByPartialCompositeKeyWithPagination 方法
func (m *MockChaincodeStub) GetStateByPartialCompositeKeyWithPagination(
	objectType string,
	keys []string,
	pageSize int32,
	bookmark string,
) (shim.StateQueryIteratorInterface, *peer.QueryResponseMetadata, error) {
	args := m.Called(objectType, keys, pageSize, bookmark)
	return args.Get(0).(shim.StateQueryIteratorInterface), args.Get(1).(*peer.QueryResponseMetadata), args.Error(2)
}

// 模拟 ClientIdentity
type MockClientIdentity struct {
	mock.Mock
}

// 模拟 GetMSPID 方法
func (m *MockClientIdentity) GetMSPID() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

// 测试 CreateCar 函数
func TestCreateCar(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 设置模拟行为
	mockStub.On("CreateCompositeKey", CAR, []string{string(AVAILABLE), "CAR001"}).Return("CAR_AVAILABLE_CAR001", nil)
	mockStub.On("GetState", "CAR_AVAILABLE_CAR001").Return([]byte(nil), nil)
	mockStub.On("CreateCompositeKey", CAR, []string{string(IN_TRANSACTION), "CAR001"}).Return("CAR_IN_TRANSACTION_CAR001", nil)
	mockStub.On("GetState", "CAR_IN_TRANSACTION_CAR001").Return([]byte(nil), nil)
	mockStub.On("CreateCompositeKey", CAR, []string{string(SOLD), "CAR001"}).Return("CAR_SOLD_CAR001", nil)
	mockStub.On("GetState", "CAR_SOLD_CAR001").Return([]byte(nil), nil)
	mockStub.On("PutState", "CAR_AVAILABLE_CAR001", mock.Anything).Return(nil)

	// 模拟 ClientIdentity
	mockClientIdentity := new(MockClientIdentity)
	mockClientIdentity.On("GetMSPID").Return(CAR_DEALER_ORG_MSPID, nil)

	// 替换原始的 getClientIdentityMSPID 方法
	originalGetClientIdentityMSPID := contract.getClientIdentityMSPID
	contract.getClientIdentityMSPID = func(ctx contractapi.TransactionContextInterface) (string, error) {
		return CAR_DEALER_ORG_MSPID, nil
	}
	defer func() { contract.getClientIdentityMSPID = originalGetClientIdentityMSPID }()

	// 执行测试
	createTime := time.Now()
	err := contract.CreateCar(mockCtx, "CAR001", "特斯拉 Model 3", "VIN123456789ABCDE", "Alice", createTime)

	// 验证结果
	assert.NoError(t, err)
	mockStub.AssertExpectations(t)
}

// 测试 QueryCar 函数
func TestQueryCar(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 准备测试数据
	car := Car{
		ID:           "CAR001",
		Model:        "特斯拉 Model 3",
		VIN:          "VIN123456789ABCDE",
		CurrentOwner: "Alice",
		Status:       AVAILABLE,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	carJSON, _ := json.Marshal(car)

	// 设置模拟行为
	mockStub.On("CreateCompositeKey", CAR, []string{string(AVAILABLE), "CAR001"}).Return("CAR_AVAILABLE_CAR001", nil)
	mockStub.On("GetState", "CAR_AVAILABLE_CAR001").Return(carJSON, nil)

	// 执行测试
	result, err := contract.QueryCar(mockCtx, "CAR001")

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, car.ID, result.ID)
	assert.Equal(t, car.Model, result.Model)
	assert.Equal(t, car.VIN, result.VIN)
	assert.Equal(t, car.CurrentOwner, result.CurrentOwner)
	assert.Equal(t, car.Status, result.Status)
	mockStub.AssertExpectations(t)
}

// 测试 CreateTransaction 函数
func TestCreateTransaction(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 准备测试数据
	car := Car{
		ID:           "CAR001",
		Model:        "特斯拉 Model 3",
		VIN:          "VIN123456789ABCDE",
		CurrentOwner: "Alice",
		Status:       AVAILABLE,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	carJSON, _ := json.Marshal(car)

	// 设置模拟行为
	mockStub.On("CreateCompositeKey", CAR, []string{string(AVAILABLE), "CAR001"}).Return("CAR_AVAILABLE_CAR001", nil)
	mockStub.On("GetState", "CAR_AVAILABLE_CAR001").Return(carJSON, nil)
	mockStub.On("DelState", "CAR_AVAILABLE_CAR001").Return(nil)
	mockStub.On("CreateCompositeKey", CAR, []string{string(IN_TRANSACTION), "CAR001"}).Return("CAR_IN_TRANSACTION_CAR001", nil)
	mockStub.On("PutState", "CAR_IN_TRANSACTION_CAR001", mock.Anything).Return(nil)
	mockStub.On("CreateCompositeKey", TRANSACTION, []string{string(PENDING), "TX001"}).Return("TRANSACTION_PENDING_TX001", nil)
	mockStub.On("PutState", "TRANSACTION_PENDING_TX001", mock.Anything).Return(nil)

	// 替换原始的 getClientIdentityMSPID 方法
	originalGetClientIdentityMSPID := contract.getClientIdentityMSPID
	contract.getClientIdentityMSPID = func(ctx contractapi.TransactionContextInterface) (string, error) {
		return TRADE_ORG_MSPID, nil
	}
	defer func() { contract.getClientIdentityMSPID = originalGetClientIdentityMSPID }()

	// 执行测试
	createTime := time.Now()
	err := contract.CreateTransaction(mockCtx, "TX001", "CAR001", "Alice", "Bob", 100000.0, createTime)

	// 验证结果
	assert.NoError(t, err)
	mockStub.AssertExpectations(t)
}

// 测试 CompleteTransaction 函数
func TestCompleteTransaction(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 准备测试数据
	transaction := Transaction{
		ID:         "TX001",
		CarID:      "CAR001",
		Seller:     "Alice",
		Buyer:      "Bob",
		Price:      100000.0,
		Status:     PENDING,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	transactionJSON, _ := json.Marshal(transaction)

	car := Car{
		ID:           "CAR001",
		Model:        "特斯拉 Model 3",
		VIN:          "VIN123456789ABCDE",
		CurrentOwner: "Alice",
		Status:       IN_TRANSACTION,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	carJSON, _ := json.Marshal(car)

	// 设置模拟行为
	mockStub.On("CreateCompositeKey", TRANSACTION, []string{string(PENDING), "TX001"}).Return("TRANSACTION_PENDING_TX001", nil)
	mockStub.On("GetState", "TRANSACTION_PENDING_TX001").Return(transactionJSON, nil)
	mockStub.On("CreateCompositeKey", CAR, []string{string(IN_TRANSACTION), "CAR001"}).Return("CAR_IN_TRANSACTION_CAR001", nil)
	mockStub.On("GetState", "CAR_IN_TRANSACTION_CAR001").Return(carJSON, nil)
	mockStub.On("DelState", "TRANSACTION_PENDING_TX001").Return(nil)
	mockStub.On("DelState", "CAR_IN_TRANSACTION_CAR001").Return(nil)
	mockStub.On("CreateCompositeKey", TRANSACTION, []string{string(COMPLETED), "TX001"}).Return("TRANSACTION_COMPLETED_TX001", nil)
	mockStub.On("PutState", "TRANSACTION_COMPLETED_TX001", mock.Anything).Return(nil)
	mockStub.On("CreateCompositeKey", CAR, []string{string(SOLD), "CAR001"}).Return("CAR_SOLD_CAR001", nil)
	mockStub.On("PutState", "CAR_SOLD_CAR001", mock.Anything).Return(nil)

	// 替换原始的 getClientIdentityMSPID 方法
	originalGetClientIdentityMSPID := contract.getClientIdentityMSPID
	contract.getClientIdentityMSPID = func(ctx contractapi.TransactionContextInterface) (string, error) {
		return BANK_ORG_MSPID, nil
	}
	defer func() { contract.getClientIdentityMSPID = originalGetClientIdentityMSPID }()

	// 执行测试
	updateTime := time.Now()
	err := contract.CompleteTransaction(mockCtx, "TX001", updateTime)

	// 验证结果
	assert.NoError(t, err)
	mockStub.AssertExpectations(t)
}

// 测试 QueryTransaction 函数
func TestQueryTransaction(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 准备测试数据
	transaction := Transaction{
		ID:         "TX001",
		CarID:      "CAR001",
		Seller:     "Alice",
		Buyer:      "Bob",
		Price:      100000.0,
		Status:     PENDING,
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	transactionJSON, _ := json.Marshal(transaction)

	// 设置模拟行为
	mockStub.On("CreateCompositeKey", TRANSACTION, []string{string(PENDING), "TX001"}).Return("TRANSACTION_PENDING_TX001", nil)
	mockStub.On("GetState", "TRANSACTION_PENDING_TX001").Return(transactionJSON, nil)
	mockStub.On("CreateCompositeKey", TRANSACTION, []string{string(COMPLETED), "TX001"}).Return("TRANSACTION_COMPLETED_TX001", nil)
	mockStub.On("GetState", "TRANSACTION_COMPLETED_TX001").Return([]byte(nil), nil)

	// 执行测试
	result, err := contract.QueryTransaction(mockCtx, "TX001")

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, transaction.ID, result.ID)
	assert.Equal(t, transaction.CarID, result.CarID)
	assert.Equal(t, transaction.Seller, result.Seller)
	assert.Equal(t, transaction.Buyer, result.Buyer)
	assert.Equal(t, transaction.Price, result.Price)
	assert.Equal(t, transaction.Status, result.Status)
	mockStub.AssertExpectations(t)
}

// 模拟 StateQueryIterator
type MockStateQueryIterator struct {
	mock.Mock
	shim.StateQueryIteratorInterface
	Results []*queryresult
	Index   int
}

type queryresult struct {
	Key    string
	Value  []byte
	Record interface{}
}

func (m *MockStateQueryIterator) HasNext() bool {
	return m.Index < len(m.Results)
}

func (m *MockStateQueryIterator) Next() (*shim.KV, error) {
	if m.Index >= len(m.Results) {
		return nil, fmt.Errorf("no more items")
	}
	result := m.Results[m.Index]
	m.Index++
	return &shim.KV{
		Key:   result.Key,
		Value: result.Value,
	}, nil
}

func (m *MockStateQueryIterator) Close() error {
	return nil
}

// 测试 QueryCarList 函数
func TestQueryCarList(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 准备测试数据
	car1 := Car{
		ID:           "CAR001",
		Model:        "特斯拉 Model 3",
		VIN:          "VIN123456789ABCDE",
		CurrentOwner: "Alice",
		Status:       AVAILABLE,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	car1JSON, _ := json.Marshal(car1)

	car2 := Car{
		ID:           "CAR002",
		Model:        "比亚迪 汉",
		VIN:          "VINFGHIJKLMNOPQRS",
		CurrentOwner: "Bob",
		Status:       AVAILABLE,
		CreateTime:   time.Now(),
		UpdateTime:   time.Now(),
	}
	car2JSON, _ := json.Marshal(car2)

	// 创建模拟迭代器
	mockIterator := &MockStateQueryIterator{
		Results: []*queryresult{
			{Key: "CAR_AVAILABLE_CAR001", Value: car1JSON, Record: car1},
			{Key: "CAR_AVAILABLE_CAR002", Value: car2JSON, Record: car2},
		},
	}

	// 创建模拟元数据
	mockMetadata := &peer.QueryResponseMetadata{
		Bookmark:            "bookmark1",
		FetchedRecordsCount: 2,
	}

	// 设置模拟行为
	mockStub.On("GetStateByPartialCompositeKeyWithPagination", CAR, []string{"AVAILABLE"}, int32(10), "").Return(mockIterator, mockMetadata, nil)

	// 执行测试
	result, err := contract.QueryCarList(mockCtx, 10, "", "AVAILABLE")

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result.Records))
	assert.Equal(t, int32(2), result.RecordsCount)
	assert.Equal(t, "bookmark1", result.Bookmark)
	assert.Equal(t, int32(2), result.FetchedRecordsCount)
	mockStub.AssertExpectations(t)
}

// 测试 AddCertificate 函数
func TestAddCertificate(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 准备测试数据
	cert := Certificate{
		CertID:       "CERT001",
		CarID:        "CAR001",
		CertType:     "REGISTRATION",
		FileHash:     "abcdef1234567890",
		FileLocation: "data/certificates/CAR001/CERT001.pdf",
		UploadTime:   time.Now(),
	}
	certJSON, _ := json.Marshal(cert)

	// 设置模拟行为
	mockStub.On("CreateCompositeKey", CERTIFICATE, []string{"CERT001"}).Return("CERTIFICATE_CERT001", nil)
	mockStub.On("GetState", "CERTIFICATE_CERT001").Return([]byte(nil), nil)
	mockStub.On("PutState", "CERTIFICATE_CERT001", mock.Anything).Return(nil)

	// 执行测试
	certString, _ := json.Marshal(cert)
	err := contract.AddCertificate(mockCtx, string(certString))

	// 验证结果
	assert.NoError(t, err)
	mockStub.AssertExpectations(t)
}

// 测试 GetCertificate 函数
func TestGetCertificate(t *testing.T) {
	// 创建模拟对象
	mockStub := new(MockChaincodeStub)
	mockCtx := &MockTransactionContext{
		MockStub: mockStub,
	}

	// 创建智能合约实例
	contract := new(SmartContract)

	// 准备测试数据
	cert := Certificate{
		CertID:       "CERT001",
		CarID:        "CAR001",
		CertType:     "REGISTRATION",
		FileHash:     "abcdef1234567890",
		FileLocation: "data/certificates/CAR001/CERT001.pdf",
		UploadTime:   time.Now(),
	}
	certJSON, _ := json.Marshal(cert)

	// 设置模拟行为
	mockStub.On("CreateCompositeKey", CERTIFICATE, []string{"CERT001"}).Return("CERTIFICATE_CERT001", nil)
	mockStub.On("GetState", "CERTIFICATE_CERT001").Return(certJSON, nil)

	// 执行测试
	result, err := contract.GetCertificate(mockCtx, "CERT001")

	// 验证结果
	assert.NoError(t, err)
	assert.Equal(t, cert.CertID, result.CertID)
	assert.Equal(t, cert.CarID, result.CarID)
	assert.Equal(t, cert.CertType, result.CertType)
	assert.Equal(t, cert.FileHash, result.FileHash)
	assert.Equal(t, cert.FileLocation, result.FileLocation)
	mockStub.AssertExpectations(t)
}
