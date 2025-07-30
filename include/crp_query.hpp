#pragma once

#include <optional>
#include <queue>
#include <vector>
#include "graph.hpp"
#include "overlay.hpp"
#include "customization.hpp"

namespace crp {

struct Path {
    float total_weight{};
    std::vector<NodeId> nodes;
};

class CRPQueryEngine {
public:
    CRPQueryEngine(const Graph& base, const OverlayGraph& overlay)
        : base_(base), overlay_(overlay) {}

    // Classic CRP multi-level Dijkstra (placeholder implementation)
    std::optional<Path> shortest_path(NodeId /*s*/, NodeId /*t*/, WeightType /*type*/) const {
        // TODO: implement multilevel Dijkstra using overlay
        return std::nullopt;
    }

    // k-shortest using Yen's algorithm on top of shortest_path (placeholder)
    std::vector<Path> k_shortest_paths(NodeId s, NodeId t, std::size_t k, WeightType type) const {
        std::vector<Path> result;
        if (auto path = shortest_path(s, t, type); path) {
            result.push_back(*path);
        }
        // TODO: generate k>1 alternatives (Yen's algorithm)
        if (result.size() > k) result.resize(k);
        return result;
    }

private:
    const Graph&       base_;
    const OverlayGraph& overlay_;
};

} // namespace crp