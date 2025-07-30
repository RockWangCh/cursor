package crp

import (
	"container/heap"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"route-planning/internal/config"
	"route-planning/pkg/geom"
	"route-planning/pkg/types"
)

// CRP引擎
type Engine struct {
	graph           *Graph
	contractedGraph *ContractedGraph
	spatialIndex    *SpatialIndex
	config          *config.GraphConfig
	
	// 缓存
	pathCache       sync.Map
	nodeImportance  []float64
	
	// 统计信息
	stats           *Statistics
}

// 新建CRP引擎
func NewEngine(cfg *config.GraphConfig) (*Engine, error) {
	engine := &Engine{
		config: cfg,
		stats:  NewStatistics(),
	}

	// 加载图数据
	if err := engine.loadGraph(); err != nil {
		return nil, fmt.Errorf("failed to load graph: %w", err)
	}

	// 预处理收缩层次
	if err := engine.preprocessContractionHierarchies(); err != nil {
		return nil, fmt.Errorf("failed to preprocess CH: %w", err)
	}

	// 构建空间索引
	if err := engine.buildSpatialIndex(); err != nil {
		return nil, fmt.Errorf("failed to build spatial index: %w", err)
	}

	return engine, nil
}

// 图结构
type Graph struct {
	Nodes    []types.Node
	Edges    []types.Edge
	AdjList  [][]EdgeRef  // 邻接表，存储边的索引
	mutex    sync.RWMutex
}

// 收缩层次图
type ContractedGraph struct {
	NodeLevels   []int            // 节点层次
	Shortcuts    []Shortcut       // 快捷方式
	UpEdges      [][]EdgeRef      // 向上边
	DownEdges    [][]EdgeRef      // 向下边
	ShortcutIdx  map[EdgeKey]int  // 快捷方式索引
}

// 快捷方式
type Shortcut struct {
	From     types.NodeID
	To       types.NodeID
	Weight   float64
	Distance int32
	Time     int32
	Via      types.NodeID  // 经过的中间节点
}

// 边引用
type EdgeRef struct {
	Index  int32
	Weight float64
}

// 空间索引
type SpatialIndex struct {
	rtree    *RTree
	gridSize float64
}

// 加载图数据
func (e *Engine) loadGraph() error {
	// 从数据库或文件加载四维图新数据
	// 这里简化为示例实现
	
	e.graph = &Graph{
		Nodes:   make([]types.Node, 0),
		Edges:   make([]types.Edge, 0),
		AdjList: make([][]EdgeRef, 0),
	}

	// TODO: 实际加载逻辑
	// 1. 从PostgreSQL加载节点和边数据
	// 2. 构建邻接表
	// 3. 验证数据完整性

	return nil
}

// 预处理收缩层次
func (e *Engine) preprocessContractionHierarchies() error {
	nodeCount := len(e.graph.Nodes)
	
	e.contractedGraph = &ContractedGraph{
		NodeLevels:  make([]int, nodeCount),
		Shortcuts:   make([]Shortcut, 0),
		UpEdges:     make([][]EdgeRef, nodeCount),
		DownEdges:   make([][]EdgeRef, nodeCount),
		ShortcutIdx: make(map[EdgeKey]int),
	}

	// 计算节点重要性
	e.calculateNodeImportance()

	// 创建按重要性排序的节点队列
	nodeQueue := make([]NodeImportance, nodeCount)
	for i := 0; i < nodeCount; i++ {
		nodeQueue[i] = NodeImportance{
			NodeID:     types.NodeID(i),
			Importance: e.nodeImportance[i],
		}
	}
	sort.Slice(nodeQueue, func(i, j int) bool {
		return nodeQueue[i].Importance < nodeQueue[j].Importance
	})

	// 按重要性顺序收缩节点
	for level, nodeInfo := range nodeQueue {
		e.contractedGraph.NodeLevels[nodeInfo.NodeID] = level
		if err := e.contractNode(nodeInfo.NodeID, level); err != nil {
			return fmt.Errorf("failed to contract node %d: %w", nodeInfo.NodeID, err)
		}
	}

	return nil
}

// 计算节点重要性
func (e *Engine) calculateNodeImportance() {
	nodeCount := len(e.graph.Nodes)
	e.nodeImportance = make([]float64, nodeCount)

	for i := 0; i < nodeCount; i++ {
		nodeID := types.NodeID(i)
		
		// 重要性因子：
		// 1. 边差值 (EdgeDifference)
		// 2. 删除的边数 (DeletedNeighbors) 
		// 3. 查询频率 (SearchSpace)
		
		edgeDiff := e.calculateEdgeDifference(nodeID)
		deletedNeighbors := len(e.graph.AdjList[i])
		searchSpace := e.estimateSearchSpace(nodeID)
		
		// 加权计算重要性
		importance := float64(edgeDiff)*1.0 + 
					float64(deletedNeighbors)*2.0 + 
					float64(searchSpace)*0.5
		
		e.nodeImportance[i] = importance
	}
}

// 计算边差值
func (e *Engine) calculateEdgeDifference(nodeID types.NodeID) int {
	neighbors := e.getNeighbors(nodeID)
	shortcuts := 0
	
	// 计算收缩此节点需要添加的快捷方式数量
	for i := 0; i < len(neighbors); i++ {
		for j := i + 1; j < len(neighbors); j++ {
			if e.needsShortcut(neighbors[i], neighbors[j], nodeID) {
				shortcuts++
			}
		}
	}
	
	return shortcuts - len(neighbors)
}

// 估算搜索空间
func (e *Engine) estimateSearchSpace(nodeID types.NodeID) int {
	// 使用局部搜索估算影响范围
	visited := make(map[types.NodeID]bool)
	queue := []types.NodeID{nodeID}
	
	hops := 0
	maxHops := 3
	
	for len(queue) > 0 && hops < maxHops {
		nextQueue := make([]types.NodeID, 0)
		
		for _, current := range queue {
			if visited[current] {
				continue
			}
			visited[current] = true
			
			for _, edgeRef := range e.graph.AdjList[current] {
				edge := e.graph.Edges[edgeRef.Index]
				neighbor := edge.To
				if !visited[neighbor] {
					nextQueue = append(nextQueue, neighbor)
				}
			}
		}
		
		queue = nextQueue
		hops++
	}
	
	return len(visited)
}

// 收缩节点
func (e *Engine) contractNode(nodeID types.NodeID, level int) error {
	neighbors := e.getNeighbors(nodeID)
	
	// 为每对邻居创建必要的快捷方式
	for i := 0; i < len(neighbors); i++ {
		for j := i + 1; j < len(neighbors); j++ {
			from, to := neighbors[i], neighbors[j]
			
			if e.needsShortcut(from, to, nodeID) {
				shortcut := e.createShortcut(from, to, nodeID)
				e.contractedGraph.Shortcuts = append(e.contractedGraph.Shortcuts, shortcut)
				
				// 更新索引
				key := EdgeKey{From: from, To: to}
				e.contractedGraph.ShortcutIdx[key] = len(e.contractedGraph.Shortcuts) - 1
				
				// 添加到上/下边列表
				if e.contractedGraph.NodeLevels[from] < level {
					e.contractedGraph.UpEdges[from] = append(e.contractedGraph.UpEdges[from], 
						EdgeRef{Index: int32(len(e.contractedGraph.Shortcuts) - 1), Weight: shortcut.Weight})
				}
				if e.contractedGraph.NodeLevels[to] < level {
					e.contractedGraph.DownEdges[to] = append(e.contractedGraph.DownEdges[to], 
						EdgeRef{Index: int32(len(e.contractedGraph.Shortcuts) - 1), Weight: shortcut.Weight})
				}
			}
		}
	}
	
	return nil
}

// 多路径查询
func (e *Engine) FindMultiplePaths(start, end geom.Point, weightType types.WeightType, maxPaths int) (*types.MultiPathResult, error) {
	startTime := time.Now()
	
	// 坐标转换和最近节点查找
	startNode, err := e.findNearestNode(start)
	if err != nil {
		return nil, fmt.Errorf("failed to find start node: %w", err)
	}
	
	endNode, err := e.findNearestNode(end)
	if err != nil {
		return nil, fmt.Errorf("failed to find end node: %w", err)
	}
	
	// 使用改进的Yen's算法计算多条路径
	paths, err := e.yensKShortestPaths(startNode, endNode, weightType, maxPaths)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate paths: %w", err)
	}
	
	// 构建结果
	result := &types.MultiPathResult{
		Paths:          paths,
		StartNode:      startNode,
		EndNode:        endNode,
		WeightType:     weightType,
		ProcessingTime: time.Since(startTime).Milliseconds(),
		CacheHit:       false,
	}
	
	// 更新统计信息
	e.stats.TotalQueries++
	e.stats.AverageResponseTime = (e.stats.AverageResponseTime + result.ProcessingTime) / 2
	
	return result, nil
}

// Yen's K最短路径算法
func (e *Engine) yensKShortestPaths(start, end types.NodeID, weightType types.WeightType, k int) ([]types.Path, error) {
	// 计算最短路径
	firstPath, err := e.bidirectionalDijkstra(start, end, weightType)
	if err != nil {
		return nil, err
	}
	
	if firstPath == nil {
		return []types.Path{}, nil
	}
	
	paths := []types.Path{*firstPath}
	candidates := NewPathHeap()
	
	for i := 1; i < k; i++ {
		// 为每个候选路径生成变体
		for j := 0; j < len(paths[i-1].Nodes)-1; j++ {
			// 临时移除边
			spurNode := paths[i-1].Nodes[j]
			rootPath := paths[i-1].Nodes[:j+1]
			
			// 移除根路径中的边
			removedEdges := e.temporaryRemoveEdges(rootPath, paths)
			
			// 计算从spur节点到终点的最短路径
			spurPath, err := e.bidirectionalDijkstra(spurNode, end, weightType)
			if err == nil && spurPath != nil {
				// 组合根路径和spur路径
				candidatePath := e.combinePaths(rootPath, spurPath)
				if candidatePath != nil {
					heap.Push(candidates, candidatePath)
				}
			}
			
			// 恢复移除的边
			e.restoreEdges(removedEdges)
		}
		
		if candidates.Len() == 0 {
			break
		}
		
		// 取出最短的候选路径
		nextPath := heap.Pop(candidates).(*types.Path)
		paths = append(paths, *nextPath)
		
		// 去重和过滤相似路径
		paths = e.filterSimilarPaths(paths)
	}
	
	return paths, nil
}

// 双向Dijkstra搜索
func (e *Engine) bidirectionalDijkstra(start, end types.NodeID, weightType types.WeightType) (*types.Path, error) {
	// 前向搜索
	forwardQueue := NewPriorityQueue()
	forwardDist := make(map[types.NodeID]float64)
	forwardPrev := make(map[types.NodeID]types.NodeID)
	
	// 后向搜索
	backwardQueue := NewPriorityQueue()
	backwardDist := make(map[types.NodeID]float64)
	backwardPrev := make(map[types.NodeID]types.NodeID)
	
	// 初始化
	heap.Push(forwardQueue, &QueueItem{NodeID: start, Distance: 0})
	forwardDist[start] = 0
	
	heap.Push(backwardQueue, &QueueItem{NodeID: end, Distance: 0})
	backwardDist[end] = 0
	
	bestPath := math.Inf(1)
	meetingNode := types.NodeID(-1)
	
	// 双向搜索主循环
	for forwardQueue.Len() > 0 && backwardQueue.Len() > 0 {
		// 前向步骤
		if forwardQueue.Len() > 0 {
			current := heap.Pop(forwardQueue).(*QueueItem)
			
			if current.Distance > bestPath {
				break
			}
			
			// 检查是否与后向搜索相遇
			if backwardDist[current.NodeID] != 0 {
				totalDist := current.Distance + backwardDist[current.NodeID]
				if totalDist < bestPath {
					bestPath = totalDist
					meetingNode = current.NodeID
				}
			}
			
			// 扩展邻居（使用收缩层次优化）
			e.expandNeighbors(current.NodeID, current.Distance, weightType, 
				forwardDist, forwardPrev, forwardQueue, true)
		}
		
		// 后向步骤
		if backwardQueue.Len() > 0 {
			current := heap.Pop(backwardQueue).(*QueueItem)
			
			if current.Distance > bestPath {
				break
			}
			
			// 检查是否与前向搜索相遇
			if forwardDist[current.NodeID] != 0 {
				totalDist := current.Distance + forwardDist[current.NodeID]
				if totalDist < bestPath {
					bestPath = totalDist
					meetingNode = current.NodeID
				}
			}
			
			// 扩展邻居
			e.expandNeighbors(current.NodeID, current.Distance, weightType, 
				backwardDist, backwardPrev, backwardQueue, false)
		}
	}
	
	if meetingNode == types.NodeID(-1) {
		return nil, nil // 无路径
	}
	
	// 重构路径
	path := e.reconstructBidirectionalPath(start, end, meetingNode, forwardPrev, backwardPrev)
	return path, nil
}

// 扩展邻居节点
func (e *Engine) expandNeighbors(nodeID types.NodeID, currentDist float64, weightType types.WeightType,
	distances map[types.NodeID]float64, predecessors map[types.NodeID]types.NodeID, 
	queue *PriorityQueue, isForward bool) {
	
	var edges []EdgeRef
	
	if isForward {
		// 前向搜索：使用向上边
		edges = e.contractedGraph.UpEdges[nodeID]
		// 也包括原始边
		edges = append(edges, e.graph.AdjList[nodeID]...)
	} else {
		// 后向搜索：使用向下边
		edges = e.contractedGraph.DownEdges[nodeID]
		// 反向遍历原始边
		for _, edgeRef := range e.graph.AdjList[nodeID] {
			edge := e.graph.Edges[edgeRef.Index]
			if edge.To == nodeID { // 反向边
				edges = append(edges, edgeRef)
			}
		}
	}
	
	for _, edgeRef := range edges {
		var neighbor types.NodeID
		var weight float64
		
		if edgeRef.Index < int32(len(e.graph.Edges)) {
			// 原始边
			edge := e.graph.Edges[edgeRef.Index]
			neighbor = edge.To
			weight = e.calculateEdgeWeight(&edge, weightType)
		} else {
			// 快捷方式
			shortcutIdx := edgeRef.Index - int32(len(e.graph.Edges))
			shortcut := e.contractedGraph.Shortcuts[shortcutIdx]
			neighbor = shortcut.To
			weight = e.calculateShortcutWeight(&shortcut, weightType)
		}
		
		newDist := currentDist + weight
		
		if oldDist, exists := distances[neighbor]; !exists || newDist < oldDist {
			distances[neighbor] = newDist
			predecessors[neighbor] = nodeID
			
			heap.Push(queue, &QueueItem{
				NodeID:   neighbor,
				Distance: newDist,
			})
		}
	}
}

// 计算边权重
func (e *Engine) calculateEdgeWeight(edge *types.Edge, weightType types.WeightType) float64 {
	switch weightType {
	case types.WeightTypeTime:
		return float64(edge.Time)
	case types.WeightTypeDistance:
		return float64(edge.Length)
	default:
		return float64(edge.Length) // 默认使用距离
	}
}

// 计算快捷方式权重
func (e *Engine) calculateShortcutWeight(shortcut *Shortcut, weightType types.WeightType) float64 {
	switch weightType {
	case types.WeightTypeTime:
		return float64(shortcut.Time)
	case types.WeightTypeDistance:
		return float64(shortcut.Distance)
	default:
		return float64(shortcut.Distance)
	}
}

// 重构双向搜索路径
func (e *Engine) reconstructBidirectionalPath(start, end, meeting types.NodeID, 
	forwardPrev, backwardPrev map[types.NodeID]types.NodeID) *types.Path {
	
	// 从起点到会合点的路径
	forwardPath := make([]types.NodeID, 0)
	current := meeting
	for current != start {
		forwardPath = append(forwardPath, current)
		current = forwardPrev[current]
	}
	forwardPath = append(forwardPath, start)
	
	// 反转前向路径
	for i, j := 0, len(forwardPath)-1; i < j; i, j = i+1, j-1 {
		forwardPath[i], forwardPath[j] = forwardPath[j], forwardPath[i]
	}
	
	// 从会合点到终点的路径
	backwardPath := make([]types.NodeID, 0)
	current = meeting
	for current != end {
		current = backwardPrev[current]
		backwardPath = append(backwardPath, current)
	}
	
	// 组合路径
	fullPath := append(forwardPath, backwardPath...)
	
	// 计算路径总权重和其他属性
	totalWeight, totalDistance, totalTime := e.calculatePathMetrics(fullPath)
	
	return &types.Path{
		Nodes:       fullPath,
		Weight:      totalWeight,
		Distance:    totalDistance,
		Time:        totalTime,
		Geometry:    e.buildPathGeometry(fullPath),
	}
}

// 查找最近节点
func (e *Engine) findNearestNode(point geom.Point) (types.NodeID, error) {
	return e.spatialIndex.FindNearest(point)
}

// 辅助结构和方法
type NodeImportance struct {
	NodeID     types.NodeID
	Importance float64
}

type EdgeKey struct {
	From, To types.NodeID
}

type QueueItem struct {
	NodeID   types.NodeID
	Distance float64
	index    int
}

// 优先级队列实现
type PriorityQueue []*QueueItem

func NewPriorityQueue() *PriorityQueue {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	return &pq
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Distance < pq[j].Distance
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*QueueItem)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// 统计信息
type Statistics struct {
	TotalQueries        int64
	CacheHits          int64
	AverageResponseTime int64
	MaxResponseTime    int64
	MinResponseTime    int64
}

func NewStatistics() *Statistics {
	return &Statistics{
		MinResponseTime: math.MaxInt64,
	}
}

// 其他辅助方法的实现...
func (e *Engine) getNeighbors(nodeID types.NodeID) []types.NodeID {
	// 实现获取邻居节点逻辑
	return nil
}

func (e *Engine) needsShortcut(from, to, via types.NodeID) bool {
	// 实现是否需要快捷方式的判断逻辑
	return false
}

func (e *Engine) createShortcut(from, to, via types.NodeID) Shortcut {
	// 实现创建快捷方式的逻辑
	return Shortcut{}
}

func (e *Engine) buildSpatialIndex() error {
	// 实现空间索引构建逻辑
	return nil
}

func (e *Engine) temporaryRemoveEdges(rootPath []types.NodeID, paths []types.Path) []types.Edge {
	// 实现临时移除边的逻辑
	return nil
}

func (e *Engine) restoreEdges(edges []types.Edge) {
	// 实现恢复边的逻辑
}

func (e *Engine) combinePaths(rootPath []types.NodeID, spurPath *types.Path) *types.Path {
	// 实现路径组合逻辑
	return nil
}

func (e *Engine) filterSimilarPaths(paths []types.Path) []types.Path {
	// 实现相似路径过滤逻辑
	return paths
}

func (e *Engine) calculatePathMetrics(path []types.NodeID) (float64, int32, int32) {
	// 实现路径指标计算逻辑
	return 0, 0, 0
}

func (e *Engine) buildPathGeometry(path []types.NodeID) string {
	// 实现路径几何构建逻辑（编码为字符串）
	return ""
}

type PathHeap []*types.Path

func NewPathHeap() *PathHeap {
	ph := make(PathHeap, 0)
	heap.Init(&ph)
	return &ph
}

func (ph PathHeap) Len() int { return len(ph) }
func (ph PathHeap) Less(i, j int) bool { return ph[i].Weight < ph[j].Weight }
func (ph PathHeap) Swap(i, j int) { ph[i], ph[j] = ph[j], ph[i] }

func (ph *PathHeap) Push(x interface{}) {
	*ph = append(*ph, x.(*types.Path))
}

func (ph *PathHeap) Pop() interface{} {
	old := *ph
	n := len(old)
	item := old[n-1]
	*ph = old[0 : n-1]
	return item
}