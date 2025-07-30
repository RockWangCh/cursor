#ifndef CRP_ROUTER_HPP
#define CRP_ROUTER_HPP

#include <vector>
#include <unordered_map>
#include <memory>
#include <optional>
#include <queue>
#include <atomic>
#include <shared_mutex>
#include "road_network.hpp"
#include "crp_data_structures.hpp"

enum class MetricType {
    TIME,
    DISTANCE
};

struct Route {
    std::vector<uint64_t> node_sequence;
    std::vector<Coordinate> waypoints;
    double duration;    // 秒
    double distance;    // 米
    std::string encoded_polyline;
    MetricType metric;
};

class CRPRouter {
private:
    // 路网数据
    std::unique_ptr<RoadNetwork> road_network;
    
    // CRP预处理数据
    std::unique_ptr<OverlayGraph> overlay_graph;
    std::unique_ptr<CellPartition> cell_partition;
    std::unique_ptr<MetricCustomization> metric_customization;
    
    // 路径缓存
    mutable std::shared_mutex cache_mutex;
    mutable LRUCache<PathCacheKey, std::vector<Route>> path_cache;
    
    // 性能统计
    mutable std::atomic<uint64_t> cache_hits{0};
    mutable std::atomic<uint64_t> cache_misses{0};
    
public:
    explicit CRPRouter(const std::string& data_path);
    ~CRPRouter() = default;
    
    // 主要接口：查找多条路径
    std::vector<Route> findRoutes(
        double start_lng, double start_lat,
        double end_lng, double end_lat,
        MetricType metric,
        int k_routes = 3
    ) const;
    
private:
    // 地图匹配：将经纬度转换为道路节点
    uint64_t mapMatching(double lng, double lat) const;
    
    // 同分区内的局部搜索
    std::vector<Route> localSearch(
        uint64_t start_node, 
        uint64_t end_node,
        MetricType metric, 
        int k
    ) const;
    
    // 跨分区的覆盖图搜索
    std::vector<Route> overlaySearch(
        uint64_t start_node, 
        uint64_t end_node,
        MetricType metric, 
        int k
    ) const;
    
    // 获取节点所属的分区
    uint32_t getCellForNode(uint64_t node_id) const;
    
    // 获取分区的边界节点
    std::vector<uint64_t> getBoundaryNodes(uint64_t node_id) const;
    
    // 计算局部代价（从节点到边界节点）
    std::unordered_map<uint64_t, double> computeLocalCosts(
        uint64_t source,
        const std::vector<uint64_t>& targets,
        MetricType metric
    ) const;
    
    // 使用Yen算法找k条最短路径
    std::vector<Route> yenKShortestPaths(
        uint64_t start,
        uint64_t end,
        MetricType metric,
        int k
    ) const;
    
    // 构建完整路径
    Route constructRoute(
        uint64_t start,
        uint64_t end,
        const std::vector<uint64_t>& overlay_path,
        MetricType metric
    ) const;
    
    // 获取边的代价
    double getEdgeCost(uint64_t edge_id, MetricType metric) const;
    
    // 编码路径为polyline
    std::string encodePolyline(const std::vector<Coordinate>& points) const;
    
    // 缓存相关
    std::optional<std::vector<Route>> checkCache(
        uint64_t start, 
        uint64_t end, 
        MetricType metric
    ) const;
    
    void updateCache(
        uint64_t start, 
        uint64_t end, 
        MetricType metric,
        const std::vector<Route>& routes
    ) const;
};

// 路径状态（用于搜索）
struct PathState {
    uint64_t node;
    std::vector<uint64_t> path;
    double cost;
    
    bool operator>(const PathState& other) const {
        return cost > other.cost;
    }
};

// 缓存键
struct PathCacheKey {
    uint64_t from;
    uint64_t to;
    MetricType metric;
    
    bool operator==(const PathCacheKey& other) const {
        return from == other.from && 
               to == other.to && 
               metric == other.metric;
    }
};

// 缓存键的哈希函数
struct PathCacheKeyHash {
    std::size_t operator()(const PathCacheKey& key) const {
        std::size_t h1 = std::hash<uint64_t>{}(key.from);
        std::size_t h2 = std::hash<uint64_t>{}(key.to);
        std::size_t h3 = std::hash<int>{}(static_cast<int>(key.metric));
        return h1 ^ (h2 << 1) ^ (h3 << 2);
    }
};

#endif // CRP_ROUTER_HPP