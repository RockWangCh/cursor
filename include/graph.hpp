#pragma once

#include <vector>
#include <cstdint>

namespace crp {

using NodeId = uint32_t;
using EdgeId = uint32_t;

struct GeoPoint {
    double lat; // GCJ-02 latitude
    double lon; // GCJ-02 longitude
};

struct Edge {
    NodeId target;      // head vertex
    float length;       // metres
    float travel_time;  // seconds (free-flow or customised)
};

// Compressed Sparse Row representation of the base graph.
class Graph {
public:
    std::vector<uint32_t> offsets; // size = num_nodes + 1
    std::vector<Edge>      edges;
    std::vector<GeoPoint>  coords;

    [[nodiscard]] size_t num_nodes() const { return offsets.empty() ? 0 : offsets.size() - 1; }

    // Range of outgoing edges for vertex u
    [[nodiscard]] auto neighbors(NodeId u) const {
        return std::make_pair(edges.begin() + offsets[u], edges.begin() + offsets[u + 1]);
    }
};

} // namespace crp