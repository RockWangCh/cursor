#pragma once

#include "graph.hpp"
#include "overlay.hpp"

namespace crp {

enum class WeightType { Time, Distance };

// Responsible for applying edge weights (traffic, vehicle profile, etc.)
class Customizer {
public:
    Customizer(const Graph& base, OverlayGraph& overlay)
        : base_(base), overlay_(overlay) {}

    // Apply new weight type to overlay graph. In real system this runs CRP customize kernel.
    void apply_weights(WeightType type) {
        (void)type;
        // TODO: run CRP customization kernel here
    }

private:
    const Graph&    base_;
    OverlayGraph&   overlay_;
};

} // namespace crp