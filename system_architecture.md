# 基于四维图新数据的CRP算法路径规划系统架构设计

## 1. 系统概述

本系统基于四维图新的高精度地图数据，采用Contraction Hierarchies（收缩层次）优化的路径规划算法（CRP），实现起终点间多条路线的快速计算，支持时间和距离两种权值类型，确保响应时间在100ms以内。

## 2. 核心技术栈

### 2.1 编程语言与框架
- **主服务**: Go 1.21+ (高并发、低延迟特性)
- **数据预处理**: C++ (高性能图算法处理)
- **Web服务**: Gin框架 (轻量级HTTP服务)
- **缓存**: Redis 7.0+ (热点路径缓存)
- **数据库**: PostgreSQL 15+ + PostGIS (地理空间数据)

### 2.2 关键算法
- **CRP (Contraction Hierarchies-based Route Planning)**
- **双向Dijkstra优化**
- **A*启发式搜索**
- **多路径Yen's K-shortest算法**

## 3. 系统架构

### 3.1 整体架构图

```
┌─────────────────────────────────────────────────────────────┐
│                    负载均衡层 (Nginx)                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                  API网关层 (Go-Gin)                         │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐           │
│  │ 路径规划API │ │ 坐标转换API │ │ 健康检查API │           │
│  └─────────────┘ └─────────────┘ └─────────────┘           │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                   核心计算层                                 │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │   CRP引擎       │ │  多路径计算器   │ │  权值计算器     ││
│  │ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ ││
│  │ │预处理图结构 │ │ │ │Yen's K算法  │ │ │ │时间权值模型 │ ││
│  │ │收缩层次结构 │ │ │ │路径去重优化 │ │ │ │距离权值模型 │ ││
│  │ │双向搜索优化 │ │ │ │相似度过滤   │ │ │ │实时路况权重 │ ││
│  │ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                   数据服务层                                 │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │   内存图存储    │ │   缓存服务      │ │  坐标转换服务   ││
│  │ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ ││
│  │ │邻接列表结构 │ │ │ │Redis集群    │ │ │ │GCJ02<->WGS84│ ││
│  │ │空间索引     │ │ │ │LRU淘汰策略  │ │ │ │高精度转换   │ ││
│  │ │分片存储     │ │ │ │热点路径缓存 │ │ │ │批量转换优化 │ ││
│  │ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                   持久化存储层                               │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │ PostgreSQL+PostGIS │ │   文件存储     │ │   监控存储      ││
│  │ ┌─────────────┐ │ │ ┌─────────────┐ │ │ ┌─────────────┐ ││
│  │ │四维图新数据 │ │ │ │预处理结果   │ │ │ │InfluxDB     │ ││
│  │ │空间索引优化 │ │ │ │静态路网文件 │ │ │ │性能指标     │ ││
│  │ │分区表设计   │ │ │ │配置文件     │ │ │ │错误日志     │ ││
│  │ └─────────────┘ │ │ └─────────────┘ │ │ └─────────────┘ ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### 3.2 核心模块设计

#### 3.2.1 CRP算法引擎
```go
type CRPEngine struct {
    // 收缩层次图结构
    contractedGraph *ContractedGraph
    // 原始图结构
    originalGraph   *Graph
    // 双向搜索优化器
    bidirectionalSearch *BidirectionalSearch
    // 预处理数据缓存
    preprocessCache sync.Map
}

type ContractedGraph struct {
    // 节点层次信息
    nodeHierarchy map[NodeID]int
    // 收缩边集合
    shortcuts     map[EdgeID]*Shortcut
    // 分层邻接表
    adjList       [][]Edge
}
```

#### 3.2.2 多路径计算器
```go
type MultiPathCalculator struct {
    crpEngine    *CRPEngine
    pathFilter   *PathFilter
    diversityCalculator *DiversityCalculator
}

type PathResult struct {
    Paths        []Path      `json:"paths"`
    TotalTime    int64       `json:"total_time_ms"`
    WeightType   WeightType  `json:"weight_type"`
    Coordinates  []Point     `json:"coordinates"`
}
```

## 4. 数据结构设计

### 4.1 四维图新数据适配

```go
// 四维图新路网节点
type NaviNode struct {
    ID          int64   `json:"node_id"`
    Longitude   float64 `json:"longitude"`  // GCJ02经度
    Latitude    float64 `json:"latitude"`   // GCJ02纬度
    Elevation   float32 `json:"elevation"`
    Type        int     `json:"node_type"`  // 1:路口 2:形状点 3:收费站
}

// 四维图新路网边
type NaviEdge struct {
    ID          int64   `json:"edge_id"`
    StartNodeID int64   `json:"start_node_id"`
    EndNodeID   int64   `json:"end_node_id"`
    Length      float32 `json:"length"`        // 米
    Time        int32   `json:"time"`          // 秒
    Speed       int16   `json:"speed"`         // km/h
    RoadLevel   int8    `json:"road_level"`    // 道路等级
    Direction   int8    `json:"direction"`     // 0:双向 1:正向 2:反向
    Geometry    []Point `json:"geometry"`      // 几何形状点
}

// 内存优化的图结构
type Graph struct {
    Nodes       []Node      `json:"nodes"`
    Edges       []Edge      `json:"edges"`
    AdjList     [][]int32   `json:"adj_list"`     // 邻接表
    SpatialIndex *SpatialIndex `json:"-"`         // 空间索引
}
```

### 4.2 权值计算模型

```go
type WeightCalculator interface {
    CalculateWeight(edge *Edge, context *RoutingContext) float64
}

// 时间权值计算器
type TimeWeightCalculator struct {
    trafficData   *TrafficData
    speedProfiles map[int]SpeedProfile
}

// 距离权值计算器
type DistanceWeightCalculator struct {
    terrainFactor map[int]float64
}

type RoutingContext struct {
    StartTime   time.Time
    WeightType  WeightType
    VehicleType VehicleType
    AvoidTypes  []int
}
```

## 5. 性能优化策略

### 5.1 算法优化

#### 5.1.1 预处理优化
```cpp
// C++实现的图预处理模块
class GraphPreprocessor {
public:
    // 收缩层次预处理
    void preprocessContractionHierarchies(const Graph& graph);
    
    // 重要性计算
    double calculateNodeImportance(NodeID node);
    
    // 边收缩
    void contractNode(NodeID node);
    
    // 快捷方式生成
    std::vector<Shortcut> generateShortcuts(NodeID node);
};
```

#### 5.1.2 搜索优化
```go
// 双向A*搜索优化
type BidirectionalAStarSearch struct {
    forwardQueue   *PriorityQueue
    backwardQueue  *PriorityQueue
    meetingPoint   NodeID
    heuristic      HeuristicFunction
}

func (search *BidirectionalAStarSearch) FindPath(
    start, target NodeID, 
    weightType WeightType,
) (*Path, error) {
    // 实现双向搜索逻辑
    // 使用收缩层次剪枝
    // A*启发式函数优化
}
```

### 5.2 内存优化

#### 5.2.1 图数据压缩
```go
// 紧凑型边结构 (16字节对齐)
type CompactEdge struct {
    Target   uint32  // 4字节 - 目标节点ID
    Weight   uint32  // 4字节 - 权值(定点数表示)
    Length   uint16  // 2字节 - 长度(米，最大65km)
    Time     uint16  // 2字节 - 时间(秒，最大18小时)
    RoadType uint8   // 1字节 - 道路类型
    Flags    uint8   // 1字节 - 标志位
    _        [2]byte // 2字节 - 对齐填充
}

// 内存池管理
type MemoryPool struct {
    nodePool *sync.Pool
    edgePool *sync.Pool
    pathPool *sync.Pool
}
```

#### 5.2.2 空间索引优化
```go
// 分层R-Tree索引
type HierarchicalRTree struct {
    levels    []RTreeLevel
    leafSize  int
    maxLevels int
}

// 地理编码缓存
type GeocodingCache struct {
    lru       *lru.Cache
    precision int // 精度控制(米)
}
```

### 5.3 缓存策略

#### 5.3.1 多级缓存架构
```go
type CacheManager struct {
    // L1: 进程内缓存 (LRU)
    l1Cache *ristretto.Cache
    
    // L2: Redis集群缓存
    l2Cache *redis.ClusterClient
    
    // L3: 预计算路径库
    pathLibrary *PathLibrary
}

// 缓存键生成策略
func generateCacheKey(start, end Point, weightType WeightType) string {
    // 使用Geohash进行空间降维
    startHash := geohash.Encode(start.Lat, start.Lng, 12)
    endHash := geohash.Encode(end.Lat, end.Lng, 12)
    return fmt.Sprintf("route:%s:%s:%d", startHash, endHash, weightType)
}
```

#### 5.3.2 预计算策略
```go
// 热点路径预计算
type HotspotPrecomputation struct {
    analyzer    *TrafficAnalyzer
    scheduler   *cron.Cron
    pathBuilder *PathBuilder
}

func (h *HotspotPrecomputation) precomputeHotspots() {
    // 分析历史查询热点
    hotspots := h.analyzer.getHotspots()
    
    // 预计算热点间路径
    for _, pair := range hotspots {
        paths := h.pathBuilder.calculateMultiplePaths(pair.Start, pair.End)
        h.cacheManager.store(pair, paths)
    }
}
```

## 6. API接口设计

### 6.1 路径规划接口

```go
// 多路径规划请求
type MultiRouteRequest struct {
    Start      Point      `json:"start" binding:"required"`
    End        Point      `json:"end" binding:"required"`
    WeightType WeightType `json:"weight_type" binding:"required"`
    MaxRoutes  int        `json:"max_routes,omitempty"`
    Options    RouteOptions `json:"options,omitempty"`
}

type Point struct {
    Longitude float64 `json:"longitude" binding:"required,min=-180,max=180"`
    Latitude  float64 `json:"latitude" binding:"required,min=-90,max=90"`
}

type RouteOptions struct {
    AvoidTolls    bool   `json:"avoid_tolls,omitempty"`
    AvoidHighways bool   `json:"avoid_highways,omitempty"`
    VehicleType   string `json:"vehicle_type,omitempty"`
    DepartureTime string `json:"departure_time,omitempty"`
}

// 响应结构
type MultiRouteResponse struct {
    Code    int           `json:"code"`
    Message string        `json:"message"`
    Data    *RouteData    `json:"data,omitempty"`
    Timing  *TimingInfo   `json:"timing"`
}

type RouteData struct {
    Routes []Route `json:"routes"`
    Summary RouteSummary `json:"summary"`
}

type Route struct {
    Distance    int64     `json:"distance"`      // 米
    Duration    int64     `json:"duration"`      // 秒
    Geometry    string    `json:"geometry"`      // 编码几何
    Instructions []Instruction `json:"instructions"`
    Waypoints   []Point   `json:"waypoints"`
}

type TimingInfo struct {
    QueryTime     int64 `json:"query_time_ms"`
    CacheHit      bool  `json:"cache_hit"`
    ProcessingTime int64 `json:"processing_time_ms"`
}
```

### 6.2 接口实现

```go
// 路径规划控制器
func (c *RouteController) calculateMultipleRoutes(ctx *gin.Context) {
    var req MultiRouteRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }
    
    start := time.Now()
    
    // 坐标系转换 (GCJ02 -> WGS84用于内部计算)
    internalStart := c.coordConverter.GCJ02ToWGS84(req.Start)
    internalEnd := c.coordConverter.GCJ02ToWGS84(req.End)
    
    // 检查缓存
    cacheKey := generateCacheKey(req.Start, req.End, req.WeightType)
    if cached, found := c.cacheManager.Get(cacheKey); found {
        ctx.JSON(200, cached)
        return
    }
    
    // 多路径计算
    routes, err := c.pathCalculator.CalculateMultiplePaths(
        internalStart, internalEnd, req.WeightType, req.MaxRoutes)
    if err != nil {
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }
    
    // 结果转换和组装
    response := c.buildResponse(routes, req)
    response.Timing = &TimingInfo{
        QueryTime:      time.Since(start).Milliseconds(),
        ProcessingTime: routes.ProcessingTime,
        CacheHit:       false,
    }
    
    // 异步缓存结果
    go c.cacheManager.SetAsync(cacheKey, response)
    
    ctx.JSON(200, response)
}
```

## 7. 部署架构

### 7.1 容器化部署

```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 go build -a -ldflags '-extldflags "-static"' -o route-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/
COPY --from=builder /app/route-service .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/data ./data
EXPOSE 8080
CMD ["./route-service"]
```

```yaml
# docker-compose.yml
version: '3.8'
services:
  route-service:
    build: .
    ports:
      - "8080:8080"
    environment:
      - REDIS_URL=redis://redis:6379
      - DB_URL=postgres://user:pass@postgres:5432/routes
    depends_on:
      - redis
      - postgres
    deploy:
      replicas: 4
      resources:
        limits:
          memory: 2G
          cpus: '1.0'
  
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
  
  postgres:
    image: postgis/postgis:15-3.3
    environment:
      POSTGRES_DB: routes
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./sql:/docker-entrypoint-initdb.d
```

### 7.2 Kubernetes部署

```yaml
# k8s-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: route-service
spec:
  replicas: 6
  selector:
    matchLabels:
      app: route-service
  template:
    metadata:
      labels:
        app: route-service
    spec:
      containers:
      - name: route-service
        image: route-service:latest
        ports:
        - containerPort: 8080
        env:
        - name: REDIS_URL
          value: "redis://redis-service:6379"
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1000m"
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: route-service
spec:
  selector:
    app: route-service
  ports:
    - protocol: TCP
      port: 80
      targetPort: 8080
  type: LoadBalancer
```

## 8. 监控和运维

### 8.1 性能监控

```go
// 性能指标收集
type MetricsCollector struct {
    registry prometheus.Registry
    
    // 请求指标
    requestDuration *prometheus.HistogramVec
    requestTotal    *prometheus.CounterVec
    
    // 算法性能指标
    algorithmDuration *prometheus.HistogramVec
    cacheHitRate     *prometheus.GaugeVec
    
    // 资源使用指标
    memoryUsage     prometheus.Gauge
    cpuUsage        prometheus.Gauge
}

func (m *MetricsCollector) recordRequest(duration time.Duration, status string) {
    m.requestDuration.WithLabelValues(status).Observe(duration.Seconds())
    m.requestTotal.WithLabelValues(status).Inc()
}
```

### 8.2 日志系统

```go
// 结构化日志
type RouteLogger struct {
    logger *logrus.Logger
}

func (l *RouteLogger) LogRouteRequest(req *MultiRouteRequest, result *RouteResult) {
    l.logger.WithFields(logrus.Fields{
        "start_lat":      req.Start.Latitude,
        "start_lng":      req.Start.Longitude,
        "end_lat":        req.End.Latitude,
        "end_lng":        req.End.Longitude,
        "weight_type":    req.WeightType,
        "route_count":    len(result.Routes),
        "processing_time": result.ProcessingTime,
        "cache_hit":      result.CacheHit,
    }).Info("Route calculation completed")
}
```

## 9. 性能指标和SLA

### 9.1 性能目标
- **响应时间**: P99 < 100ms, P95 < 50ms
- **吞吐量**: 10,000 QPS+
- **可用性**: 99.9%
- **数据准确性**: 与四维图新官方结果偏差 < 5%

### 9.2 容量规划
- **内存**: 每实例8GB (预加载核心路网数据)
- **CPU**: 每实例4核 (并行计算优化)
- **存储**: PostgreSQL 1TB+ SSD
- **缓存**: Redis集群 64GB+

## 10. 安全和稳定性

### 10.1 限流和熔断
```go
// 令牌桶限流
type RateLimiter struct {
    limiter *rate.Limiter
    burst   int
}

// 熔断器
type CircuitBreaker struct {
    state       int32
    failures    int32
    lastFailure time.Time
    threshold   int32
}
```

### 10.2 灾备方案
- **多区域部署**: 支持跨区域故障切换
- **数据备份**: 实时数据同步和定期备份
- **降级策略**: 算法降级为简单最短路径
- **监控告警**: 实时监控关键指标，异常自动告警

这个架构设计确保了系统在满足100ms响应时间要求的同时，提供高可用性、高并发处理能力和准确的路径规划结果。