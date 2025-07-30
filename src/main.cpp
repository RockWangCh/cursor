#include "graph.hpp"
#include "overlay.hpp"
#include "customization.hpp"
#include "crp_query.hpp"
#include "router_service.hpp"

#include <iostream>

int main() {
    // --- Initialise empty graph (placeholder) ---
    crp::Graph graph;
    crp::OverlayGraph overlay;

    // TODO: load graph & overlay files

    // --- Customise weights (default: time) ---
    crp::Customizer customizer(graph, overlay);
    customizer.apply_weights(crp::WeightType::Time);

    // --- Build query engine & service ---
    crp::CRPQueryEngine engine(graph, overlay);
    crp::RouterService  service(engine);

    // --- Build request ---
    crp::RouteRequest req{};
    req.origin_lat = 31.2304;
    req.origin_lon = 121.4737;
    req.dest_lat   = 31.2905;
    req.dest_lon   = 121.2197;
    req.k          = 3;

    auto resp = service.handle(req);

    std::cout << "Found " << resp.paths.size() << " path(s)\n";
    return 0;
}