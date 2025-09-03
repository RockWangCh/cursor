#!/usr/bin/env python3
"""
CCH算法演示启动脚本
运行此脚本来体验CCH算法的完整功能
"""

import sys
import os

def main():
    """主函数：提供用户交互界面"""
    print("=" * 60)
    print("  CCH (Customizable Contraction Hierarchies) 算法演示")
    print("=" * 60)
    print()
    
    print("请选择演示模式：")
    print("1. 基础算法演示 - 运行简单示例")
    print("2. 城市路网演示 - 完整的城市交通网络模拟")
    print("3. 性能基准测试 - 与传统算法性能对比")
    print("4. 算法原理说明 - 查看详细的算法解释")
    print("5. 退出")
    print()
    
    while True:
        try:
            choice = input("请输入选择 (1-5): ").strip()
            
            if choice == "1":
                print("\n启动基础算法演示...")
                run_basic_demo()
                break
                
            elif choice == "2":
                print("\n启动城市路网演示...")
                run_city_demo()
                break
                
            elif choice == "3":
                print("\n启动性能基准测试...")
                run_benchmark()
                break
                
            elif choice == "4":
                print("\n显示算法原理说明...")
                show_algorithm_explanation()
                break
                
            elif choice == "5":
                print("感谢使用CCH算法演示！")
                sys.exit(0)
                
            else:
                print("无效选择，请输入1-5之间的数字。")
                
        except KeyboardInterrupt:
            print("\n\n程序被用户中断。")
            sys.exit(0)
        except Exception as e:
            print(f"发生错误: {e}")

def run_basic_demo():
    """运行基础算法演示"""
    try:
        from cch_algorithm import create_sample_graph, CCHPathfinder
        
        print("创建示例图...")
        graph = create_sample_graph()
        
        print("初始化CCH寻路器...")
        pathfinder = CCHPathfinder(graph)
        
        print("执行预处理...")
        pathfinder.preprocess()
        
        print("查找最短路径...")
        start, end = 0, 5
        distance, path = pathfinder.find_shortest_path(start, end)
        
        print(f"\n结果:")
        print(f"从节点 {start} 到节点 {end}")
        print(f"最短距离: {distance}")
        print(f"路径: {' → '.join(map(str, path))}")
        
        print("\n基础演示完成！")
        
    except ImportError as e:
        print(f"导入模块失败: {e}")
        print("请确保 cch_algorithm.py 文件存在且可访问。")
    except Exception as e:
        print(f"演示过程中发生错误: {e}")

def run_city_demo():
    """运行城市路网演示"""
    try:
        from cch_example import main as city_main
        city_main()
        
    except ImportError as e:
        print(f"导入模块失败: {e}")
        print("请确保 cch_example.py 文件存在且可访问。")
    except Exception as e:
        print(f"演示过程中发生错误: {e}")

def run_benchmark():
    """运行性能基准测试"""
    try:
        from cch_example import benchmark_algorithms
        benchmark_algorithms()
        
    except ImportError as e:
        print(f"导入模块失败: {e}")
        print("请确保 cch_example.py 文件存在且可访问。")
    except Exception as e:
        print(f"基准测试过程中发生错误: {e}")

def show_algorithm_explanation():
    """显示算法原理说明"""
    try:
        # 读取并显示算法说明文档
        if os.path.exists("cch_pathfinding_guide.md"):
            with open("cch_pathfinding_guide.md", "r", encoding="utf-8") as f:
                content = f.read()
                print("\n" + "=" * 60)
                print(content)
                print("=" * 60)
        else:
            print("算法说明文档不存在。")
            
        # 询问是否查看优化说明
        if input("\n是否查看性能优化说明？(y/n): ").lower().startswith('y'):
            if os.path.exists("cch_optimizations.md"):
                with open("cch_optimizations.md", "r", encoding="utf-8") as f:
                    content = f.read()
                    print("\n" + "=" * 60)
                    print(content)
                    print("=" * 60)
            else:
                print("优化说明文档不存在。")
                
    except Exception as e:
        print(f"读取文档时发生错误: {e}")

def check_dependencies():
    """检查依赖项"""
    required_modules = ['heapq', 'collections', 'typing', 'time', 'random']
    missing_modules = []
    
    for module in required_modules:
        try:
            __import__(module)
        except ImportError:
            missing_modules.append(module)
    
    if missing_modules:
        print(f"缺少必要的模块: {', '.join(missing_modules)}")
        print("请安装缺少的模块后重试。")
        return False
    
    return True

if __name__ == "__main__":
    # 检查依赖项
    if not check_dependencies():
        sys.exit(1)
    
    # 运行主程序
    main()