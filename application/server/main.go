package main

import (
	"application/api"
	"application/config"
	"application/pkg/fabric"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置
	if err := config.InitConfig(); err != nil {
		log.Fatalf("初始化配置失败：%v", err)
	}

	// 初始化 Fabric 客户端
	if err := fabric.InitFabric(); err != nil {
		log.Fatalf("初始化Fabric客户端失败：%v", err)
	}

	// 创建 Gin 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	apiGroup := r.Group("/api")

	// 注册路由
	vehicleAgencyHandler := api.NewVehicleAgencyHandler()
	tradingPlatformHandler := api.NewTradingPlatformHandler()
	bankHandler := api.NewBankHandler()

	// 车辆管理机构的接口
	vehicle := apiGroup.Group("/vehicle-agency")
	{
		// 创建车辆信息
		vehicle.POST("/vehicle/create", vehicleAgencyHandler.CreateVehicle)
		// 查询车辆接口
		vehicle.GET("/vehicle/:id", vehicleAgencyHandler.QueryVehicle)
		vehicle.GET("/vehicle/list", vehicleAgencyHandler.QueryVehicleList)
		// 查询区块接口
		vehicle.GET("/block/list", vehicleAgencyHandler.QueryBlockList)
	}

	// 交易平台的接口
	trading := apiGroup.Group("/trading-platform")
	{
		// 生成交易
		trading.POST("/transaction/create", tradingPlatformHandler.CreateTransaction)
		// 查询车辆接口
		trading.GET("/vehicle/:id", tradingPlatformHandler.QueryVehicle)
		// 查询交易接口
		trading.GET("/transaction/:txId", tradingPlatformHandler.QueryTransaction)
		trading.GET("/transaction/list", tradingPlatformHandler.QueryTransactionList)
		// 查询区块接口
		trading.GET("/block/list", tradingPlatformHandler.QueryBlockList)
	}

	// 银行的接口
	bank := apiGroup.Group("/bank")
	{
		// 完成交易
		bank.POST("/transaction/complete/:txId", bankHandler.CompleteTransaction)
		// 查询交易接口
		bank.GET("/transaction/:txId", bankHandler.QueryTransaction)
		bank.GET("/transaction/list", bankHandler.QueryTransactionList)
		// 查询区块接口
		bank.GET("/block/list", bankHandler.QueryBlockList)
	}

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败：%v", err)
	}
}
