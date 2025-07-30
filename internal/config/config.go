package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"
)

// 系统配置
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Cache    CacheConfig    `json:"cache"`
	Graph    GraphConfig    `json:"graph"`
	Monitor  MonitorConfig  `json:"monitor"`
}

// 服务器配置
type ServerConfig struct {
	Address      string `json:"address"`
	Mode         string `json:"mode"`          // debug, release
	ReadTimeout  int    `json:"read_timeout"`  // 秒
	WriteTimeout int    `json:"write_timeout"` // 秒
	IdleTimeout  int    `json:"idle_timeout"`  // 秒
}

// 数据库配置
type DatabaseConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Database     string `json:"database"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	MaxOpenConns int    `json:"max_open_conns"`
	MaxIdleConns int    `json:"max_idle_conns"`
	MaxLifetime  int    `json:"max_lifetime"` // 秒
}

// 缓存配置
type CacheConfig struct {
	Redis RedisConfig `json:"redis"`
	L1    L1Config    `json:"l1"`
}

// Redis配置
type RedisConfig struct {
	Addresses    []string `json:"addresses"`
	Password     string   `json:"password"`
	DB           int      `json:"db"`
	PoolSize     int      `json:"pool_size"`
	MinIdleConns int      `json:"min_idle_conns"`
	MaxRetries   int      `json:"max_retries"`
	DialTimeout  int      `json:"dial_timeout"`  // 毫秒
	ReadTimeout  int      `json:"read_timeout"`  // 毫秒
	WriteTimeout int      `json:"write_timeout"` // 毫秒
}

// L1缓存配置
type L1Config struct {
	MaxSize   int64         `json:"max_size"`   // 最大条目数
	TTL       time.Duration `json:"ttl"`        // 生存时间
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// 图配置
type GraphConfig struct {
	DataPath      string `json:"data_path"`
	IndexPath     string `json:"index_path"`
	PreloadMemory bool   `json:"preload_memory"`
	
	// 预处理配置
	Preprocessing PreprocessingConfig `json:"preprocessing"`
	
	// 空间索引配置
	SpatialIndex SpatialIndexConfig `json:"spatial_index"`
}

// 预处理配置
type PreprocessingConfig struct {
	Enabled        bool    `json:"enabled"`
	NodeImportance NodeImportanceConfig `json:"node_importance"`
	Parallelism    int     `json:"parallelism"`
}

// 节点重要性配置
type NodeImportanceConfig struct {
	EdgeDiffWeight     float64 `json:"edge_diff_weight"`
	NeighborWeight     float64 `json:"neighbor_weight"`
	SearchSpaceWeight  float64 `json:"search_space_weight"`
}

// 空间索引配置
type SpatialIndexConfig struct {
	Type     string  `json:"type"`      // rtree, grid
	GridSize float64 `json:"grid_size"` // 网格大小(度)
	MaxDepth int     `json:"max_depth"` // R-Tree最大深度
	MinItems int     `json:"min_items"` // 最小条目数
}

// 监控配置
type MonitorConfig struct {
	MetricsPath    string        `json:"metrics_path"`
	HealthPath     string        `json:"health_path"`
	UpdateInterval time.Duration `json:"update_interval"`
}

// 加载配置
func Load() (*Config, error) {
	cfg := getDefaultConfig()
	
	// 从环境变量覆盖配置
	if err := loadFromEnv(cfg); err != nil {
		return nil, fmt.Errorf("failed to load env config: %w", err)
	}
	
	// 从配置文件加载（如果存在）
	configFile := os.Getenv("CONFIG_FILE")
	if configFile != "" {
		if err := loadFromFile(cfg, configFile); err != nil {
			return nil, fmt.Errorf("failed to load config file: %w", err)
		}
	}
	
	// 验证配置
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	
	return cfg, nil
}

// 默认配置
func getDefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:      ":8080",
			Mode:         "debug",
			ReadTimeout:  30,
			WriteTimeout: 30,
			IdleTimeout:  60,
		},
		Database: DatabaseConfig{
			Host:         "localhost",
			Port:         5432,
			Database:     "routes",
			Username:     "postgres",
			Password:     "",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  300,
		},
		Cache: CacheConfig{
			Redis: RedisConfig{
				Addresses:    []string{"localhost:6379"},
				Password:     "",
				DB:           0,
				PoolSize:     10,
				MinIdleConns: 3,
				MaxRetries:   3,
				DialTimeout:  5000,
				ReadTimeout:  3000,
				WriteTimeout: 3000,
			},
			L1: L1Config{
				MaxSize:         10000,
				TTL:             time.Minute * 5,
				CleanupInterval: time.Minute,
			},
		},
		Graph: GraphConfig{
			DataPath:      "./data",
			IndexPath:     "./index",
			PreloadMemory: true,
			Preprocessing: PreprocessingConfig{
				Enabled: true,
				NodeImportance: NodeImportanceConfig{
					EdgeDiffWeight:    1.0,
					NeighborWeight:    2.0,
					SearchSpaceWeight: 0.5,
				},
				Parallelism: 4,
			},
			SpatialIndex: SpatialIndexConfig{
				Type:     "rtree",
				GridSize: 0.001,
				MaxDepth: 16,
				MinItems: 8,
			},
		},
		Monitor: MonitorConfig{
			MetricsPath:    "/metrics",
			HealthPath:     "/health",
			UpdateInterval: time.Second * 30,
		},
	}
}

// 从环境变量加载配置
func loadFromEnv(cfg *Config) error {
	// 服务器配置
	if addr := os.Getenv("SERVER_ADDRESS"); addr != "" {
		cfg.Server.Address = addr
	}
	if mode := os.Getenv("SERVER_MODE"); mode != "" {
		cfg.Server.Mode = mode
	}
	if timeout := os.Getenv("SERVER_READ_TIMEOUT"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil {
			cfg.Server.ReadTimeout = val
		}
	}
	
	// 数据库配置
	if host := os.Getenv("DB_HOST"); host != "" {
		cfg.Database.Host = host
	}
	if port := os.Getenv("DB_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			cfg.Database.Port = val
		}
	}
	if db := os.Getenv("DB_NAME"); db != "" {
		cfg.Database.Database = db
	}
	if user := os.Getenv("DB_USER"); user != "" {
		cfg.Database.Username = user
	}
	if pass := os.Getenv("DB_PASSWORD"); pass != "" {
		cfg.Database.Password = pass
	}
	
	// Redis配置
	if redisURL := os.Getenv("REDIS_URL"); redisURL != "" {
		cfg.Cache.Redis.Addresses = []string{redisURL}
	}
	if redisPass := os.Getenv("REDIS_PASSWORD"); redisPass != "" {
		cfg.Cache.Redis.Password = redisPass
	}
	
	// 图数据配置
	if dataPath := os.Getenv("GRAPH_DATA_PATH"); dataPath != "" {
		cfg.Graph.DataPath = dataPath
	}
	if indexPath := os.Getenv("GRAPH_INDEX_PATH"); indexPath != "" {
		cfg.Graph.IndexPath = indexPath
	}
	
	return nil
}

// 从文件加载配置
func loadFromFile(cfg *Config, filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	
	return json.Unmarshal(data, cfg)
}

// 验证配置
func validate(cfg *Config) error {
	// 验证服务器配置
	if cfg.Server.Address == "" {
		return fmt.Errorf("server address cannot be empty")
	}
	
	// 验证数据库配置
	if cfg.Database.Host == "" {
		return fmt.Errorf("database host cannot be empty")
	}
	if cfg.Database.Database == "" {
		return fmt.Errorf("database name cannot be empty")
	}
	
	// 验证Redis配置
	if len(cfg.Cache.Redis.Addresses) == 0 {
		return fmt.Errorf("redis addresses cannot be empty")
	}
	
	// 验证图配置
	if cfg.Graph.DataPath == "" {
		return fmt.Errorf("graph data path cannot be empty")
	}
	
	return nil
}

// 保存配置到文件
func (c *Config) SaveToFile(filename string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}

// 获取数据库连接字符串
func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.Database.Username,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.Database,
	)
}