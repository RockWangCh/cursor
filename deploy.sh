#!/bin/bash

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查系统要求
check_requirements() {
    log_info "检查系统要求..."
    
    # 检查Docker
    if ! command -v docker &> /dev/null; then
        log_error "Docker未安装，请先安装Docker"
        exit 1
    fi
    
    # 检查Docker Compose
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose未安装，请先安装Docker Compose"
        exit 1
    fi
    
    # 检查可用内存
    available_memory=$(free -m | awk 'NR==2{printf "%.0f", $7}')
    if [ "$available_memory" -lt 8192 ]; then
        log_warning "可用内存不足8GB，建议增加内存以获得最佳性能"
    fi
    
    # 检查磁盘空间
    available_space=$(df -m . | awk 'NR==2{print $4}')
    if [ "$available_space" -lt 10240 ]; then
        log_error "可用磁盘空间不足10GB，请释放空间"
        exit 1
    fi
    
    log_success "系统要求检查通过"
}

# 创建必要目录
create_directories() {
    log_info "创建必要目录..."
    
    mkdir -p data/{nodes,edges,processed}
    mkdir -p index
    mkdir -p logs
    mkdir -p configs/{nginx,redis,grafana,prometheus}
    mkdir -p sql/init
    
    log_success "目录创建完成"
}

# 生成配置文件
generate_configs() {
    log_info "生成配置文件..."
    
    # Nginx配置
    cat > configs/nginx.conf << 'EOF'
worker_processes auto;
error_log /var/log/nginx/error.log warn;
pid /var/run/nginx.pid;

events {
    worker_connections 1024;
    use epoll;
    multi_accept on;
}

http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;

    log_format main '$remote_addr - $remote_user [$time_local] "$request" '
                   '$status $body_bytes_sent "$http_referer" '
                   '"$http_user_agent" "$http_x_forwarded_for" '
                   'rt=$request_time uct="$upstream_connect_time" '
                   'uht="$upstream_header_time" urt="$upstream_response_time"';

    access_log /var/log/nginx/access.log main;

    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;
    keepalive_timeout 65;
    types_hash_max_size 2048;

    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    upstream route_backend {
        least_conn;
        server route-service:8080 max_fails=3 fail_timeout=30s;
        keepalive 32;
    }

    server {
        listen 80;
        server_name _;

        location /health {
            access_log off;
            proxy_pass http://route_backend;
        }

        location /api {
            proxy_pass http://route_backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
            proxy_connect_timeout 5s;
            proxy_send_timeout 60s;
            proxy_read_timeout 60s;
        }

        location /metrics {
            proxy_pass http://route_backend;
            allow 172.20.0.0/16;
            deny all;
        }
    }
}
EOF

    # Redis配置
    cat > configs/redis.conf << 'EOF'
# 基本配置
bind 0.0.0.0
port 6379
tcp-backlog 511
timeout 300
tcp-keepalive 300

# 内存配置
maxmemory 512mb
maxmemory-policy allkeys-lru

# 持久化配置
save 900 1
save 300 10
save 60 10000
stop-writes-on-bgsave-error yes
rdbcompression yes
rdbchecksum yes
dbfilename dump.rdb
dir /data

# 日志配置
loglevel notice
logfile ""

# 性能配置
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-size -2
list-compress-depth 0
set-max-intset-entries 512
zset-max-ziplist-entries 128
zset-max-ziplist-value 64
hll-sparse-max-bytes 3000
EOF

    # Prometheus配置
    cat > configs/prometheus.yml << 'EOF'
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']

  - job_name: 'route-service'
    static_configs:
      - targets: ['route-service:8080']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'postgres'
    static_configs:
      - targets: ['postgres:5432']

  - job_name: 'redis'
    static_configs:
      - targets: ['redis:6379']
EOF

    # 数据库初始化脚本
    cat > sql/init/01_create_tables.sql << 'EOF'
-- 启用PostGIS扩展
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS pg_stat_statements;

-- 创建节点表
CREATE TABLE IF NOT EXISTS nodes (
    id BIGSERIAL PRIMARY KEY,
    longitude DOUBLE PRECISION NOT NULL,
    latitude DOUBLE PRECISION NOT NULL,
    elevation REAL,
    node_type SMALLINT NOT NULL DEFAULT 1,
    geom GEOMETRY(POINT, 4326),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建边表
CREATE TABLE IF NOT EXISTS edges (
    id BIGSERIAL PRIMARY KEY,
    from_node_id BIGINT NOT NULL REFERENCES nodes(id),
    to_node_id BIGINT NOT NULL REFERENCES nodes(id),
    length INTEGER NOT NULL,
    time_cost INTEGER NOT NULL,
    speed SMALLINT NOT NULL,
    road_level SMALLINT NOT NULL DEFAULT 1,
    direction SMALLINT NOT NULL DEFAULT 0,
    geom GEOMETRY(LINESTRING, 4326),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_nodes_geom ON nodes USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_edges_geom ON edges USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_edges_from_node ON edges (from_node_id);
CREATE INDEX IF NOT EXISTS idx_edges_to_node ON edges (to_node_id);
CREATE INDEX IF NOT EXISTS idx_nodes_location ON nodes (longitude, latitude);

-- 创建查询统计表
CREATE TABLE IF NOT EXISTS query_stats (
    id BIGSERIAL PRIMARY KEY,
    start_lon DOUBLE PRECISION NOT NULL,
    start_lat DOUBLE PRECISION NOT NULL,
    end_lon DOUBLE PRECISION NOT NULL,
    end_lat DOUBLE PRECISION NOT NULL,
    weight_type SMALLINT NOT NULL,
    response_time INTEGER NOT NULL,
    cache_hit BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_query_stats_time ON query_stats (created_at);
EOF

    log_success "配置文件生成完成"
}

# 构建和启动服务
deploy_services() {
    log_info "构建和启动服务..."
    
    # 构建镜像
    docker-compose build --no-cache
    
    # 启动服务
    docker-compose up -d
    
    log_success "服务启动完成"
}

# 健康检查
health_check() {
    log_info "执行健康检查..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        log_info "健康检查尝试 $attempt/$max_attempts"
        
        # 检查路径规划服务
        if curl -f http://localhost:8080/health > /dev/null 2>&1; then
            log_success "路径规划服务健康检查通过"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            log_error "健康检查失败，服务未能正常启动"
            docker-compose logs route-service
            exit 1
        fi
        
        sleep 10
        ((attempt++))
    done
}

# 性能测试
performance_test() {
    log_info "开始性能测试..."
    
    # 创建性能测试脚本
    cat > performance_test.py << 'EOF'
#!/usr/bin/env python3
import requests
import time
import json
import statistics
import concurrent.futures
import random

BASE_URL = "http://localhost:8080"

# 测试数据：北京市内的一些坐标点
test_points = [
    (116.3974, 39.9093),  # 天安门
    (116.4074, 39.9042),  # 王府井
    (116.3833, 39.9167),  # 西单
    (116.4167, 39.9167),  # 东单
    (116.3906, 39.8906),  # 前门
    (116.4190, 39.9260),  # 雍和宫
    (116.3420, 39.9280),  # 西直门
    (116.4570, 39.9380),  # 朝阳门
]

def make_request(start_point, end_point):
    """发起路径规划请求"""
    payload = {
        "start": {"longitude": start_point[0], "latitude": start_point[1]},
        "end": {"longitude": end_point[0], "latitude": end_point[1]},
        "weight_type": 0,  # 距离优先
        "max_routes": 3
    }
    
    start_time = time.time()
    try:
        response = requests.post(
            f"{BASE_URL}/api/v1/route/multiple",
            json=payload,
            timeout=5
        )
        response_time = (time.time() - start_time) * 1000  # 转换为毫秒
        
        if response.status_code == 200:
            return response_time, True, response.json()
        else:
            return response_time, False, None
    except Exception as e:
        response_time = (time.time() - start_time) * 1000
        return response_time, False, str(e)

def concurrent_test(num_requests=100, concurrency=10):
    """并发测试"""
    print(f"开始并发测试: {num_requests} 请求, {concurrency} 并发")
    
    # 准备测试数据
    test_data = []
    for _ in range(num_requests):
        start = random.choice(test_points)
        end = random.choice([p for p in test_points if p != start])
        test_data.append((start, end))
    
    # 执行并发测试
    response_times = []
    success_count = 0
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as executor:
        futures = [executor.submit(make_request, start, end) for start, end in test_data]
        
        for future in concurrent.futures.as_completed(futures):
            response_time, success, result = future.result()
            response_times.append(response_time)
            if success:
                success_count += 1
    
    # 统计结果
    if response_times:
        avg_time = statistics.mean(response_times)
        p95_time = statistics.quantiles(response_times, n=20)[18]  # P95
        p99_time = statistics.quantiles(response_times, n=100)[98]  # P99
        max_time = max(response_times)
        min_time = min(response_times)
        
        print(f"\n性能测试结果:")
        print(f"总请求数: {num_requests}")
        print(f"成功请求数: {success_count}")
        print(f"成功率: {success_count/num_requests*100:.2f}%")
        print(f"平均响应时间: {avg_time:.2f} ms")
        print(f"P95响应时间: {p95_time:.2f} ms")
        print(f"P99响应时间: {p99_time:.2f} ms")
        print(f"最大响应时间: {max_time:.2f} ms")
        print(f"最小响应时间: {min_time:.2f} ms")
        
        # 检查性能要求
        if p99_time <= 100:
            print(f"✅ P99响应时间 {p99_time:.2f}ms 满足 <100ms 的要求")
        else:
            print(f"❌ P99响应时间 {p99_time:.2f}ms 超过 100ms 要求")
            
        return avg_time, p95_time, p99_time, success_count/num_requests
    else:
        print("没有收到任何响应")
        return 0, 0, 0, 0

if __name__ == "__main__":
    # 先测试单个请求
    print("测试单个请求...")
    start_point = test_points[0]
    end_point = test_points[1]
    response_time, success, result = make_request(start_point, end_point)
    
    if success:
        print(f"单个请求成功，响应时间: {response_time:.2f} ms")
        print(f"响应数据: {json.dumps(result, indent=2, ensure_ascii=False)}")
    else:
        print(f"单个请求失败: {result}")
        exit(1)
    
    # 并发测试
    concurrent_test(100, 10)
    concurrent_test(500, 20)
    concurrent_test(1000, 50)
EOF

    # 运行性能测试
    python3 performance_test.py
    
    # 清理测试文件
    rm -f performance_test.py
    
    log_success "性能测试完成"
}

# 显示服务状态
show_status() {
    log_info "显示服务状态..."
    
    echo "=== Docker 容器状态 ==="
    docker-compose ps
    
    echo -e "\n=== 服务访问地址 ==="
    echo "路径规划API: http://localhost:8080/api/v1/route/multiple"
    echo "健康检查: http://localhost:8080/health"
    echo "Prometheus: http://localhost:9090"
    echo "Grafana: http://localhost:3000 (admin/admin123)"
    echo "系统状态: http://localhost:8080/api/v1/status"
    
    echo -e "\n=== 示例API调用 ==="
    cat << 'EOF'
curl -X POST http://localhost:8080/api/v1/route/multiple \
  -H "Content-Type: application/json" \
  -d '{
    "start": {"longitude": 116.3974, "latitude": 39.9093},
    "end": {"longitude": 116.4074, "latitude": 39.9042},
    "weight_type": 0,
    "max_routes": 3
  }'
EOF
}

# 清理函数
cleanup() {
    log_info "清理资源..."
    docker-compose down -v
    docker system prune -f
    log_success "清理完成"
}

# 主函数
main() {
    case "${1:-deploy}" in
        "check")
            check_requirements
            ;;
        "deploy")
            check_requirements
            create_directories
            generate_configs
            deploy_services
            health_check
            show_status
            ;;
        "test")
            health_check
            performance_test
            ;;
        "status")
            show_status
            ;;
        "cleanup")
            cleanup
            ;;
        "help")
            echo "用法: $0 [check|deploy|test|status|cleanup|help]"
            echo "  check   - 检查系统要求"
            echo "  deploy  - 完整部署系统"
            echo "  test    - 运行性能测试"
            echo "  status  - 显示服务状态"
            echo "  cleanup - 清理所有资源"
            echo "  help    - 显示帮助信息"
            ;;
        *)
            log_error "未知命令: $1"
            echo "使用 '$0 help' 查看帮助信息"
            exit 1
            ;;
    esac
}

# 脚本入口
main "$@"