package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	
	"route-planning/internal/config"
	"route-planning/internal/handler"
	"route-planning/internal/service"
	"route-planning/internal/cache"
	"route-planning/internal/database"
	"route-planning/pkg/crp"
	"route-planning/pkg/middleware"
	"route-planning/pkg/monitor"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// 初始化监控
	metrics := monitor.NewMetricsCollector()
	metrics.Init()

	// 初始化数据库连接
	db, err := database.NewPostgreSQL(cfg.Database)
	if err != nil {
		logger.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 初始化缓存
	cacheManager, err := cache.NewManager(cfg.Cache)
	if err != nil {
		logger.Fatalf("Failed to init cache: %v", err)
	}

	// 初始化CRP引擎
	crpEngine, err := crp.NewEngine(cfg.Graph)
	if err != nil {
		logger.Fatalf("Failed to init CRP engine: %v", err)
	}

	// 初始化服务层
	routeService := service.NewRouteService(
		crpEngine,
		cacheManager,
		db,
		logger,
		metrics,
	)

	// 初始化处理器
	routeHandler := handler.NewRouteHandler(routeService, logger, metrics)

	// 设置Gin模式
	if cfg.Server.Mode == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	router := setupRouter(routeHandler, metrics)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// 优雅启动
	go func() {
		logger.Infof("Starting server on %s", cfg.Server.Address)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited")
}

func setupRouter(routeHandler *handler.RouteHandler, metrics *monitor.MetricsCollector) *gin.Engine {
	router := gin.New()

	// 中间件
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.Metrics(metrics))
	router.Use(middleware.RateLimit())

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "timestamp": time.Now().Unix()})
	})

	// API路由组
	v1 := router.Group("/api/v1")
	{
		// 路径规划接口
		v1.POST("/route/multiple", routeHandler.CalculateMultipleRoutes)
		v1.POST("/route/single", routeHandler.CalculateSingleRoute)
		
		// 坐标转换接口
		v1.POST("/coord/convert", routeHandler.ConvertCoordinates)
		
		// 系统状态接口
		v1.GET("/status", routeHandler.GetSystemStatus)
	}

	// 监控指标接口
	router.GET("/metrics", gin.WrapH(metrics.Handler()))

	return router
}