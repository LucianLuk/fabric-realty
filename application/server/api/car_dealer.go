package api

import (
	"application/service"
	"application/utils"
	"strconv"
	"strings" // Import strings package

	"github.com/gin-gonic/gin"
)

type CarDealerHandler struct {
	carService         *service.CarDealerService
	certificateService *service.CertificateService // Add certificate service
}

func NewCarDealerHandler() *CarDealerHandler {
	return &CarDealerHandler{
		carService:         &service.CarDealerService{},
		certificateService: &service.CertificateService{}, // Initialize certificate service
	}
}

// CreateCar 创建汽车信息（仅汽车经销商组织可以调用）
func (h *CarDealerHandler) CreateCar(c *gin.Context) {
	var req struct {
		ID    string `json:"id"`    // 车辆唯一标识，例如车牌号
		Model string `json:"model"` // 车型
		VIN   string `json:"vin"`   // 车辆识别代号 (Vehicle Identification Number)
		Owner string `json:"owner"` // 车主
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "汽车信息格式错误")
		return
	}

	err := h.carService.CreateCar(req.ID, req.Model, req.VIN, req.Owner)
	if err != nil {
		utils.ServerError(c, "创建汽车信息失败："+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "汽车信息创建成功", nil)
}

// QueryCar 查询汽车信息
func (h *CarDealerHandler) QueryCar(c *gin.Context) {
	id := c.Param("id")
	car, err := h.carService.QueryCar(id)
	if err != nil {
		utils.ServerError(c, "查询汽车信息失败："+err.Error())
		return
	}

	utils.Success(c, car)
}

// QueryCarList 分页查询汽车列表
func (h *CarDealerHandler) QueryCarList(c *gin.Context) {
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	bookmark := c.DefaultQuery("bookmark", "")
	status := c.DefaultQuery("status", "") // 状态，例如 "待售", "已售"

	result, err := h.carService.QueryCarList(int32(pageSize), bookmark, status)
	if err != nil {
		utils.ServerError(c, "查询汽车列表失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}

// --- Certificate Handlers ---

// UploadCertificate 上传证书文件
func (h *CarDealerHandler) UploadCertificate(c *gin.Context) {
	carId := c.Param("carId")
	if carId == "" {
		utils.BadRequest(c, "缺少车辆ID (carId)")
		return
	}

	// Get certificate type from form data
	certType := c.PostForm("certType")
	if certType == "" {
		utils.BadRequest(c, "缺少证书类型 (certType)")
		return
	}

	// Get file from form data
	file, err := c.FormFile("certificateFile") // "certificateFile" is the expected form field name
	if err != nil {
		utils.BadRequest(c, "获取上传文件失败: "+err.Error())
		return
	}

	// Call the service to handle upload and chaincode interaction
	certPayload, err := h.certificateService.AddCertificate(carId, certType, file)
	if err != nil {
		utils.ServerError(c, "上传证书失败: "+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "证书上传成功", certPayload)
}

// ListCertificates 获取车辆关联的证书列表
func (h *CarDealerHandler) ListCertificates(c *gin.Context) {
	carId := c.Param("carId")
	if carId == "" {
		utils.BadRequest(c, "缺少车辆ID (carId)")
		return
	}

	certificates, err := h.certificateService.GetCertificatesByCar(carId)
	if err != nil {
		utils.ServerError(c, "查询证书列表失败: "+err.Error())
		return
	}

	utils.Success(c, certificates)
}

// VerifyCertificateHandler 验证证书文件完整性
func (h *CarDealerHandler) VerifyCertificateHandler(c *gin.Context) {
	certId := c.Param("certId")
	if certId == "" {
		utils.BadRequest(c, "缺少证书ID (certId)")
		return
	}

	match, storedHash, currentHash, err := h.certificateService.VerifyCertificate(certId)
	if err != nil {
		utils.ServerError(c, "验证证书失败: "+err.Error())
		return
	}

	if match {
		utils.SuccessWithMessage(c, "验证成功：文件哈希与链上记录一致", gin.H{
			"match":       true,
			"storedHash":  storedHash,
			"currentHash": currentHash,
		})
	} else {
		utils.SuccessWithMessage(c, "验证失败：文件哈希与链上记录不一致", gin.H{
			"match":       false,
			"storedHash":  storedHash,
			"currentHash": currentHash,
		})
	}
}

// VerifyUploadedCertificateHandler handles uploading a file for verification against the original certificate.
func (h *CarDealerHandler) VerifyUploadedCertificateHandler(c *gin.Context) {
	carId := c.Param("carId")
	if carId == "" {
		utils.BadRequest(c, "缺少车辆ID (carId)")
		return
	}

	// Get file from form data
	file, err := c.FormFile("verificationFile") // Expecting "verificationFile" form field
	if err != nil {
		utils.BadRequest(c, "获取待验证文件失败: "+err.Error())
		return
	}

	// Call the service to handle comparison
	match, storedHash, currentHash, err := h.certificateService.VerifyUploadedCertificate(carId, file)
	if err != nil {
		// Handle specific errors like "no original certificate" differently if needed
		if strings.Contains(err.Error(), "没有已上传的原始证书记录") {
			utils.NotFound(c, err.Error())
		} else {
			utils.ServerError(c, "验证上传文件失败: "+err.Error())
		}
		return
	}

	// Return result
	if match {
		utils.SuccessWithMessage(c, "验证成功：上传文件与原始证书哈希一致", gin.H{
			"match":       true,
			"storedHash":  storedHash,
			"currentHash": currentHash,
		})
	} else {
		utils.SuccessWithMessage(c, "验证失败：上传文件与原始证书哈希不一致", gin.H{
			"match":       false,
			"storedHash":  storedHash,
			"currentHash": currentHash,
		})
	}
}

// QueryBlockList 分页查询区块列表
func (h *CarDealerHandler) QueryBlockList(c *gin.Context) {
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	result, err := h.carService.QueryBlockList(pageSize, pageNum)
	if err != nil {
		utils.ServerError(c, "查询区块列表失败: "+err.Error())
		return
	}

	utils.Success(c, result)
}
