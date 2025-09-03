"""
CCH算法实际应用示例
演示如何在城市路网中使用CCH算法进行路径规划
"""

from cch_algorithm import CCHGraph, CCHPathfinder
import time
import random

class CityRoadNetwork:
    """城市路网示例"""
    
    def __init__(self):
        self.graph = CCHGraph()
        self.location_names = {}  # 节点ID到位置名称的映射
    
    def build_city_network(self):
        """构建城市路网"""
        print("构建城市路网...")
        
        # 定义城市中的重要地点
        locations = {
            0: "市中心",
            1: "火车站", 
            2: "机场",
            3: "大学",
            4: "医院",
            5: "商业区",
            6: "住宅区A",
            7: "住宅区B",
            8: "工业区",
            9: "港口",
            10: "体育馆",
            11: "公园",
            12: "政府大楼",
            13: "购物中心",
            14: "科技园"
        }
        
        self.location_names = locations
        
        # 定义道路连接和距离（公里）
        roads = [
            # 主要交通枢纽连接
            (0, 1, 3.2),   # 市中心 - 火车站
            (0, 2, 25.5),  # 市中心 - 机场
            (0, 5, 2.1),   # 市中心 - 商业区
            (0, 12, 1.8),  # 市中心 - 政府大楼
            
            # 火车站连接
            (1, 3, 4.5),   # 火车站 - 大学
            (1, 6, 6.2),   # 火车站 - 住宅区A
            (1, 8, 8.1),   # 火车站 - 工业区
            
            # 机场连接
            (2, 8, 12.3),  # 机场 - 工业区
            (2, 9, 18.7),  # 机场 - 港口
            (2, 14, 15.4), # 机场 - 科技园
            
            # 大学周边
            (3, 4, 2.8),   # 大学 - 医院
            (3, 6, 3.5),   # 大学 - 住宅区A
            (3, 11, 1.9),  # 大学 - 公园
            
            # 医院连接
            (4, 5, 4.2),   # 医院 - 商业区
            (4, 7, 5.1),   # 医院 - 住宅区B
            
            # 商业区连接
            (5, 13, 1.5),  # 商业区 - 购物中心
            (5, 10, 3.8),  # 商业区 - 体育馆
            (5, 7, 4.6),   # 商业区 - 住宅区B
            
            # 住宅区连接
            (6, 7, 7.3),   # 住宅区A - 住宅区B
            (6, 11, 2.4),  # 住宅区A - 公园
            (7, 10, 3.1),  # 住宅区B - 体育馆
            (7, 13, 5.2),  # 住宅区B - 购物中心
            
            # 工业区和港口
            (8, 9, 6.8),   # 工业区 - 港口
            (8, 14, 9.2),  # 工业区 - 科技园
            
            # 其他连接
            (9, 14, 22.1), # 港口 - 科技园
            (10, 11, 2.7), # 体育馆 - 公园
            (11, 12, 2.3), # 公园 - 政府大楼
            (12, 13, 3.4), # 政府大楼 - 购物中心
            (13, 14, 11.8) # 购物中心 - 科技园
        ]
        
        # 添加道路到图中（双向道路）
        for from_loc, to_loc, distance in roads:
            self.graph.add_undirected_edge(from_loc, to_loc, distance)
        
        print(f"城市路网构建完成：{len(locations)}个地点，{len(roads)}条道路")
    
    def get_location_name(self, node_id):
        """获取位置名称"""
        return self.location_names.get(node_id, f"位置{node_id}")

def benchmark_algorithms():
    """性能基准测试：比较CCH与传统Dijkstra算法"""
    print("\n=== 性能基准测试 ===")
    
    # 创建大规模随机图
    def create_large_random_graph(num_nodes=1000, num_edges=5000):
        graph = CCHGraph()
        
        # 添加随机边
        for _ in range(num_edges):
            from_node = random.randint(0, num_nodes - 1)
            to_node = random.randint(0, num_nodes - 1)
            if from_node != to_node:
                weight = random.uniform(1.0, 100.0)
                graph.add_edge(from_node, to_node, weight)
        
        return graph
    
    # 传统Dijkstra算法实现
    def dijkstra(graph, start, end):
        import heapq
        distances = {node: float('inf') for node in graph.nodes}
        distances[start] = 0
        pq = [(0, start)]
        previous = {}
        
        while pq:
            current_dist, current = heapq.heappop(pq)
            
            if current == end:
                break
            
            if current_dist > distances[current]:
                continue
            
            for neighbor, weight in graph.edges[current]:
                distance = current_dist + weight
                if distance < distances[neighbor]:
                    distances[neighbor] = distance
                    previous[neighbor] = current
                    heapq.heappush(pq, (distance, neighbor))
        
        return distances[end]
    
    print("创建大规模随机图...")
    large_graph = create_large_random_graph(500, 2000)
    
    # CCH预处理
    print("CCH预处理...")
    start_time = time.time()
    cch_pathfinder = CCHPathfinder(large_graph)
    cch_pathfinder.preprocess()
    preprocess_time = time.time() - start_time
    print(f"CCH预处理时间: {preprocess_time:.3f}秒")
    
    # 测试查询
    test_queries = [(random.choice(list(large_graph.nodes)), 
                    random.choice(list(large_graph.nodes))) 
                   for _ in range(50)]
    
    print("\n执行查询测试...")
    
    # CCH查询测试
    cch_times = []
    cch_results = []
    for start, end in test_queries:
        start_time = time.time()
        distance, path = cch_pathfinder.find_shortest_path(start, end)
        query_time = time.time() - start_time
        cch_times.append(query_time)
        cch_results.append(distance)
    
    # Dijkstra查询测试
    dijkstra_times = []
    dijkstra_results = []
    for start, end in test_queries:
        start_time = time.time()
        distance = dijkstra(large_graph, start, end)
        query_time = time.time() - start_time
        dijkstra_times.append(query_time)
        dijkstra_results.append(distance)
    
    # 结果比较
    avg_cch_time = sum(cch_times) / len(cch_times)
    avg_dijkstra_time = sum(dijkstra_times) / len(dijkstra_times)
    
    print(f"\n查询性能比较（{len(test_queries)}次查询平均）:")
    print(f"CCH查询时间: {avg_cch_time*1000:.3f}毫秒")
    print(f"Dijkstra查询时间: {avg_dijkstra_time*1000:.3f}毫秒")
    print(f"加速比: {avg_dijkstra_time/avg_cch_time:.1f}x")
    
    # 验证结果正确性
    correct_results = sum(1 for c, d in zip(cch_results, dijkstra_results) 
                         if abs(c - d) < 0.001 or (c == float('inf') and d == float('inf')))
    print(f"结果正确性: {correct_results}/{len(test_queries)} ({100*correct_results/len(test_queries):.1f}%)")

def main():
    """主函数：演示CCH算法的使用"""
    print("CCH寻路算法演示")
    print("=" * 50)
    
    # 1. 创建城市路网
    city = CityRoadNetwork()
    city.build_city_network()
    
    # 2. 创建CCH寻路器并预处理
    print("\n开始CCH预处理...")
    pathfinder = CCHPathfinder(city.graph)
    
    start_time = time.time()
    pathfinder.preprocess()
    preprocess_time = time.time() - start_time
    
    print(f"预处理完成，耗时: {preprocess_time:.3f}秒")
    
    # 3. 演示路径查询
    print("\n=== 路径查询演示 ===")
    
    # 常见的出行场景
    scenarios = [
        (6, 1, "从住宅区A到火车站"),
        (3, 2, "从大学到机场"),
        (0, 9, "从市中心到港口"),
        (4, 13, "从医院到购物中心"),
        (8, 11, "从工业区到公园"),
        (14, 5, "从科技园到商业区")
    ]
    
    for start, end, description in scenarios:
        start_time = time.time()
        distance, path = pathfinder.find_shortest_path(start, end)
        query_time = time.time() - start_time
        
        print(f"\n{description}:")
        print(f"  起点: {city.get_location_name(start)}")
        print(f"  终点: {city.get_location_name(end)}")
        print(f"  距离: {distance:.1f}公里")
        print(f"  查询时间: {query_time*1000:.3f}毫秒")
        
        if path:
            path_names = [city.get_location_name(node) for node in path]
            print(f"  路径: {' → '.join(path_names)}")
        else:
            print("  无法到达目标位置")
    
    # 4. 性能基准测试
    benchmark_algorithms()
    
    # 5. 实时查询演示
    print("\n=== 实时查询演示 ===")
    print("可以输入起点和终点编号进行查询（输入-1退出）:")
    
    # 显示所有位置
    print("\n可用位置:")
    for node_id, name in city.location_names.items():
        print(f"  {node_id}: {name}")
    
    while True:
        try:
            start_input = input("\n请输入起点编号: ")
            if start_input == "-1":
                break
            
            end_input = input("请输入终点编号: ")
            if end_input == "-1":
                break
            
            start = int(start_input)
            end = int(end_input)
            
            if start not in city.location_names or end not in city.location_names:
                print("无效的位置编号！")
                continue
            
            start_time = time.time()
            distance, path = pathfinder.find_shortest_path(start, end)
            query_time = time.time() - start_time
            
            print(f"\n查询结果:")
            print(f"  从 {city.get_location_name(start)} 到 {city.get_location_name(end)}")
            print(f"  距离: {distance:.1f}公里")
            print(f"  查询时间: {query_time*1000:.3f}毫秒")
            
            if path:
                path_names = [city.get_location_name(node) for node in path]
                print(f"  路径: {' → '.join(path_names)}")
            
        except ValueError:
            print("请输入有效的数字！")
        except KeyboardInterrupt:
            break
    
    print("\n演示结束，感谢使用CCH寻路算法！")

if __name__ == "__main__":
    main()