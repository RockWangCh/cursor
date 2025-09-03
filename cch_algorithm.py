"""
CCH (Customizable Contraction Hierarchies) 寻路算法实现
作者：AI Assistant
描述：完整的CCH算法实现，包括预处理和查询阶段
"""

import heapq
from collections import defaultdict, deque
from typing import Dict, List, Tuple, Set, Optional
import math

class CCHGraph:
    """CCH图数据结构"""
    
    def __init__(self):
        # 原始图结构
        self.nodes: Set[int] = set()
        self.edges: Dict[int, List[Tuple[int, float]]] = defaultdict(list)  # {from: [(to, weight), ...]}
        self.reverse_edges: Dict[int, List[Tuple[int, float]]] = defaultdict(list)  # 反向边
        
        # CCH层次结构
        self.node_levels: Dict[int, int] = {}  # 节点层次
        self.shortcuts: Dict[Tuple[int, int], float] = {}  # 快捷边
        self.contracted_neighbors: Dict[int, Set[int]] = defaultdict(set)  # 收缩时的邻居关系
        
        # 预处理状态
        self.is_preprocessed = False
    
    def add_edge(self, from_node: int, to_node: int, weight: float):
        """添加有向边"""
        self.nodes.add(from_node)
        self.nodes.add(to_node)
        self.edges[from_node].append((to_node, weight))
        self.reverse_edges[to_node].append((from_node, weight))
    
    def add_undirected_edge(self, node1: int, node2: int, weight: float):
        """添加无向边"""
        self.add_edge(node1, node2, weight)
        self.add_edge(node2, node1, weight)

class CCHPathfinder:
    """CCH寻路算法实现"""
    
    def __init__(self, graph: CCHGraph):
        self.graph = graph
        self.node_importance: Dict[int, float] = {}
    
    def calculate_node_importance(self, node: int) -> float:
        """
        计算节点重要性得分
        重要性基于：
        1. 边数差异 (Edge Difference)
        2. 删除复杂度 (Deleted Neighbors)
        3. 搜索空间大小 (Search Space Size)
        """
        if node not in self.graph.nodes:
            return float('inf')
        
        # 1. 计算边数差异
        in_edges = len(self.graph.reverse_edges[node])
        out_edges = len(self.graph.edges[node])
        
        # 模拟收缩，计算需要添加的快捷边数量
        shortcuts_needed = 0
        for pred, pred_weight in self.graph.reverse_edges[node]:
            for succ, succ_weight in self.graph.edges[node]:
                if pred != succ:
                    # 检查是否需要添加快捷边
                    direct_path = pred_weight + succ_weight
                    if not self._has_alternative_path(pred, succ, direct_path, node):
                        shortcuts_needed += 1
        
        edge_difference = shortcuts_needed - in_edges - out_edges
        
        # 2. 删除的邻居数量
        deleted_neighbors = len(self.graph.contracted_neighbors[node])
        
        # 3. 搜索空间大小（简化版本）
        search_space = in_edges * out_edges
        
        # 综合得分（权重可调）
        importance = (
            2.0 * edge_difference + 
            1.0 * deleted_neighbors + 
            0.5 * search_space
        )
        
        return importance
    
    def _has_alternative_path(self, start: int, end: int, max_length: float, avoid_node: int) -> bool:
        """检查是否存在不经过指定节点的替代路径"""
        if start == end:
            return True
        
        # 使用Dijkstra算法查找替代路径（限制搜索深度）
        distances = {start: 0.0}
        pq = [(0.0, start)]
        visited = set()
        
        while pq:
            dist, node = heapq.heappop(pq)
            
            if node in visited:
                continue
            visited.add(node)
            
            if node == end:
                return dist <= max_length
            
            if dist > max_length:
                continue
            
            for neighbor, weight in self.graph.edges[node]:
                if neighbor == avoid_node or neighbor in visited:
                    continue
                
                new_dist = dist + weight
                if new_dist <= max_length:
                    if neighbor not in distances or new_dist < distances[neighbor]:
                        distances[neighbor] = new_dist
                        heapq.heappush(pq, (new_dist, neighbor))
        
        return False
    
    def preprocess(self):
        """预处理阶段：构建CCH层次结构"""
        print("开始CCH预处理...")
        
        # 1. 初始化节点重要性
        remaining_nodes = set(self.graph.nodes)
        level = 0
        
        while remaining_nodes:
            print(f"处理层次 {level}，剩余节点数: {len(remaining_nodes)}")
            
            # 2. 计算当前所有节点的重要性
            importance_scores = {}
            for node in remaining_nodes:
                importance_scores[node] = self.calculate_node_importance(node)
            
            # 3. 选择重要性最低的节点进行收缩
            if not importance_scores:
                break
            
            # 选择重要性最低的节点（可以选择多个）
            min_importance = min(importance_scores.values())
            nodes_to_contract = [
                node for node, score in importance_scores.items() 
                if score <= min_importance + 0.1  # 允许小的误差
            ]
            
            # 限制每次收缩的节点数量
            nodes_to_contract = nodes_to_contract[:max(1, len(remaining_nodes) // 10)]
            
            # 4. 收缩选中的节点
            for node in nodes_to_contract:
                self._contract_node(node, level)
                remaining_nodes.remove(node)
            
            level += 1
        
        self.graph.is_preprocessed = True
        print("CCH预处理完成！")
    
    def _contract_node(self, node: int, level: int):
        """收缩指定节点"""
        self.graph.node_levels[node] = level
        
        # 获取所有入边和出边
        in_edges = list(self.graph.reverse_edges[node])
        out_edges = list(self.graph.edges[node])
        
        # 为每对前驱-后继节点添加快捷边
        for pred, pred_weight in in_edges:
            for succ, succ_weight in out_edges:
                if pred != succ and pred != node and succ != node:
                    shortcut_weight = pred_weight + succ_weight
                    
                    # 检查是否需要添加快捷边
                    if not self._has_alternative_path(pred, succ, shortcut_weight, node):
                        # 添加快捷边
                        key = (pred, succ)
                        if key not in self.graph.shortcuts or self.graph.shortcuts[key] > shortcut_weight:
                            self.graph.shortcuts[key] = shortcut_weight
                            
                            # 更新图结构
                            self._add_shortcut_edge(pred, succ, shortcut_weight)
                        
                        # 记录收缩关系
                        self.graph.contracted_neighbors[pred].add(succ)
                        self.graph.contracted_neighbors[succ].add(pred)
        
        # 从图中移除节点（标记为已收缩）
        self.graph.edges[node] = []
        self.graph.reverse_edges[node] = []
    
    def _add_shortcut_edge(self, from_node: int, to_node: int, weight: float):
        """添加快捷边到图结构中"""
        # 更新正向边
        updated = False
        for i, (neighbor, old_weight) in enumerate(self.graph.edges[from_node]):
            if neighbor == to_node:
                if weight < old_weight:
                    self.graph.edges[from_node][i] = (to_node, weight)
                updated = True
                break
        
        if not updated:
            self.graph.edges[from_node].append((to_node, weight))
        
        # 更新反向边
        updated = False
        for i, (neighbor, old_weight) in enumerate(self.graph.reverse_edges[to_node]):
            if neighbor == from_node:
                if weight < old_weight:
                    self.graph.reverse_edges[to_node][i] = (from_node, weight)
                updated = True
                break
        
        if not updated:
            self.graph.reverse_edges[to_node].append((from_node, weight))
    
    def find_shortest_path(self, start: int, end: int) -> Tuple[float, List[int]]:
        """
        查找最短路径
        返回：(距离, 路径节点列表)
        """
        if not self.graph.is_preprocessed:
            raise RuntimeError("图未经过预处理，请先调用preprocess()方法")
        
        if start == end:
            return 0.0, [start]
        
        # 双向搜索
        forward_dist, forward_prev = self._forward_search(start)
        backward_dist, backward_prev = self._backward_search(end)
        
        # 找到最佳会合点
        best_distance = float('inf')
        best_meeting_point = None
        
        for node in forward_dist:
            if node in backward_dist:
                total_dist = forward_dist[node] + backward_dist[node]
                if total_dist < best_distance:
                    best_distance = total_dist
                    best_meeting_point = node
        
        if best_meeting_point is None:
            return float('inf'), []
        
        # 重构路径
        path = self._reconstruct_path(start, end, best_meeting_point, forward_prev, backward_prev)
        
        return best_distance, path
    
    def _forward_search(self, start: int) -> Tuple[Dict[int, float], Dict[int, int]]:
        """前向搜索（向上搜索）"""
        distances = {start: 0.0}
        previous = {}
        pq = [(0.0, start)]
        visited = set()
        
        while pq:
            dist, node = heapq.heappop(pq)
            
            if node in visited:
                continue
            visited.add(node)
            
            node_level = self.graph.node_levels.get(node, 0)
            
            for neighbor, weight in self.graph.edges[node]:
                if neighbor in visited:
                    continue
                
                neighbor_level = self.graph.node_levels.get(neighbor, 0)
                
                # 只使用向上的边
                if neighbor_level >= node_level:
                    new_dist = dist + weight
                    
                    if neighbor not in distances or new_dist < distances[neighbor]:
                        distances[neighbor] = new_dist
                        previous[neighbor] = node
                        heapq.heappush(pq, (new_dist, neighbor))
        
        return distances, previous
    
    def _backward_search(self, end: int) -> Tuple[Dict[int, float], Dict[int, int]]:
        """后向搜索（向下搜索）"""
        distances = {end: 0.0}
        previous = {}
        pq = [(0.0, end)]
        visited = set()
        
        while pq:
            dist, node = heapq.heappop(pq)
            
            if node in visited:
                continue
            visited.add(node)
            
            node_level = self.graph.node_levels.get(node, 0)
            
            for neighbor, weight in self.graph.reverse_edges[node]:
                if neighbor in visited:
                    continue
                
                neighbor_level = self.graph.node_levels.get(neighbor, 0)
                
                # 只使用向下的边
                if neighbor_level <= node_level:
                    new_dist = dist + weight
                    
                    if neighbor not in distances or new_dist < distances[neighbor]:
                        distances[neighbor] = new_dist
                        previous[neighbor] = node
                        heapq.heappush(pq, (new_dist, neighbor))
        
        return distances, previous
    
    def _reconstruct_path(self, start: int, end: int, meeting_point: int, 
                         forward_prev: Dict[int, int], backward_prev: Dict[int, int]) -> List[int]:
        """重构路径"""
        # 从起点到会合点的路径
        forward_path = []
        current = meeting_point
        while current is not None:
            forward_path.append(current)
            current = forward_prev.get(current)
        forward_path.reverse()
        
        # 从会合点到终点的路径
        backward_path = []
        current = backward_prev.get(meeting_point)
        while current is not None:
            backward_path.append(current)
            current = backward_prev.get(current)
        
        # 合并路径
        full_path = forward_path + backward_path
        
        return full_path

def create_sample_graph() -> CCHGraph:
    """创建示例图"""
    graph = CCHGraph()
    
    # 添加节点和边（构建一个简单的路网）
    edges = [
        (0, 1, 4), (0, 2, 2),
        (1, 2, 1), (1, 3, 5),
        (2, 3, 8), (2, 4, 10),
        (3, 4, 2), (3, 5, 6),
        (4, 5, 3)
    ]
    
    for from_node, to_node, weight in edges:
        graph.add_undirected_edge(from_node, to_node, weight)
    
    return graph

if __name__ == "__main__":
    # 创建示例图
    graph = create_sample_graph()
    
    # 创建CCH寻路器
    pathfinder = CCHPathfinder(graph)
    
    # 预处理
    pathfinder.preprocess()
    
    # 查找最短路径
    start, end = 0, 5
    distance, path = pathfinder.find_shortest_path(start, end)
    
    print(f"\n从节点 {start} 到节点 {end} 的最短路径:")
    print(f"距离: {distance}")
    print(f"路径: {' -> '.join(map(str, path))}")
    
    # 测试多个路径
    test_pairs = [(0, 4), (1, 5), (2, 3)]
    print("\n其他路径测试:")
    for s, e in test_pairs:
        dist, path = pathfinder.find_shortest_path(s, e)
        print(f"从 {s} 到 {e}: 距离={dist}, 路径={' -> '.join(map(str, path))}")