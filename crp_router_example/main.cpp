#include <iostream>
#include <memory>
#include <thread>
#include <chrono>
#include <httplib.h>
#include <nlohmann/json.hpp>
#include "crp_router.hpp"
#include "performance_monitor.hpp"

using json = nlohmann::json;
using namespace std::chrono;

class RouteService {
private:
    std::unique_ptr<CRPRouter> router;
    std::unique_ptr<PerformanceMonitor> monitor;
    
public:
    RouteService(const std::string& data_path) {
        router = std::make_unique<CRPRouter>(data_path);
        monitor = std::make_unique<PerformanceMonitor>();
    }
    
    json handleRouteRequest(const json& request) {
        auto start_time = high_resolution_clock::now();
        
        try {
            // 解析请求参数
            double origin_lng = request["origin"]["lng"];
            double origin_lat = request["origin"]["lat"];
            double dest_lng = request["destination"]["lng"];
            double dest_lat = request["destination"]["lat"];
            std::string metric_str = request.value("metric", "time");
            int alternatives = request.value("alternatives", 3);
            
            // 转换度量类型
            MetricType metric = (metric_str == "distance") ? 
                MetricType::DISTANCE : MetricType::TIME;
            
            // 执行路径计算
            auto routes = router->findRoutes(
                origin_lng, origin_lat,
                dest_lng, dest_lat,
                metric, alternatives
            );
            
            // 构建响应
            json response;
            response["routes"] = json::array();
            
            for (const auto& route : routes) {
                json route_obj;
                route_obj["duration"] = route.duration;
                route_obj["distance"] = route.distance;
                route_obj["geometry"] = route.encoded_polyline;
                
                // 添加关键路径点
                json waypoints = json::array();
                for (const auto& point : route.waypoints) {
                    waypoints.push_back({
                        {"lng", point.longitude},
                        {"lat", point.latitude}
                    });
                }
                route_obj["waypoints"] = waypoints;
                
                response["routes"].push_back(route_obj);
            }
            
            // 计算耗时
            auto end_time = high_resolution_clock::now();
            auto elapsed = duration_cast<milliseconds>(end_time - start_time);
            
            response["request_id"] = generateRequestId();
            response["elapsed_ms"] = elapsed.count();
            response["status"] = "success";
            
            // 记录性能指标
            monitor->recordRequest(elapsed.count(), routes.size());
            
            return response;
            
        } catch (const std::exception& e) {
            json error_response;
            error_response["status"] = "error";
            error_response["message"] = e.what();
            
            auto end_time = high_resolution_clock::now();
            auto elapsed = duration_cast<milliseconds>(end_time - start_time);
            error_response["elapsed_ms"] = elapsed.count();
            
            return error_response;
        }
    }
    
private:
    std::string generateRequestId() {
        static std::atomic<uint64_t> counter{0};
        auto timestamp = duration_cast<milliseconds>(
            system_clock::now().time_since_epoch()
        ).count();
        return std::to_string(timestamp) + "-" + 
               std::to_string(counter.fetch_add(1));
    }
};

int main(int argc, char* argv[]) {
    if (argc < 2) {
        std::cerr << "Usage: " << argv[0] << " <data_path>" << std::endl;
        return 1;
    }
    
    const std::string data_path = argv[1];
    const int port = 8080;
    const int num_threads = std::thread::hardware_concurrency();
    
    std::cout << "Initializing CRP Router Service..." << std::endl;
    std::cout << "Data path: " << data_path << std::endl;
    std::cout << "Port: " << port << std::endl;
    std::cout << "Threads: " << num_threads << std::endl;
    
    // 初始化路由服务
    RouteService service(data_path);
    
    // 创建HTTP服务器
    httplib::Server server;
    server.set_payload_max_length(1024 * 1024); // 1MB
    
    // 设置路由端点
    server.Post("/route/calculate", 
        [&service](const httplib::Request& req, httplib::Response& res) {
            try {
                auto request = json::parse(req.body);
                auto response = service.handleRouteRequest(request);
                
                res.set_content(response.dump(), "application/json");
                res.status = 200;
            } catch (const json::parse_error& e) {
                json error;
                error["status"] = "error";
                error["message"] = "Invalid JSON format";
                res.set_content(error.dump(), "application/json");
                res.status = 400;
            }
        }
    );
    
    // 健康检查端点
    server.Get("/health", 
        [](const httplib::Request& req, httplib::Response& res) {
            json health;
            health["status"] = "healthy";
            health["timestamp"] = duration_cast<seconds>(
                system_clock::now().time_since_epoch()
            ).count();
            res.set_content(health.dump(), "application/json");
        }
    );
    
    // 性能指标端点
    server.Get("/metrics",
        [&service](const httplib::Request& req, httplib::Response& res) {
            // 这里应该返回Prometheus格式的metrics
            res.set_content(
                "# HELP crp_requests_total Total number of requests\n"
                "# TYPE crp_requests_total counter\n"
                "crp_requests_total 0\n"
                "# HELP crp_request_duration_seconds Request duration\n"
                "# TYPE crp_request_duration_seconds histogram\n",
                "text/plain"
            );
        }
    );
    
    std::cout << "Starting HTTP server on port " << port << "..." << std::endl;
    
    // 启动服务器
    if (!server.listen("0.0.0.0", port)) {
        std::cerr << "Failed to start server" << std::endl;
        return 1;
    }
    
    return 0;
}