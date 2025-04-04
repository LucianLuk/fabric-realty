package api

import (
	"application/service"
	"application/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type VehicleAgencyHandler struct {
	vehicleService *service.VehicleAgencyService
}

func NewVehicleAgencyHandler() *VehicleAgencyHandler {
	return &VehicleAgencyHandler{
		vehicleService: &service.VehicleAgencyService{},
	}
}

// CreateVehicle 创建车辆信息（仅车辆管理机构组织可以调用）
func (h *VehicleAgencyHandler) CreateVehicle(c *gin.Context) {
	var req struct {
		ID        string  `json:"id"`
		Model     string  `json:"model"`
		Year      int     `json:"year"`
		Brand     string  `json:"brand"`
		Mileage   float64 `json:"mileage"`
		Condition string  `json:"condition"`
		Owner     string  `json:"owner"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "车辆信息格式错误")
		return
	}

	err := h.vehicleService.CreateVehicle(req.ID, req.Model, req.Year, req.Brand, req.Mileage, req.Condition, req.Owner)
	if err != nil {
		utils.ServerError(c, "创建车辆信息失败："+err.Error())
		return
	}

	utils.SuccessWithMessage(c, "车辆信息创建成功", nil)
}

// QueryVehicle 查询车辆信息
func (h *VehicleAgencyHandler) QueryVehicle(c *gin.Context) {
	id := c.Param("id")
	vehicle, err := h.vehicleService.QueryVehicle(id)
	if err != nil {
		utils.ServerError(c, "查询车辆信息失败："+err.Error())
		return
	}

	utils.Success(c, vehicle)
}

// QueryVehicleList 分页查询车辆列表
func (h *VehicleAgencyHandler) QueryVehicleList(c *gin.Context) {
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	bookmark := c.DefaultQuery("bookmark", "")
	status := c.DefaultQuery("status", "")

	result, err := h.vehicleService.QueryVehicleList(int32(pageSize), bookmark, status)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, result)
}

// QueryBlockList 分页查询区块列表
func (h *VehicleAgencyHandler) QueryBlockList(c *gin.Context) {
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	pageNum, _ := strconv.Atoi(c.DefaultQuery("pageNum", "1"))

	result, err := h.vehicleService.QueryBlockList(pageSize, pageNum)
	if err != nil {
		utils.ServerError(c, err.Error())
		return
	}

	utils.Success(c, result)
}
