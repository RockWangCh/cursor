# 基于CRP算法的高性能路径规划系统架构设计

## 1. 系统概述

### 1.1 背景与目标
- **数据源**：四维图新地图数据
- **算法**：CRP (Customizable Route Planning) 算法
- **性能要求**：响应时间 < 100ms
- **功能要求**：支持多条路线计算，支持时间/距离权值
- **坐标系**：GCJ02经纬度

### 1.2 CRP算法简介
CRP算法是一种两阶段的路径规划算法：
- **预处理阶段**：构建多级分区和边界图
- **查询阶段**：利用预处理数据快速计算路径

## 2. 系统架构设计

### 2.1 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                        客户端应用层                           │
├─────────────────────────────────────────────────────────────┤
│                    API Gateway / 负载均衡                     │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  路径规划   │  │   地图匹配  │  │   POI搜索   │        │
│  │   服务集群  │  │    服务     │  │    服务     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │ CRP查询引擎 │  │  预处理引擎 │  │  数据更新   │        │
│  │  (内存计算) │  │ (离线处理)  │  │    服务     │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
├─────────────────────────────────────────────────────────────┤
│                      数据存储层                               │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │  Redis集群  │  │  分布式文件 │  │  PostgreSQL │        │
│  │ (热点数据)  │  │系统(预处理) │  │ (原始数据)  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└─────────────────────────────────────────────────────────────┘
```

### 2.2 核心组件设计

#### 2.2.1 CRP查询引擎
```
功能：实时路径计算
技术栈：C++ / Rust
内存结构：
- Overlay Graph (覆盖图)
- Cell Index (分区索引)
- Metric Customization (度量定制)
- Path Cache (路径缓存)
```

#### 2.2.2 预处理引擎
```
功能：离线构建CRP数据结构
处理流程：
1. 道路网络分区 (Multi-level Partitioning)
2. 边界节点识别 (Boundary Node Detection)
3. 覆盖图构建 (Overlay Graph Construction)
4. 度量预计算 (Metric Preprocessing)
```

#### 2.2.3 数据更新服务
```
功能：增量更新路网数据
更新策略：
- 定期全量更新
- 实时增量更新
- 版本控制管理
```

## 3. 数据模型设计

### 3.1 四维图新数据适配

```cpp
// 道路节点
struct RoadNode {
    uint64_t node_id;
    double longitude;  // GCJ02
    double latitude;   // GCJ02
    uint16_t level;    // 道路等级
    vector<uint64_t> out_edges;
    vector<uint64_t> in_edges;
};

// 道路边
struct RoadEdge {
    uint64_t edge_id;
    uint64_t from_node;
    uint64_t to_node;
    float length;      // 米
    float travel_time; // 秒
    uint8_t road_class;
    uint8_t direction;
    vector<pair<double, double>> geometry; // 形状点
};

// CRP分区
struct CRPCell {
    uint32_t cell_id;
    uint8_t level;
    vector<uint64_t> nodes;
    vector<uint64_t> boundary_nodes;
    unordered_map<uint64_t, uint64_t> overlay_edges;
};
```

### 3.2 预处理数据结构

```cpp
// 覆盖图
struct OverlayGraph {
    unordered_map<uint64_t, vector<OverlayEdge>> adjacency_list;
    unordered_map<uint64_t, CellInfo> node_to_cell;
};

// 度量定制
struct MetricCustomization {
    enum MetricType { TIME, DISTANCE };
    unordered_map<uint64_t, float> edge_weights;
    unordered_map<pair<uint64_t, uint64_t>, float> shortcut_weights;
};
```

## 4. 算法实现细节

### 4.1 CRP查询流程

```cpp
class CRPRouter {
public:
    struct Route {
        vector<uint64_t> node_sequence;
        float total_cost;
        MetricType metric;
    };
    
    vector<Route> findRoutes(
        double start_lng, double start_lat,
        double end_lng, double end_lat,
        MetricType metric,
        int k_routes = 3
    ) {
        // 1. 坐标转换与地图匹配
        auto start_node = mapMatching(start_lng, start_lat);
        auto end_node = mapMatching(end_lng, end_lat);
        
        // 2. 确定查询分区
        auto start_cell = getCellForNode(start_node);
        auto end_cell = getCellForNode(end_node);
        
        // 3. 执行CRP查询
        if (start_cell == end_cell) {
            // 同分区查询
            return localSearch(start_node, end_node, metric, k_routes);
        } else {
            // 跨分区查询
            return overlaySearch(start_node, end_node, metric, k_routes);
        }
    }
    
private:
    vector<Route> overlaySearch(
        uint64_t start, uint64_t end, 
        MetricType metric, int k
    ) {
        // 1. 找到边界节点
        auto start_boundaries = getBoundaryNodes(start);
        auto end_boundaries = getBoundaryNodes(end);
        
        // 2. 计算到边界的局部路径
        auto start_costs = computeLocalCosts(start, start_boundaries);
        auto end_costs = computeLocalCosts(end_boundaries, end);
        
        // 3. 在覆盖图上搜索
        priority_queue<PathState> pq;
        vector<Route> results;
        
        // Yen's k-shortest paths algorithm
        for (auto& sb : start_boundaries) {
            pq.push({sb, 0, start_costs[sb]});
        }
        
        while (!pq.empty() && results.size() < k) {
            auto state = pq.top();
            pq.pop();
            
            if (end_boundaries.count(state.node)) {
                // 构建完整路径
                auto route = constructRoute(
                    start, state.node, end, 
                    state.path, end_costs[state.node]
                );
                results.push_back(route);
            }
            
            // 扩展邻居
            for (auto& edge : overlay_graph[state.node]) {
                float new_cost = state.cost + getEdgeCost(edge, metric);
                pq.push({edge.to, state.path + edge.to, new_cost});
            }
        }
        
        return results;
    }
};
```

### 4.2 性能优化策略

#### 4.2.1 内存优化
```cpp
// 使用内存池减少分配开销
template<typename T>
class MemoryPool {
    vector<T*> chunks;
    size_t chunk_size;
    T* current;
    size_t remaining;
};

// 压缩存储
struct CompactEdge {
    uint32_t to_node : 24;    // 24位节点ID
    uint32_t weight : 8;      // 8位权重(量化)
};
```

#### 4.2.2 并行计算
```cpp
// 使用OpenMP并行化
vector<Route> parallelKShortestPaths(
    uint64_t start, uint64_t end, int k
) {
    vector<Route> all_routes;
    
    #pragma omp parallel
    {
        vector<Route> local_routes;
        
        #pragma omp for nowait
        for (int i = 0; i < num_start_boundaries; ++i) {
            auto routes = searchFromBoundary(
                start_boundaries[i], end_boundaries
            );
            local_routes.insert(
                local_routes.end(), 
                routes.begin(), 
                routes.end()
            );
        }
        
        #pragma omp critical
        all_routes.insert(
            all_routes.end(), 
            local_routes.begin(), 
            local_routes.end()
        );
    }
    
    // 合并并选择最优k条
    return selectTopK(all_routes, k);
}
```

#### 4.2.3 缓存策略
```cpp
class PathCache {
    struct CacheKey {
        uint64_t from, to;
        MetricType metric;
        
        bool operator==(const CacheKey& other) const {
            return from == other.from && 
                   to == other.to && 
                   metric == other.metric;
        }
    };
    
    struct CacheKeyHash {
        size_t operator()(const CacheKey& key) const {
            return hash<uint64_t>()(key.from) ^ 
                   (hash<uint64_t>()(key.to) << 1) ^
                   (hash<int>()(key.metric) << 2);
        }
    };
    
    unordered_map<CacheKey, vector<Route>, CacheKeyHash> cache;
    LRUEviction<CacheKey> lru;
    
public:
    optional<vector<Route>> get(
        uint64_t from, uint64_t to, MetricType metric
    ) {
        CacheKey key{from, to, metric};
        auto it = cache.find(key);
        if (it != cache.end()) {
            lru.touch(key);
            return it->second;
        }
        return nullopt;
    }
};
```

## 5. 系统部署架构

### 5.1 容器化部署

```yaml
# docker-compose.yml
version: '3.8'
services:
  crp-router:
    image: crp-router:latest
    deploy:
      replicas: 4
      resources:
        limits:
          cpus: '4'
          memory: 16G
        reservations:
          memory: 8G
    environment:
      - OVERLAY_GRAPH_PATH=/data/overlay
      - CACHE_SIZE=4GB
      - WORKER_THREADS=8
    volumes:
      - overlay-data:/data/overlay:ro
    
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on:
      - crp-router
```

### 5.2 负载均衡配置

```nginx
upstream crp_backend {
    least_conn;
    server crp-router-1:8080 weight=1 max_fails=3 fail_timeout=30s;
    server crp-router-2:8080 weight=1 max_fails=3 fail_timeout=30s;
    server crp-router-3:8080 weight=1 max_fails=3 fail_timeout=30s;
    server crp-router-4:8080 weight=1 max_fails=3 fail_timeout=30s;
    keepalive 32;
}

server {
    listen 80;
    
    location /route {
        proxy_pass http://crp_backend;
        proxy_http_version 1.1;
        proxy_set_header Connection "";
        proxy_connect_timeout 50ms;
        proxy_send_timeout 100ms;
        proxy_read_timeout 100ms;
    }
}
```

## 6. 性能监控与优化

### 6.1 关键性能指标

```cpp
struct PerformanceMetrics {
    atomic<uint64_t> total_requests;
    atomic<uint64_t> cache_hits;
    atomic<uint64_t> p50_latency_us;
    atomic<uint64_t> p95_latency_us;
    atomic<uint64_t> p99_latency_us;
    
    void recordRequest(uint64_t latency_us, bool cache_hit) {
        total_requests.fetch_add(1);
        if (cache_hit) cache_hits.fetch_add(1);
        updateLatencyPercentiles(latency_us);
    }
};
```

### 6.2 性能调优建议

1. **预热策略**
   - 启动时加载热点区域数据到内存
   - 预计算常用OD对的路径

2. **动态调整**
   - 根据请求模式动态调整缓存策略
   - 自适应分区粒度

3. **硬件优化**
   - 使用高频CPU（推荐主频 > 3.5GHz）
   - 配置大容量内存（推荐 >= 64GB）
   - 使用NVMe SSD存储预处理数据

## 7. API设计

### 7.1 RESTful API

```yaml
openapi: 3.0.0
paths:
  /route/calculate:
    post:
      summary: 计算路径
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                origin:
                  type: object
                  properties:
                    lng: 
                      type: number
                      description: GCJ02经度
                    lat:
                      type: number
                      description: GCJ02纬度
                destination:
                  type: object
                  properties:
                    lng:
                      type: number
                    lat:
                      type: number
                metric:
                  type: string
                  enum: [time, distance]
                alternatives:
                  type: integer
                  default: 3
                  description: 返回路径数量
      responses:
        200:
          description: 成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  routes:
                    type: array
                    items:
                      type: object
                      properties:
                        duration:
                          type: number
                          description: 耗时（秒）
                        distance:
                          type: number
                          description: 距离（米）
                        geometry:
                          type: string
                          description: 编码的路径几何
                  request_id:
                    type: string
                  elapsed_ms:
                    type: number
```

### 7.2 gRPC接口

```protobuf
syntax = "proto3";

service RouteService {
    rpc CalculateRoute(RouteRequest) returns (RouteResponse);
    rpc CalculateRouteStream(stream RouteRequest) returns (stream RouteResponse);
}

message RouteRequest {
    Coordinate origin = 1;
    Coordinate destination = 2;
    MetricType metric = 3;
    int32 alternatives = 4;
}

message Coordinate {
    double longitude = 1;
    double latitude = 2;
}

enum MetricType {
    TIME = 0;
    DISTANCE = 1;
}

message RouteResponse {
    repeated Route routes = 1;
    string request_id = 2;
    int64 elapsed_ms = 3;
}

message Route {
    double duration = 1;
    double distance = 2;
    string encoded_polyline = 3;
    repeated Coordinate waypoints = 4;
}
```

## 8. 测试方案

### 8.1 性能测试

```bash
# 使用wrk进行压力测试
wrk -t12 -c400 -d30s \
    -s route_test.lua \
    --latency \
    http://localhost/route/calculate

# route_test.lua
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"

local origins = {
    {116.397428, 39.90923},
    {121.473701, 31.230416},
    {113.264385, 23.129110}
}

local destinations = {
    {116.383966, 39.913725},
    {121.506191, 31.245414},
    {113.324520, 23.099994}
}

request = function()
    local idx = math.random(1, 3)
    local body = string.format([[{
        "origin": {"lng": %f, "lat": %f},
        "destination": {"lng": %f, "lat": %f},
        "metric": "time",
        "alternatives": 3
    }]], origins[idx][1], origins[idx][2], 
         destinations[idx][1], destinations[idx][2])
    return wrk.format("POST", "/route/calculate", nil, body)
end
```

### 8.2 准确性验证

```python
# 与参考实现对比
def validate_routes(crp_route, reference_route):
    # 验证路径长度差异
    length_diff = abs(crp_route.distance - reference_route.distance)
    assert length_diff / reference_route.distance < 0.01  # 1%误差
    
    # 验证时间差异
    time_diff = abs(crp_route.duration - reference_route.duration)
    assert time_diff / reference_route.duration < 0.02  # 2%误差
    
    # 验证关键节点
    key_nodes_crp = extract_key_nodes(crp_route)
    key_nodes_ref = extract_key_nodes(reference_route)
    assert jaccard_similarity(key_nodes_crp, key_nodes_ref) > 0.95
```

## 9. 扩展性设计

### 9.1 水平扩展
- 无状态服务设计，支持动态扩缩容
- 数据分片存储，支持分布式部署
- 使用消息队列处理异步任务

### 9.2 功能扩展
- 支持多种路径约束（避让、必经点）
- 支持实时路况融合
- 支持多模式交通（步行、骑行、公交）

## 10. 总结

本方案通过CRP算法的高效实现，结合内存计算、并行处理、智能缓存等技术，能够在100ms内完成路径规划请求。系统采用微服务架构，具有良好的扩展性和可维护性，适合大规模商业化部署。

关键成功因素：
1. 高质量的预处理数据
2. 优化的内存数据结构
3. 高效的缓存策略
4. 合理的系统架构设计
5. 完善的监控和调优机制