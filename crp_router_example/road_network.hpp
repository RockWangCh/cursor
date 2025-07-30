#ifndef ROAD_NETWORK_HPP
#define ROAD_NETWORK_HPP

#include <vector>
#include <unordered_map>
#include <unordered_set>
#include <string>
#include <memory>
#include <cmath>

// 道路节点
struct RoadNode {
    uint64_t node_id;
    double longitude;  // GCJ02
    double latitude;   // GCJ02
    uint16_t level;    // 道路等级
    std::vector<uint64_t> out_edges;
    std::vector<uint64_t> in_edges;
    
    // 属性
    bool is_traffic_light = false;
    bool is_toll_station = false;
};

// 道路边
struct RoadEdge {
    uint64_t edge_id;
    uint64_t from_node;
    uint64_t to_node;
    float length;      // 米
    float travel_time; // 秒
    uint8_t road_class; // 道路等级
    uint8_t direction;  // 方向限制
    uint8_t num_lanes;  // 车道数
    float speed_limit;  // 限速 km/h
    
    // 几何形状点
    std::vector<std::pair<double, double>> geometry;
    
    // 属性
    bool is_highway = false;
    bool is_tunnel = false;
    bool is_bridge = false;
    std::string road_name;
};

// 空间索引（用于地图匹配）
class SpatialIndex {
private:
    struct GridCell {
        std::vector<uint64_t> node_ids;
    };
    
    // 网格参数
    double min_lng, max_lng;
    double min_lat, max_lat;
    int grid_cols, grid_rows;
    double cell_width, cell_height;
    
    // 网格存储
    std::vector<std::vector<GridCell>> grid;
    
    // 节点位置缓存
    std::unordered_map<uint64_t, std::pair<double, double>> node_positions;
    
public:
    SpatialIndex(double min_lng, double max_lng, 
                 double min_lat, double max_lat,
                 int cols = 1000, int rows = 1000)
        : min_lng(min_lng), max_lng(max_lng),
          min_lat(min_lat), max_lat(max_lat),
          grid_cols(cols), grid_rows(rows) {
        
        cell_width = (max_lng - min_lng) / grid_cols;
        cell_height = (max_lat - min_lat) / grid_rows;
        grid.resize(grid_rows, std::vector<GridCell>(grid_cols));
    }
    
    // 添加节点到索引
    void addNode(uint64_t node_id, double lng, double lat) {
        int col = static_cast<int>((lng - min_lng) / cell_width);
        int row = static_cast<int>((lat - min_lat) / cell_height);
        
        if (col >= 0 && col < grid_cols && row >= 0 && row < grid_rows) {
            grid[row][col].node_ids.push_back(node_id);
            node_positions[node_id] = {lng, lat};
        }
    }
    
    // 查找最近的节点
    uint64_t findNearestNode(double lng, double lat, double max_distance_m = 50.0) const {
        int center_col = static_cast<int>((lng - min_lng) / cell_width);
        int center_row = static_cast<int>((lat - min_lat) / cell_height);
        
        uint64_t nearest_node = 0;
        double min_distance = max_distance_m;
        
        // 搜索半径（以网格数计）
        int search_radius = static_cast<int>(
            std::ceil(max_distance_m / (111320.0 * cell_width))
        );
        
        // 在周围网格中搜索
        for (int r = -search_radius; r <= search_radius; r++) {
            for (int c = -search_radius; c <= search_radius; c++) {
                int row = center_row + r;
                int col = center_col + c;
                
                if (row >= 0 && row < grid_rows && col >= 0 && col < grid_cols) {
                    for (uint64_t node_id : grid[row][col].node_ids) {
                        auto it = node_positions.find(node_id);
                        if (it != node_positions.end()) {
                            double dist = calculateDistance(
                                lng, lat, it->second.first, it->second.second
                            );
                            if (dist < min_distance) {
                                min_distance = dist;
                                nearest_node = node_id;
                            }
                        }
                    }
                }
            }
        }
        
        return nearest_node;
    }
    
private:
    // 计算两点间距离（米）
    double calculateDistance(double lng1, double lat1, 
                           double lng2, double lat2) const {
        const double R = 6371000.0; // 地球半径（米）
        double lat1_rad = lat1 * M_PI / 180.0;
        double lat2_rad = lat2 * M_PI / 180.0;
        double delta_lat = (lat2 - lat1) * M_PI / 180.0;
        double delta_lng = (lng2 - lng1) * M_PI / 180.0;
        
        double a = sin(delta_lat/2) * sin(delta_lat/2) +
                   cos(lat1_rad) * cos(lat2_rad) *
                   sin(delta_lng/2) * sin(delta_lng/2);
        double c = 2 * atan2(sqrt(a), sqrt(1-a));
        
        return R * c;
    }
};

// 路网类
class RoadNetwork {
private:
    // 节点和边存储
    std::unordered_map<uint64_t, RoadNode> nodes;
    std::unordered_map<uint64_t, RoadEdge> edges;
    
    // 空间索引
    std::unique_ptr<SpatialIndex> spatial_index;
    
    // 统计信息
    size_t total_nodes = 0;
    size_t total_edges = 0;
    
public:
    RoadNetwork() = default;
    
    // 加载路网数据
    bool loadFromFile(const std::string& filename);
    
    // 获取节点
    const RoadNode* getNode(uint64_t node_id) const {
        auto it = nodes.find(node_id);
        return (it != nodes.end()) ? &it->second : nullptr;
    }
    
    // 获取边
    const RoadEdge* getEdge(uint64_t edge_id) const {
        auto it = edges.find(edge_id);
        return (it != edges.end()) ? &it->second : nullptr;
    }
    
    // 获取节点的出边
    std::vector<uint64_t> getOutEdges(uint64_t node_id) const {
        auto node = getNode(node_id);
        return node ? node->out_edges : std::vector<uint64_t>();
    }
    
    // 获取节点的入边
    std::vector<uint64_t> getInEdges(uint64_t node_id) const {
        auto node = getNode(node_id);
        return node ? node->in_edges : std::vector<uint64_t>();
    }
    
    // 地图匹配
    uint64_t mapMatch(double lng, double lat) const {
        if (spatial_index) {
            return spatial_index->findNearestNode(lng, lat);
        }
        return 0;
    }
    
    // 添加节点
    void addNode(const RoadNode& node) {
        nodes[node.node_id] = node;
        total_nodes++;
        
        // 添加到空间索引
        if (spatial_index) {
            spatial_index->addNode(node.node_id, node.longitude, node.latitude);
        }
    }
    
    // 添加边
    void addEdge(const RoadEdge& edge) {
        edges[edge.edge_id] = edge;
        total_edges++;
        
        // 更新节点的边列表
        if (auto* from_node = const_cast<RoadNode*>(getNode(edge.from_node))) {
            from_node->out_edges.push_back(edge.edge_id);
        }
        if (auto* to_node = const_cast<RoadNode*>(getNode(edge.to_node))) {
            to_node->in_edges.push_back(edge.edge_id);
        }
    }
    
    // 初始化空间索引
    void initializeSpatialIndex() {
        // 计算边界
        double min_lng = 180.0, max_lng = -180.0;
        double min_lat = 90.0, max_lat = -90.0;
        
        for (const auto& [id, node] : nodes) {
            min_lng = std::min(min_lng, node.longitude);
            max_lng = std::max(max_lng, node.longitude);
            min_lat = std::min(min_lat, node.latitude);
            max_lat = std::max(max_lat, node.latitude);
        }
        
        // 创建索引
        spatial_index = std::make_unique<SpatialIndex>(
            min_lng - 0.01, max_lng + 0.01,
            min_lat - 0.01, max_lat + 0.01
        );
        
        // 添加所有节点
        for (const auto& [id, node] : nodes) {
            spatial_index->addNode(id, node.longitude, node.latitude);
        }
    }
    
    // 获取统计信息
    size_t getNodeCount() const { return total_nodes; }
    size_t getEdgeCount() const { return total_edges; }
};

// 四维图新数据适配器
class NavinfoDataAdapter {
public:
    // 从四维图新格式加载路网
    static std::unique_ptr<RoadNetwork> loadNavinfoData(
        const std::string& node_file,
        const std::string& edge_file,
        const std::string& geometry_file
    );
    
    // 转换坐标系（如果需要）
    static std::pair<double, double> wgs84ToGcj02(double lng, double lat);
    static std::pair<double, double> gcj02ToWgs84(double lng, double lat);
    
private:
    // 解析节点文件
    static void parseNodeFile(const std::string& filename, RoadNetwork& network);
    
    // 解析边文件
    static void parseEdgeFile(const std::string& filename, RoadNetwork& network);
    
    // 解析几何文件
    static void parseGeometryFile(const std::string& filename, RoadNetwork& network);
};

#endif // ROAD_NETWORK_HPP