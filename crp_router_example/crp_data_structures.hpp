#ifndef CRP_DATA_STRUCTURES_HPP
#define CRP_DATA_STRUCTURES_HPP

#include <vector>
#include <unordered_map>
#include <unordered_set>
#include <memory>
#include <list>

// 坐标结构
struct Coordinate {
    double longitude;  // GCJ02
    double latitude;   // GCJ02
};

// CRP分区信息
struct CRPCell {
    uint32_t cell_id;
    uint8_t level;                              // 分区层级
    std::vector<uint64_t> nodes;                // 分区内的节点
    std::unordered_set<uint64_t> boundary_nodes; // 边界节点
    std::unordered_map<uint64_t, uint64_t> parent_cell; // 上层分区映射
};

// 覆盖图边
struct OverlayEdge {
    uint64_t from;
    uint64_t to;
    uint64_t edge_id;
    double base_cost;     // 基础代价
    bool is_shortcut;     // 是否是捷径边
    std::vector<uint64_t> underlying_path; // 底层路径（如果是捷径）
};

// 覆盖图
class OverlayGraph {
private:
    // 邻接表表示
    std::unordered_map<uint64_t, std::vector<OverlayEdge>> adjacency_list;
    
    // 节点到分区的映射
    std::unordered_map<uint64_t, uint32_t> node_to_cell;
    
    // 分区信息
    std::unordered_map<uint32_t, CRPCell> cells;
    
public:
    // 获取节点的出边
    const std::vector<OverlayEdge>& getOutEdges(uint64_t node) const {
        static const std::vector<OverlayEdge> empty;
        auto it = adjacency_list.find(node);
        return (it != adjacency_list.end()) ? it->second : empty;
    }
    
    // 获取节点所属分区
    uint32_t getCellId(uint64_t node) const {
        auto it = node_to_cell.find(node);
        return (it != node_to_cell.end()) ? it->second : 0;
    }
    
    // 获取分区信息
    const CRPCell* getCell(uint32_t cell_id) const {
        auto it = cells.find(cell_id);
        return (it != cells.end()) ? &it->second : nullptr;
    }
    
    // 添加覆盖边
    void addOverlayEdge(const OverlayEdge& edge) {
        adjacency_list[edge.from].push_back(edge);
    }
    
    // 设置节点分区
    void setNodeCell(uint64_t node, uint32_t cell_id) {
        node_to_cell[node] = cell_id;
    }
    
    // 添加分区
    void addCell(const CRPCell& cell) {
        cells[cell.cell_id] = cell;
    }
};

// 分区划分
class CellPartition {
private:
    std::vector<std::unique_ptr<CRPCell>> cells;
    std::unordered_map<uint64_t, uint32_t> node_to_cell;
    uint8_t num_levels;
    
public:
    explicit CellPartition(uint8_t levels) : num_levels(levels) {}
    
    // 获取节点的分区ID
    uint32_t getCellId(uint64_t node_id) const {
        auto it = node_to_cell.find(node_id);
        return (it != node_to_cell.end()) ? it->second : 0;
    }
    
    // 获取分区
    const CRPCell* getCell(uint32_t cell_id) const {
        if (cell_id < cells.size()) {
            return cells[cell_id].get();
        }
        return nullptr;
    }
    
    // 获取边界节点
    std::vector<uint64_t> getBoundaryNodes(uint64_t node_id) const {
        auto cell_id = getCellId(node_id);
        auto cell = getCell(cell_id);
        if (cell) {
            return std::vector<uint64_t>(
                cell->boundary_nodes.begin(),
                cell->boundary_nodes.end()
            );
        }
        return {};
    }
    
    // 添加分区
    void addCell(std::unique_ptr<CRPCell> cell) {
        uint32_t id = cell->cell_id;
        if (id >= cells.size()) {
            cells.resize(id + 1);
        }
        cells[id] = std::move(cell);
        
        // 更新节点映射
        for (auto node : cells[id]->nodes) {
            node_to_cell[node] = id;
        }
    }
};

// 度量定制
class MetricCustomization {
private:
    // 边权重（针对不同度量类型）
    std::unordered_map<uint64_t, double> time_weights;
    std::unordered_map<uint64_t, double> distance_weights;
    
    // 捷径权重
    struct ShortcutKey {
        uint64_t from;
        uint64_t to;
        
        bool operator==(const ShortcutKey& other) const {
            return from == other.from && to == other.to;
        }
    };
    
    struct ShortcutKeyHash {
        std::size_t operator()(const ShortcutKey& key) const {
            return std::hash<uint64_t>{}(key.from) ^ 
                   (std::hash<uint64_t>{}(key.to) << 1);
        }
    };
    
    std::unordered_map<ShortcutKey, double, ShortcutKeyHash> time_shortcuts;
    std::unordered_map<ShortcutKey, double, ShortcutKeyHash> distance_shortcuts;
    
public:
    // 获取边权重
    double getEdgeWeight(uint64_t edge_id, MetricType metric) const {
        if (metric == MetricType::TIME) {
            auto it = time_weights.find(edge_id);
            return (it != time_weights.end()) ? it->second : 0.0;
        } else {
            auto it = distance_weights.find(edge_id);
            return (it != distance_weights.end()) ? it->second : 0.0;
        }
    }
    
    // 获取捷径权重
    double getShortcutWeight(uint64_t from, uint64_t to, MetricType metric) const {
        ShortcutKey key{from, to};
        if (metric == MetricType::TIME) {
            auto it = time_shortcuts.find(key);
            return (it != time_shortcuts.end()) ? it->second : -1.0;
        } else {
            auto it = distance_shortcuts.find(key);
            return (it != distance_shortcuts.end()) ? it->second : -1.0;
        }
    }
    
    // 设置边权重
    void setEdgeWeight(uint64_t edge_id, MetricType metric, double weight) {
        if (metric == MetricType::TIME) {
            time_weights[edge_id] = weight;
        } else {
            distance_weights[edge_id] = weight;
        }
    }
    
    // 设置捷径权重
    void setShortcutWeight(uint64_t from, uint64_t to, MetricType metric, double weight) {
        ShortcutKey key{from, to};
        if (metric == MetricType::TIME) {
            time_shortcuts[key] = weight;
        } else {
            distance_shortcuts[key] = weight;
        }
    }
};

// LRU缓存实现
template<typename Key, typename Value>
class LRUCache {
private:
    size_t capacity;
    std::list<std::pair<Key, Value>> cache_list;
    std::unordered_map<Key, typename std::list<std::pair<Key, Value>>::iterator> cache_map;
    
public:
    explicit LRUCache(size_t cap = 10000) : capacity(cap) {}
    
    std::optional<Value> get(const Key& key) {
        auto it = cache_map.find(key);
        if (it == cache_map.end()) {
            return std::nullopt;
        }
        
        // 移动到列表前端
        cache_list.splice(cache_list.begin(), cache_list, it->second);
        return it->second->second;
    }
    
    void put(const Key& key, const Value& value) {
        auto it = cache_map.find(key);
        
        if (it != cache_map.end()) {
            // 更新已存在的项
            it->second->second = value;
            cache_list.splice(cache_list.begin(), cache_list, it->second);
            return;
        }
        
        // 添加新项
        if (cache_list.size() >= capacity) {
            // 删除最久未使用的项
            auto last = cache_list.end();
            --last;
            cache_map.erase(last->first);
            cache_list.pop_back();
        }
        
        cache_list.emplace_front(key, value);
        cache_map[key] = cache_list.begin();
    }
    
    void clear() {
        cache_list.clear();
        cache_map.clear();
    }
    
    size_t size() const {
        return cache_list.size();
    }
};

#endif // CRP_DATA_STRUCTURES_HPP