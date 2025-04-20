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
	carDealerHandler := api.NewCarDealerHandler()
	tradingPlatformHandler := api.NewTradingPlatformHandler()
	bankHandler := api.NewBankHandler()

	// 汽车经销商的接口
	car := apiGroup.Group("/car-dealer")
	{
		// 创建汽车信息
		car.POST("/car/create", carDealerHandler.CreateCar)
		// 查询汽车接口
		car.GET("/car/:id", carDealerHandler.QueryCar)
		car.GET("/car/list", carDealerHandler.QueryCarList)
		// 证书接口 (修改路径以避免冲突)
		car.POST("/certificates/:carId", carDealerHandler.UploadCertificate)                              // 上传证书
		car.GET("/certificates/:carId", carDealerHandler.ListCertificates)                                // 获取证书列表
		car.GET("/certificates/verify/:certId", carDealerHandler.VerifyCertificateHandler)                // 验证证书 (修改路径)
		car.POST("/certificates/verify-upload/:carId", carDealerHandler.VerifyUploadedCertificateHandler) // 上传文件进行验证 (新增)
		// 查询区块接口
		car.GET("/block/list", carDealerHandler.QueryBlockList)
	}

	// 交易平台的接口
	trading := apiGroup.Group("/trading-platform")
	{
		// 生成交易
		trading.POST("/transaction/create", tradingPlatformHandler.CreateTransaction)
		// 查询汽车接口
		trading.GET("/car/:id", tradingPlatformHandler.QueryCar)
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

	// 配置静态文件服务 (新增)
	// 将 URL 路径 /api/files/ 映射到服务器本地的 ./data/ 目录
	// 例如: 访问 /api/files/certificates/car1/cert1.pdf 会读取 ./data/certificates/car1/cert1.pdf
	r.Static("/api/files", "./data")

	// 启动服务器
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.Port)
	if err := r.Run(addr); err != nil {
		log.Fatalf("启动服务器失败：%v", err)
	}
}
