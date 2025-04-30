package server_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 模拟服务层
type MockCarDealerService struct {
	mock.Mock
}

func (m *MockCarDealerService) CreateCar(id, model, vin, owner string) error {
	args := m.Called(id, model, vin, owner)
	return args.Error(0)
}

func (m *MockCarDealerService) QueryCar(id string) (map[string]interface{}, error) {
	args := m.Called(id)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCarDealerService) QueryCarList(pageSize int32, bookmark string, status string) (map[string]interface{}, error) {
	args := m.Called(pageSize, bookmark, status)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockCarDealerService) QueryBlockList(pageSize int, pageNum int) (interface{}, error) {
	args := m.Called(pageSize, pageNum)
	return args.Get(0), args.Error(1)
}

// 模拟交易平台服务
type MockTradingPlatformService struct {
	mock.Mock
}

func (m *MockTradingPlatformService) CreateTransaction(txID, carID, seller, buyer string, price float64) error {
	args := m.Called(txID, carID, seller, buyer, price)
	return args.Error(0)
}

func (m *MockTradingPlatformService) QueryCar(id string) (map[string]interface{}, error) {
	args := m.Called(id)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockTradingPlatformService) QueryTransaction(txID string) (map[string]interface{}, error) {
	args := m.Called(txID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockTradingPlatformService) QueryTransactionList(pageSize int32, bookmark string, status string) (map[string]interface{}, error) {
	args := m.Called(pageSize, bookmark, status)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockTradingPlatformService) QueryBlockList(pageSize int, pageNum int) (interface{}, error) {
	args := m.Called(pageSize, pageNum)
	return args.Get(0), args.Error(1)
}

// 模拟银行服务
type MockBankService struct {
	mock.Mock
}

func (m *MockBankService) CompleteTransaction(txID string) error {
	args := m.Called(txID)
	return args.Error(0)
}

func (m *MockBankService) QueryTransaction(txID string) (map[string]interface{}, error) {
	args := m.Called(txID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockBankService) QueryTransactionList(pageSize int32, bookmark string, status string) (map[string]interface{}, error) {
	args := m.Called(pageSize, bookmark, status)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockBankService) QueryBlockList(pageSize int, pageNum int) (interface{}, error) {
	args := m.Called(pageSize, pageNum)
	return args.Get(0), args.Error(1)
}

// 模拟证书服务
type MockCertificateService struct {
	mock.Mock
}

func (m *MockCertificateService) UploadCertificate(carID, certType, fileHash, fileLocation string) error {
	args := m.Called(carID, certType, fileHash, fileLocation)
	return args.Error(0)
}

func (m *MockCertificateService) ListCertificates(carID string) ([]map[string]interface{}, error) {
	args := m.Called(carID)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func (m *MockCertificateService) VerifyCertificate(certID string) (map[string]interface{}, error) {
	args := m.Called(certID)
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// 测试 CarDealerHandler.CreateCar
func TestCreateCar(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockCarDealerService)
	mockService.On("CreateCar", "CAR001", "特斯拉 Model 3", "VIN123456789ABCDE", "Alice").Return(nil)

	// 创建请求体
	requestBody := map[string]interface{}{
		"id":    "CAR001",
		"model": "特斯拉 Model 3",
		"vin":   "VIN123456789ABCDE",
		"owner": "Alice",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// 创建请求
	req, _ := http.NewRequest("POST", "/api/car-dealer/car/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 路由
	r := gin.New()
	r.POST("/api/car-dealer/car/create", func(c *gin.Context) {
		var body struct {
			ID    string `json:"id"`
			Model string `json:"model"`
			VIN   string `json:"vin"`
			Owner string `json:"owner"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := mockService.CreateCar(body.ID, body.Model, body.VIN, body.Owner)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "汽车信息创建成功"})
	})

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "汽车信息创建成功", response["message"])
	mockService.AssertExpectations(t)
}

// 测试 CarDealerHandler.QueryCar
func TestQueryCar(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockCarDealerService)
	mockCar := map[string]interface{}{
		"id":           "CAR001",
		"model":        "特斯拉 Model 3",
		"vin":          "VIN123456789ABCDE",
		"currentOwner": "Alice",
		"status":       "AVAILABLE",
		"createTime":   time.Now().Format(time.RFC3339),
		"updateTime":   time.Now().Format(time.RFC3339),
	}
	mockService.On("QueryCar", "CAR001").Return(mockCar, nil)

	// 创建请求
	req, _ := http.NewRequest("GET", "/api/car-dealer/car/CAR001", nil)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 路由
	r := gin.New()
	r.GET("/api/car-dealer/car/:id", func(c *gin.Context) {
		id := c.Param("id")
		car, err := mockService.QueryCar(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, car)
	})

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "CAR001", response["id"])
	assert.Equal(t, "特斯拉 Model 3", response["model"])
	assert.Equal(t, "VIN123456789ABCDE", response["vin"])
	assert.Equal(t, "Alice", response["currentOwner"])
	assert.Equal(t, "AVAILABLE", response["status"])
	mockService.AssertExpectations(t)
}

// 测试 TradingPlatformHandler.CreateTransaction
func TestCreateTransaction(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockTradingPlatformService)
	mockService.On("CreateTransaction", "TX001", "CAR001", "Alice", "Bob", 100000.0).Return(nil)

	// 创建请求体
	requestBody := map[string]interface{}{
		"id":     "TX001",
		"carId":  "CAR001",
		"seller": "Alice",
		"buyer":  "Bob",
		"price":  100000.0,
	}
	jsonBody, _ := json.Marshal(requestBody)

	// 创建请求
	req, _ := http.NewRequest("POST", "/api/trading-platform/transaction/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 路由
	r := gin.New()
	r.POST("/api/trading-platform/transaction/create", func(c *gin.Context) {
		var body struct {
			ID     string  `json:"id"`
			CarID  string  `json:"carId"`
			Seller string  `json:"seller"`
			Buyer  string  `json:"buyer"`
			Price  float64 `json:"price"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := mockService.CreateTransaction(body.ID, body.CarID, body.Seller, body.Buyer, body.Price)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "交易创建成功"})
	})

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "交易创建成功", response["message"])
	mockService.AssertExpectations(t)
}

// 测试 BankHandler.CompleteTransaction
func TestCompleteTransaction(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockBankService)
	mockService.On("CompleteTransaction", "TX001").Return(nil)

	// 创建请求
	req, _ := http.NewRequest("POST", "/api/bank/transaction/complete/TX001", nil)

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 路由
	r := gin.New()
	r.POST("/api/bank/transaction/complete/:txId", func(c *gin.Context) {
		txID := c.Param("txId")
		err := mockService.CompleteTransaction(txID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "交易完成成功"})
	})

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "交易完成成功", response["message"])
	mockService.AssertExpectations(t)
}

// 测试 CarDealerHandler.UploadCertificate
func TestUploadCertificate(t *testing.T) {
	// 设置 Gin 为测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟服务
	mockService := new(MockCertificateService)
	mockService.On("UploadCertificate", "CAR001", "REGISTRATION", "abcdef1234567890", "data/certificates/CAR001/cert.pdf").Return(nil)

	// 创建请求体
	requestBody := map[string]interface{}{
		"certType":     "REGISTRATION",
		"fileHash":     "abcdef1234567890",
		"fileLocation": "data/certificates/CAR001/cert.pdf",
	}
	jsonBody, _ := json.Marshal(requestBody)

	// 创建请求
	req, _ := http.NewRequest("POST", "/api/car-dealer/certificates/CAR001", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 创建 Gin 路由
	r := gin.New()
	r.POST("/api/car-dealer/certificates/:carId", func(c *gin.Context) {
		carID := c.Param("carId")
		var body struct {
			CertType     string `json:"certType"`
			FileHash     string `json:"fileHash"`
			FileLocation string `json:"fileLocation"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := mockService.UploadCertificate(carID, body.CertType, body.FileHash, body.FileLocation)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "证书上传成功"})
	})

	// 执行请求
	r.ServeHTTP(w, req)

	// 验证结果
	assert.Equal(t, http.StatusOK, w.Code)
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "证书上传成功", response["message"])
	mockService.AssertExpectations(t)
}
