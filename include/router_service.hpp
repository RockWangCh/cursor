#pragma once

#include "crp_query.hpp"

namespace crp {

struct RouteRequest {
    double origin_lat;
    double origin_lon;
    double dest_lat;
    double dest_lon;

    std::size_t k = 1;
    WeightType  weight = WeightType::Time;
};

struct RouteResponse {
    std::vector<Path> paths;
};

class RouterService {
public:
    explicit RouterService(CRPQueryEngine& engine) : engine_(engine) {}

    RouteResponse handle(const RouteRequest& req) {
        RouteResponse resp;
        // TODO: snap lat/lon to nearest node id
        NodeId s = 0;
        NodeId t = 0;
        resp.paths = engine_.k_shortest_paths(s, t, req.k, req.weight);
        return resp;
    }

private:
    CRPQueryEngine& engine_;
};

} // namespace crp