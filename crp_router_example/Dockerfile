# 构建阶段
FROM ubuntu:22.04 AS builder

# 安装构建依赖
RUN apt-get update && apt-get install -y \
    build-essential \
    cmake \
    git \
    wget \
    libomp-dev \
    && rm -rf /var/lib/apt/lists/*

# 设置工作目录
WORKDIR /build

# 复制源代码
COPY . .

# 构建项目
RUN mkdir build && cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release .. && \
    make -j$(nproc)

# 运行阶段
FROM ubuntu:22.04

# 安装运行时依赖
RUN apt-get update && apt-get install -y \
    libomp5 \
    && rm -rf /var/lib/apt/lists/*

# 创建非root用户
RUN useradd -m -s /bin/bash crprouter

# 复制构建产物
COPY --from=builder /build/build/crp_router_service /usr/local/bin/
COPY --from=builder /build/data /data

# 设置权限
RUN chown -R crprouter:crprouter /data

# 切换用户
USER crprouter

# 设置环境变量
ENV OMP_NUM_THREADS=8
ENV OVERLAY_GRAPH_PATH=/data/overlay
ENV CACHE_SIZE=4GB

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# 启动命令
CMD ["crp_router_service", "/data"]