// 游戏画布和上下文
const canvas = document.getElementById('gameCanvas');
const ctx = canvas.getContext('2d');

// 游戏配置
const gridSize = 20;
const tileCount = canvas.width / gridSize;

// 游戏状态
let snake = [
    {x: 10, y: 10}
];
let food = {};
let dx = 0;
let dy = 0;
let score = 0;
let highScore = localStorage.getItem('snakeHighScore') || 0;
let gameRunning = true;
let gamePaused = false;

// 初始化游戏
function init() {
    document.getElementById('highScore').textContent = highScore;
    generateFood();
    gameLoop();
}

// 生成食物
function generateFood() {
    food = {
        x: Math.floor(Math.random() * tileCount),
        y: Math.floor(Math.random() * tileCount)
    };
    
    // 确保食物不在蛇身上
    for (let segment of snake) {
        if (segment.x === food.x && segment.y === food.y) {
            generateFood();
            return;
        }
    }
}

// 绘制游戏元素
function draw() {
    // 清空画布
    ctx.fillStyle = 'black';
    ctx.fillRect(0, 0, canvas.width, canvas.height);
    
    // 绘制蛇
    ctx.fillStyle = '#4CAF50';
    for (let segment of snake) {
        ctx.fillRect(segment.x * gridSize, segment.y * gridSize, gridSize - 2, gridSize - 2);
    }
    
    // 绘制蛇头（不同颜色）
    if (snake.length > 0) {
        ctx.fillStyle = '#8BC34A';
        ctx.fillRect(snake[0].x * gridSize, snake[0].y * gridSize, gridSize - 2, gridSize - 2);
    }
    
    // 绘制食物
    ctx.fillStyle = '#FF5722';
    ctx.beginPath();
    ctx.arc(
        food.x * gridSize + gridSize / 2,
        food.y * gridSize + gridSize / 2,
        gridSize / 2 - 1,
        0,
        2 * Math.PI
    );
    ctx.fill();
    
    // 绘制网格线（可选）
    ctx.strokeStyle = '#333';
    ctx.lineWidth = 1;
    for (let i = 0; i <= tileCount; i++) {
        ctx.beginPath();
        ctx.moveTo(i * gridSize, 0);
        ctx.lineTo(i * gridSize, canvas.height);
        ctx.stroke();
        
        ctx.beginPath();
        ctx.moveTo(0, i * gridSize);
        ctx.lineTo(canvas.width, i * gridSize);
        ctx.stroke();
    }
}

// 更新游戏状态
function update() {
    if (!gameRunning || gamePaused) return;
    
    // 移动蛇头
    const head = {x: snake[0].x + dx, y: snake[0].y + dy};
    
    // 检查边界碰撞
    if (head.x < 0 || head.x >= tileCount || head.y < 0 || head.y >= tileCount) {
        gameOver();
        return;
    }
    
    // 检查自身碰撞
    for (let segment of snake) {
        if (head.x === segment.x && head.y === segment.y) {
            gameOver();
            return;
        }
    }
    
    snake.unshift(head);
    
    // 检查是否吃到食物
    if (head.x === food.x && head.y === food.y) {
        score += 10;
        document.getElementById('score').textContent = score;
        generateFood();
        
        // 检查最高分
        if (score > highScore) {
            highScore = score;
            localStorage.setItem('snakeHighScore', highScore);
            document.getElementById('highScore').textContent = highScore;
        }
    } else {
        // 如果没吃到食物，移除尾部
        snake.pop();
    }
}

// 游戏结束
function gameOver() {
    gameRunning = false;
    document.getElementById('finalScore').textContent = score;
    document.getElementById('gameOver').style.display = 'block';
}

// 重新开始游戏
function restartGame() {
    snake = [{x: 10, y: 10}];
    dx = 0;
    dy = 0;
    score = 0;
    document.getElementById('score').textContent = score;
    document.getElementById('gameOver').style.display = 'none';
    gameRunning = true;
    gamePaused = false;
    generateFood();
}

// 切换暂停状态
function togglePause() {
    if (!gameRunning) return;
    gamePaused = !gamePaused;
    document.querySelector('.btn').textContent = gamePaused ? '继续' : '暂停';
}

// 键盘事件处理
document.addEventListener('keydown', (e) => {
    if (!gameRunning) return;
    
    switch(e.key) {
        case 'ArrowUp':
            if (dy !== 1) { // 防止反向移动
                dx = 0;
                dy = -1;
            }
            break;
        case 'ArrowDown':
            if (dy !== -1) {
                dx = 0;
                dy = 1;
            }
            break;
        case 'ArrowLeft':
            if (dx !== 1) {
                dx = -1;
                dy = 0;
            }
            break;
        case 'ArrowRight':
            if (dx !== -1) {
                dx = 1;
                dy = 0;
            }
            break;
        case ' ': // 空格键暂停
            e.preventDefault();
            togglePause();
            break;
    }
});

// 主游戏循环
function gameLoop() {
    update();
    draw();
    setTimeout(gameLoop, 150); // 控制游戏速度
}

// 触摸控制（移动设备支持）
let touchStartX = 0;
let touchStartY = 0;

canvas.addEventListener('touchstart', (e) => {
    e.preventDefault();
    touchStartX = e.touches[0].clientX;
    touchStartY = e.touches[0].clientY;
});

canvas.addEventListener('touchend', (e) => {
    e.preventDefault();
    if (!gameRunning) return;
    
    const touchEndX = e.changedTouches[0].clientX;
    const touchEndY = e.changedTouches[0].clientY;
    
    const deltaX = touchEndX - touchStartX;
    const deltaY = touchEndY - touchStartY;
    
    const minSwipeDistance = 30;
    
    if (Math.abs(deltaX) > Math.abs(deltaY)) {
        // 水平滑动
        if (Math.abs(deltaX) > minSwipeDistance) {
            if (deltaX > 0 && dx !== -1) {
                // 向右
                dx = 1;
                dy = 0;
            } else if (deltaX < 0 && dx !== 1) {
                // 向左
                dx = -1;
                dy = 0;
            }
        }
    } else {
        // 垂直滑动
        if (Math.abs(deltaY) > minSwipeDistance) {
            if (deltaY > 0 && dy !== -1) {
                // 向下
                dx = 0;
                dy = 1;
            } else if (deltaY < 0 && dy !== 1) {
                // 向上
                dx = 0;
                dy = -1;
            }
        }
    }
});

// 防止页面滚动
canvas.addEventListener('touchmove', (e) => {
    e.preventDefault();
});

// 启动游戏
init();