#ifndef PERFORMANCE_MONITOR_HPP
#define PERFORMANCE_MONITOR_HPP

#include <atomic>
#include <chrono>
#include <vector>
#include <mutex>
#include <algorithm>
#include <numeric>
#include <iomanip>
#include <sstream>

class PerformanceMonitor {
private:
    // 请求计数
    std::atomic<uint64_t> total_requests{0};
    std::atomic<uint64_t> successful_requests{0};
    std::atomic<uint64_t> failed_requests{0};
    
    // 响应时间统计（毫秒）
    mutable std::mutex latency_mutex;
    std::vector<double> latencies;
    const size_t max_samples = 10000;
    
    // 路径数量统计
    std::atomic<uint64_t> total_routes_computed{0};
    
    // 时间戳
    std::chrono::steady_clock::time_point start_time;
    
public:
    PerformanceMonitor() 
        : start_time(std::chrono::steady_clock::now()) {
        latencies.reserve(max_samples);
    }
    
    // 记录请求
    void recordRequest(double latency_ms, int routes_found, bool success = true) {
        total_requests.fetch_add(1);
        
        if (success) {
            successful_requests.fetch_add(1);
            total_routes_computed.fetch_add(routes_found);
            
            // 记录延迟
            std::lock_guard<std::mutex> lock(latency_mutex);
            if (latencies.size() >= max_samples) {
                // 使用循环缓冲区
                latencies.erase(latencies.begin());
            }
            latencies.push_back(latency_ms);
        } else {
            failed_requests.fetch_add(1);
        }
    }
    
    // 获取统计信息
    struct Stats {
        uint64_t total_requests;
        uint64_t successful_requests;
        uint64_t failed_requests;
        double success_rate;
        double avg_latency_ms;
        double p50_latency_ms;
        double p95_latency_ms;
        double p99_latency_ms;
        double requests_per_second;
        double avg_routes_per_request;
        uint64_t uptime_seconds;
    };
    
    Stats getStats() const {
        Stats stats;
        stats.total_requests = total_requests.load();
        stats.successful_requests = successful_requests.load();
        stats.failed_requests = failed_requests.load();
        
        if (stats.total_requests > 0) {
            stats.success_rate = 100.0 * stats.successful_requests / stats.total_requests;
        } else {
            stats.success_rate = 0.0;
        }
        
        // 计算延迟百分位数
        std::vector<double> sorted_latencies;
        {
            std::lock_guard<std::mutex> lock(latency_mutex);
            sorted_latencies = latencies;
        }
        
        if (!sorted_latencies.empty()) {
            std::sort(sorted_latencies.begin(), sorted_latencies.end());
            
            stats.avg_latency_ms = std::accumulate(
                sorted_latencies.begin(), sorted_latencies.end(), 0.0
            ) / sorted_latencies.size();
            
            stats.p50_latency_ms = percentile(sorted_latencies, 0.50);
            stats.p95_latency_ms = percentile(sorted_latencies, 0.95);
            stats.p99_latency_ms = percentile(sorted_latencies, 0.99);
        } else {
            stats.avg_latency_ms = 0.0;
            stats.p50_latency_ms = 0.0;
            stats.p95_latency_ms = 0.0;
            stats.p99_latency_ms = 0.0;
        }
        
        // 计算吞吐量
        auto now = std::chrono::steady_clock::now();
        auto duration = std::chrono::duration_cast<std::chrono::seconds>(
            now - start_time
        ).count();
        stats.uptime_seconds = duration;
        
        if (duration > 0) {
            stats.requests_per_second = 
                static_cast<double>(stats.total_requests) / duration;
        } else {
            stats.requests_per_second = 0.0;
        }
        
        // 平均每请求的路径数
        if (stats.successful_requests > 0) {
            stats.avg_routes_per_request = 
                static_cast<double>(total_routes_computed.load()) / 
                stats.successful_requests;
        } else {
            stats.avg_routes_per_request = 0.0;
        }
        
        return stats;
    }
    
    // 生成Prometheus格式的指标
    std::string getPrometheusMetrics() const {
        auto stats = getStats();
        std::stringstream ss;
        
        ss << "# HELP crp_requests_total Total number of routing requests\n";
        ss << "# TYPE crp_requests_total counter\n";
        ss << "crp_requests_total " << stats.total_requests << "\n\n";
        
        ss << "# HELP crp_requests_success_total Total number of successful requests\n";
        ss << "# TYPE crp_requests_success_total counter\n";
        ss << "crp_requests_success_total " << stats.successful_requests << "\n\n";
        
        ss << "# HELP crp_requests_failed_total Total number of failed requests\n";
        ss << "# TYPE crp_requests_failed_total counter\n";
        ss << "crp_requests_failed_total " << stats.failed_requests << "\n\n";
        
        ss << "# HELP crp_request_duration_seconds Request latency in seconds\n";
        ss << "# TYPE crp_request_duration_seconds summary\n";
        ss << "crp_request_duration_seconds{quantile=\"0.5\"} " 
           << std::fixed << std::setprecision(6) 
           << stats.p50_latency_ms / 1000.0 << "\n";
        ss << "crp_request_duration_seconds{quantile=\"0.95\"} " 
           << stats.p95_latency_ms / 1000.0 << "\n";
        ss << "crp_request_duration_seconds{quantile=\"0.99\"} " 
           << stats.p99_latency_ms / 1000.0 << "\n";
        ss << "crp_request_duration_seconds_sum " 
           << stats.avg_latency_ms * stats.successful_requests / 1000.0 << "\n";
        ss << "crp_request_duration_seconds_count " 
           << stats.successful_requests << "\n\n";
        
        ss << "# HELP crp_routes_computed_total Total number of routes computed\n";
        ss << "# TYPE crp_routes_computed_total counter\n";
        ss << "crp_routes_computed_total " << total_routes_computed.load() << "\n\n";
        
        ss << "# HELP crp_uptime_seconds Service uptime in seconds\n";
        ss << "# TYPE crp_uptime_seconds gauge\n";
        ss << "crp_uptime_seconds " << stats.uptime_seconds << "\n";
        
        return ss.str();
    }
    
    // 重置统计
    void reset() {
        total_requests = 0;
        successful_requests = 0;
        failed_requests = 0;
        total_routes_computed = 0;
        
        std::lock_guard<std::mutex> lock(latency_mutex);
        latencies.clear();
        
        start_time = std::chrono::steady_clock::now();
    }
    
private:
    // 计算百分位数
    double percentile(const std::vector<double>& sorted_data, double p) const {
        if (sorted_data.empty()) return 0.0;
        
        size_t n = sorted_data.size();
        double idx = p * (n - 1);
        size_t lower = static_cast<size_t>(idx);
        size_t upper = lower + 1;
        
        if (upper >= n) return sorted_data[n - 1];
        
        double weight = idx - lower;
        return sorted_data[lower] * (1 - weight) + sorted_data[upper] * weight;
    }
};

#endif // PERFORMANCE_MONITOR_HPP