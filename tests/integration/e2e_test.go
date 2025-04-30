package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// 测试配置
const (
	BaseURL = "http://localhost:8888/api" // 后端服务地址
)

// 测试数据
var (
	carID        string
	txID         string
	certID       string
	fileLocation string
)

// 测试客户端
type TestClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// 创建新的测试客户端
func NewTestClient(baseURL string) *TestClient {
	return &TestClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// 发送 GET 请求
func (c *TestClient) Get(path string, result interface{}) error {
	resp, err := c.HTTPClient.Get(c.BaseURL + path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// 发送 POST 请求
func (c *TestClient) Post(path string, body interface{}, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+path, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}

// 测试端到端流程
func TestE2EFlow(t *testing.T) {
	// 跳过测试，除非明确指定要运行集成测试
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration tests. Set RUN_INTEGRATION_TESTS=true to run.")
	}

	// 创建测试客户端
	client := NewTestClient(BaseURL)

	// 1. 汽车经销商创建汽车
	t.Run("1. 创建汽车", func(t *testing.T) {
		// 准备请求数据
		carID = fmt.Sprintf("CAR%d", time.Now().Unix())
		createCarRequest := map[string]interface{}{
			"id":    carID,
			"model": "特斯拉 Model 3",
			"vin":   "VIN123456789ABCDE",
			"owner": "Alice",
		}

		// 发送请求
		var response map[string]interface{}
		err := client.Post("/car-dealer/car/create", createCarRequest, &response)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, "汽车信息创建成功", response["message"])
	})

	// 2. 汽车经销商查询汽车
	t.Run("2. 查询汽车", func(t *testing.T) {
		// 发送请求
		var car map[string]interface{}
		err := client.Get("/car-dealer/car/"+carID, &car)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, carID, car["id"])
		assert.Equal(t, "特斯拉 Model 3", car["model"])
		assert.Equal(t, "VIN123456789ABCDE", car["vin"])
		assert.Equal(t, "Alice", car["currentOwner"])
		assert.Equal(t, "AVAILABLE", car["status"])
	})

	// 3. 汽车经销商上传证书
	t.Run("3. 上传证书", func(t *testing.T) {
		// 准备请求数据
		fileLocation = fmt.Sprintf("data/certificates/%s/cert.pdf", carID)
		uploadCertRequest := map[string]interface{}{
			"certType":     "REGISTRATION",
			"fileHash":     "abcdef1234567890",
			"fileLocation": fileLocation,
		}

		// 发送请求
		var response map[string]interface{}
		err := client.Post("/car-dealer/certificates/"+carID, uploadCertRequest, &response)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, "证书上传成功", response["message"])
	})

	// 4. 汽车经销商查询证书列表
	t.Run("4. 查询证书列表", func(t *testing.T) {
		// 发送请求
		var certificates []map[string]interface{}
		err := client.Get("/car-dealer/certificates/"+carID, &certificates)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, 1, len(certificates))
		assert.Equal(t, carID, certificates[0]["carId"])
		assert.Equal(t, "REGISTRATION", certificates[0]["certType"])
		assert.Equal(t, "abcdef1234567890", certificates[0]["fileHash"])
		assert.Equal(t, fileLocation, certificates[0]["fileLocation"])

		// 保存证书 ID 用于后续测试
		certID = certificates[0]["certId"].(string)
	})

	// 5. 交易平台创建交易
	t.Run("5. 创建交易", func(t *testing.T) {
		// 准备请求数据
		txID = fmt.Sprintf("TX%d", time.Now().Unix())
		createTxRequest := map[string]interface{}{
			"id":     txID,
			"carId":  carID,
			"seller": "Alice",
			"buyer":  "Bob",
			"price":  100000.0,
		}

		// 发送请求
		var response map[string]interface{}
		err := client.Post("/trading-platform/transaction/create", createTxRequest, &response)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, "交易创建成功", response["message"])
	})

	// 6. 交易平台查询交易
	t.Run("6. 查询交易", func(t *testing.T) {
		// 发送请求
		var transaction map[string]interface{}
		err := client.Get("/trading-platform/transaction/"+txID, &transaction)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, txID, transaction["id"])
		assert.Equal(t, carID, transaction["carId"])
		assert.Equal(t, "Alice", transaction["seller"])
		assert.Equal(t, "Bob", transaction["buyer"])
		assert.Equal(t, float64(100000), transaction["price"])
		assert.Equal(t, "PENDING", transaction["status"])
	})

	// 7. 交易平台查询汽车状态
	t.Run("7. 查询汽车状态", func(t *testing.T) {
		// 发送请求
		var car map[string]interface{}
		err := client.Get("/trading-platform/car/"+carID, &car)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, carID, car["id"])
		assert.Equal(t, "IN_TRANSACTION", car["status"])
	})

	// 8. 银行完成交易
	t.Run("8. 完成交易", func(t *testing.T) {
		// 发送请求
		var response map[string]interface{}
		err := client.Post("/bank/transaction/complete/"+txID, nil, &response)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, "交易完成成功", response["message"])
	})

	// 9. 银行查询交易状态
	t.Run("9. 查询交易状态", func(t *testing.T) {
		// 发送请求
		var transaction map[string]interface{}
		err := client.Get("/bank/transaction/"+txID, &transaction)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, txID, transaction["id"])
		assert.Equal(t, "COMPLETED", transaction["status"])
	})

	// 10. 汽车经销商查询汽车状态
	t.Run("10. 查询汽车状态", func(t *testing.T) {
		// 发送请求
		var car map[string]interface{}
		err := client.Get("/car-dealer/car/"+carID, &car)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, carID, car["id"])
		assert.Equal(t, "SOLD", car["status"])
		assert.Equal(t, "Bob", car["currentOwner"])
	})

	// 11. 汽车经销商验证证书
	t.Run("11. 验证证书", func(t *testing.T) {
		// 发送请求
		var result map[string]interface{}
		err := client.Get("/car-dealer/certificates/verify/"+certID, &result)

		// 验证结果
		assert.NoError(t, err)
		assert.Equal(t, true, result["match"])
		assert.Equal(t, "abcdef1234567890", result["storedHash"])
	})
}
