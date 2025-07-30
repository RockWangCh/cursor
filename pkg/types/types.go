package types

import "time"

// 节点ID类型
type NodeID int64

// 边ID类型
type EdgeID int64

// 权值类型
type WeightType int

const (
	WeightTypeDistance WeightType = iota
	WeightTypeTime
)

// 四维图新节点结构
type Node struct {
	ID        NodeID  `json:"node_id"`
	Longitude float64 `json:"longitude"`  // GCJ02经度
	Latitude  float64 `json:"latitude"`   // GCJ02纬度
	Elevation float32 `json:"elevation"`
	Type      int8    `json:"node_type"`  // 1:路口 2:形状点 3:收费站
}

// 四维图新边结构
type Edge struct {
	ID          EdgeID  `json:"edge_id"`
	From        NodeID  `json:"from_node_id"`
	To          NodeID  `json:"to_node_id"`
	Length      int32   `json:"length"`        // 米
	Time        int32   `json:"time"`          // 秒
	Speed       int16   `json:"speed"`         // km/h
	RoadLevel   int8    `json:"road_level"`    // 道路等级
	Direction   int8    `json:"direction"`     // 0:双向 1:正向 2:反向
	Geometry    []Point `json:"geometry"`      // 几何形状点
}

// 点结构
type Point struct {
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// 路径结构
type Path struct {
	Nodes       []NodeID `json:"nodes"`
	Weight      float64  `json:"weight"`
	Distance    int32    `json:"distance"`    // 米
	Time        int32    `json:"time"`        // 秒
	Geometry    string   `json:"geometry"`    // 编码几何
	Instructions []Instruction `json:"instructions"`
}

// 导航指令
type Instruction struct {
	Type        string  `json:"type"`        // turn_left, turn_right, straight, etc.
	Description string  `json:"description"` // 文字描述
	Distance    int32   `json:"distance"`    // 到下一个指令的距离
	Position    Point   `json:"position"`    // 指令位置
}

// 多路径结果
type MultiPathResult struct {
	Paths          []Path      `json:"paths"`
	StartNode      NodeID      `json:"start_node"`
	EndNode        NodeID      `json:"end_node"`
	WeightType     WeightType  `json:"weight_type"`
	ProcessingTime int64       `json:"processing_time_ms"`
	CacheHit       bool        `json:"cache_hit"`
}

// 路由请求
type RouteRequest struct {
	Start      Point      `json:"start" binding:"required"`
	End        Point      `json:"end" binding:"required"`
	WeightType WeightType `json:"weight_type" binding:"required"`
	MaxRoutes  int        `json:"max_routes,omitempty"`
	Options    RouteOptions `json:"options,omitempty"`
}

// 路由选项
type RouteOptions struct {
	AvoidTolls    bool      `json:"avoid_tolls,omitempty"`
	AvoidHighways bool      `json:"avoid_highways,omitempty"`
	VehicleType   string    `json:"vehicle_type,omitempty"`
	DepartureTime *time.Time `json:"departure_time,omitempty"`
}

// 路由响应
type RouteResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *RouteData   `json:"data,omitempty"`
	Timing  *TimingInfo  `json:"timing"`
}

// 路由数据
type RouteData struct {
	Routes  []Route      `json:"routes"`
	Summary RouteSummary `json:"summary"`
}

// 路由概要
type RouteSummary struct {
	TotalDistance int32 `json:"total_distance"`
	TotalTime     int32 `json:"total_time"`
	RouteCount    int   `json:"route_count"`
}

// 路由
type Route struct {
	Distance     int32         `json:"distance"`      // 米
	Duration     int32         `json:"duration"`      // 秒
	Geometry     string        `json:"geometry"`      // 编码几何
	Instructions []Instruction `json:"instructions"`
	Waypoints    []Point       `json:"waypoints"`
}

// 时间信息
type TimingInfo struct {
	QueryTime      int64 `json:"query_time_ms"`
	CacheHit       bool  `json:"cache_hit"`
	ProcessingTime int64 `json:"processing_time_ms"`
}

// 坐标转换请求
type CoordConvertRequest struct {
	Points    []Point    `json:"points" binding:"required"`
	FromCRS   string     `json:"from_crs" binding:"required"`  // GCJ02, WGS84, BD09
	ToCRS     string     `json:"to_crs" binding:"required"`
}

// 坐标转换响应
type CoordConvertResponse struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Data    []Point `json:"data,omitempty"`
}

// 系统状态
type SystemStatus struct {
	Version        string            `json:"version"`
	Uptime         int64             `json:"uptime_seconds"`
	MemoryUsage    MemoryUsage       `json:"memory_usage"`
	GraphStats     GraphStats        `json:"graph_stats"`
	CacheStats     CacheStats        `json:"cache_stats"`
	Performance    PerformanceStats  `json:"performance"`
}

// 内存使用
type MemoryUsage struct {
	Allocated    uint64 `json:"allocated_mb"`
	TotalAlloc   uint64 `json:"total_alloc_mb"`
	SystemMemory uint64 `json:"system_memory_mb"`
	GCCount      uint32 `json:"gc_count"`
}

// 图统计
type GraphStats struct {
	NodeCount     int64 `json:"node_count"`
	EdgeCount     int64 `json:"edge_count"`
	ShortcutCount int64 `json:"shortcut_count"`
	IndexSize     int64 `json:"index_size_mb"`
}

// 缓存统计
type CacheStats struct {
	L1Size    int64   `json:"l1_size"`
	L1HitRate float64 `json:"l1_hit_rate"`
	L2Size    int64   `json:"l2_size"`
	L2HitRate float64 `json:"l2_hit_rate"`
}

// 性能统计
type PerformanceStats struct {
	TotalQueries     int64   `json:"total_queries"`
	QueriesPerSecond float64 `json:"queries_per_second"`
	AvgResponseTime  float64 `json:"avg_response_time_ms"`
	P95ResponseTime  float64 `json:"p95_response_time_ms"`
	P99ResponseTime  float64 `json:"p99_response_time_ms"`
}