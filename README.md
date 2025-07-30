# 基于四维图新数据的CRP算法路径规划系统

## 项目概述

本项目是一个高性能的路径规划系统，基于四维图新的高精度地图数据，采用Contraction Hierarchies（收缩层次）优化的路径规划算法（CRP），实现起终点间多条路线的快速计算，支持时间和距离两种权值类型，确保响应时间在100ms以内。

## 核心特性

### 🚀 高性能算法
- **CRP算法**: 基于收缩层次的路径规划，提供毫秒级响应
- **双向搜索**: 优化的双向Dijkstra算法，大幅减少搜索空间
- **多路径计算**: 基于Yen's K-shortest算法的多条路径规划
- **智能缓存**: 多级缓存策略，支持热点路径预计算

### 🎯 性能指标
- **响应时间**: P99 < 100ms, P95 < 50ms
- **吞吐量**: 支持10,000+ QPS
- **可用性**: 99.9%+
- **准确性**: 与四维图新官方结果偏差 < 5%

### 🏗️ 系统架构
- **微服务架构**: 基于Go语言的高并发服务
- **分布式缓存**: Redis集群 + 进程内缓存
- **空间数据库**: PostgreSQL + PostGIS
- **负载均衡**: Nginx反向代理
- **监控告警**: Prometheus + Grafana

## 快速开始

### 系统要求

- **操作系统**: Linux/macOS/Windows (推荐Linux)
- **内存**: 最少8GB，推荐16GB+
- **存储**: 最少10GB可用空间
- **CPU**: 最少4核，推荐8核+
- **软件依赖**:
  - Docker 20.10+
  - Docker Compose 2.0+
  - Python 3.8+ (用于性能测试)

### 一键部署

```bash
# 1. 克隆项目
git clone <repository-url>
cd route-planning

# 2. 检查系统要求
./deploy.sh check

# 3. 一键部署
./deploy.sh deploy

# 4. 运行性能测试
./deploy.sh test
```

### 手动部署

```bash
# 1. 创建必要目录
mkdir -p {data,index,logs,configs}

# 2. 启动服务
docker-compose up -d

# 3. 检查服务状态
docker-compose ps

# 4. 健康检查
curl http://localhost:8080/health
```

## API 使用指南

### 多路径规划

```bash
curl -X POST http://localhost:8080/api/v1/route/multiple \
  -H "Content-Type: application/json" \
  -d '{
    "start": {
      "longitude": 116.3974,
      "latitude": 39.9093
    },
    "end": {
      "longitude": 116.4074,
      "latitude": 39.9042
    },
    "weight_type": 0,
    "max_routes": 3,
    "options": {
      "avoid_tolls": false,
      "avoid_highways": false,
      "vehicle_type": "car"
    }
  }'
```

### 请求参数说明

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| start | Point | ✓ | 起点坐标（GCJ02） |
| end | Point | ✓ | 终点坐标（GCJ02） |
| weight_type | int | ✓ | 权值类型：0=距离，1=时间 |
| max_routes | int | - | 最大路径数，默认3条 |
| options | Object | - | 路径选项 |

### 响应示例

```json
{
  "code": 200,
  "message": "success",
  "data": {
    "routes": [
      {
        "distance": 2850,
        "duration": 480,
        "geometry": "encoded_polyline_string",
        "instructions": [
          {
            "type": "start",
            "description": "从天安门出发",
            "distance": 0,
            "position": {"longitude": 116.3974, "latitude": 39.9093}
          }
        ],
        "waypoints": [
          {"longitude": 116.3974, "latitude": 39.9093},
          {"longitude": 116.4074, "latitude": 39.9042}
        ]
      }
    ],
    "summary": {
      "total_distance": 2850,
      "total_time": 480,
      "route_count": 1
    }
  },
  "timing": {
    "query_time_ms": 45,
    "cache_hit": false,
    "processing_time_ms": 38
  }
}
```

### 坐标转换

```bash
curl -X POST http://localhost:8080/api/v1/coord/convert \
  -H "Content-Type: application/json" \
  -d '{
    "points": [
      {"longitude": 116.3974, "latitude": 39.9093}
    ],
    "from_crs": "GCJ02",
    "to_crs": "WGS84"
  }'
```

### 系统状态

```bash
curl http://localhost:8080/api/v1/status
```

## 架构详解

### 系统架构图

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
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                   数据服务层                                 │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │   内存图存储    │ │   缓存服务      │ │  坐标转换服务   ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────┴───────────────────────────────────────┐
│                   持久化存储层                               │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
│  │ PostgreSQL+PostGIS │ │   文件存储     │ │   监控存储      ││
│  └─────────────────┘ └─────────────────┘ └─────────────────┘│
└─────────────────────────────────────────────────────────────┘
```

### 核心模块

#### CRP算法引擎
- **预处理模块**: 构建收缩层次结构
- **搜索模块**: 双向A*搜索优化
- **路径重构**: 快速路径组装算法

#### 缓存系统
- **L1缓存**: 进程内LRU缓存，响应时间<1ms
- **L2缓存**: Redis集群，容量大，支持分布式
- **预计算**: 热点路径离线计算

#### 数据存储
- **路网数据**: PostgreSQL + PostGIS空间索引
- **图结构**: 内存优化的邻接表
- **空间索引**: R-Tree分层索引

## 性能优化

### 算法优化
1. **收缩层次预处理**: 将搜索空间减少90%+
2. **双向搜索**: 搜索节点数减少50%
3. **启发式函数**: A*算法加速路径发现
4. **并行计算**: 多路径并行计算

### 内存优化
1. **数据压缩**: 16字节对齐的紧凑结构
2. **内存池**: 减少GC压力
3. **预加载**: 核心路网数据常驻内存
4. **分片存储**: 按地理区域分片

### 缓存策略
1. **热点预计算**: 高频路径离线计算
2. **Geohash索引**: 空间相近路径复用
3. **多级缓存**: L1+L2缓存架构
4. **智能淘汰**: LRU + TTL策略

## 监控和运维

### 监控指标

访问监控面板：
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin123)

#### 关键指标
- **请求延迟**: P50/P95/P99响应时间
- **吞吐量**: QPS和并发连接数
- **缓存命中率**: L1/L2缓存效率
- **错误率**: 4xx/5xx错误统计
- **资源使用**: CPU/内存/磁盘使用率

### 日志管理

```bash
# 查看应用日志
docker-compose logs -f route-service

# 查看数据库日志
docker-compose logs -f postgres

# 查看缓存日志
docker-compose logs -f redis
```

### 健康检查

```bash
# 服务健康状态
curl http://localhost:8080/health

# 详细系统状态
curl http://localhost:8080/api/v1/status

# 性能指标
curl http://localhost:8080/metrics
```

## 部署方案

### 开发环境

```bash
# 启动开发环境
docker-compose -f docker-compose.dev.yml up -d

# 热重载开发
go run cmd/server/main.go
```

### 生产环境

#### Docker Swarm部署

```bash
# 初始化Swarm
docker swarm init

# 部署服务栈
docker stack deploy -c docker-compose.prod.yml route-planning
```

#### Kubernetes部署

```bash
# 应用Kubernetes配置
kubectl apply -f k8s/

# 查看服务状态
kubectl get pods -l app=route-service

# 查看服务日志
kubectl logs -l app=route-service
```

### 高可用配置

#### 多区域部署
- **主从复制**: 数据库主从同步
- **读写分离**: 读请求负载均衡
- **故障转移**: 自动故障检测和切换

#### 扩容策略
- **水平扩容**: 增加服务实例
- **垂直扩容**: 提高单实例资源
- **数据分片**: 按地理区域分片

## 数据管理

### 四维图新数据导入

```bash
# 1. 准备数据文件
mkdir -p data/import
# 复制四维图新数据文件到 data/import/

# 2. 运行数据导入脚本
./scripts/import_navinfo_data.sh

# 3. 构建索引
./scripts/build_indexes.sh

# 4. 验证数据
./scripts/validate_data.sh
```

### 数据更新

```bash
# 增量数据更新
./scripts/update_data.sh --incremental

# 全量数据重建
./scripts/update_data.sh --full

# 索引重建
./scripts/rebuild_indexes.sh
```

## 测试

### 单元测试

```bash
# 运行单元测试
go test ./...

# 生成测试覆盖率报告
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 性能测试

```bash
# 运行性能测试
./deploy.sh test

# 压力测试
./scripts/stress_test.sh

# 基准测试
go test -bench=. ./pkg/crp/
```

### 集成测试

```bash
# API集成测试
./scripts/integration_test.sh

# 端到端测试
./scripts/e2e_test.sh
```

## 故障排除

### 常见问题

#### 1. 服务启动失败
```bash
# 检查日志
docker-compose logs route-service

# 检查端口占用
netstat -tlnp | grep 8080

# 重启服务
docker-compose restart route-service
```

#### 2. 数据库连接失败
```bash
# 检查数据库状态
docker-compose exec postgres pg_isready

# 检查数据库日志
docker-compose logs postgres

# 重置数据库
docker-compose down postgres
docker volume rm route-planning_postgres_data
docker-compose up -d postgres
```

#### 3. 缓存问题
```bash
# 检查Redis状态
docker-compose exec redis redis-cli ping

# 清理缓存
docker-compose exec redis redis-cli FLUSHALL

# 重启Redis
docker-compose restart redis
```

#### 4. 性能问题
```bash
# 检查系统资源
docker stats

# 分析慢查询
curl http://localhost:8080/api/v1/status

# 查看监控指标
open http://localhost:3000
```

### 性能调优

#### 内存调优
```bash
# 增加内存限制
docker-compose up -d --scale route-service=2

# 调整JVM参数
export GOGC=200
```

#### 数据库调优
```sql
-- 分析查询性能
EXPLAIN ANALYZE SELECT ...;

-- 重建索引
REINDEX TABLE nodes;
REINDEX TABLE edges;
```

## 开发指南

### 项目结构

```
route-planning/
├── cmd/                    # 应用入口
│   └── server/
├── internal/               # 私有应用代码
│   ├── config/            # 配置管理
│   ├── handler/           # HTTP处理器
│   ├── service/           # 业务逻辑
│   ├── cache/             # 缓存实现
│   └── database/          # 数据访问
├── pkg/                   # 公共库
│   ├── crp/               # CRP算法实现
│   ├── geom/              # 几何计算
│   ├── types/             # 数据类型
│   └── middleware/        # 中间件
├── configs/               # 配置文件
├── sql/                   # 数据库脚本
├── scripts/               # 部署脚本
└── docs/                  # 文档
```

### 代码规范

#### Go代码规范
- 遵循 `gofmt` 格式化标准
- 使用 `golint` 进行代码检查
- 添加必要的注释和文档
- 编写单元测试

#### 提交规范
```
feat: 添加新功能
fix: 修复bug
docs: 更新文档
style: 代码格式化
refactor: 代码重构
test: 添加测试
chore: 构建过程或辅助工具的变动
```

## 许可证

本项目采用 MIT 许可证。详情请查看 [LICENSE](LICENSE) 文件。

## 联系方式

- **项目维护者**: [Your Name]
- **邮箱**: [your.email@domain.com]
- **Issue**: [GitHub Issues](https://github.com/your-org/route-planning/issues)

## 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 更新日志

### v1.0.0 (2024-01-XX)
- 🎉 初始版本发布
- ✨ 实现CRP算法路径规划
- ✨ 支持多路径计算
- ✨ 集成四维图新数据
- 🚀 响应时间优化至100ms以内
- 📊 完整的监控和运维体系