#!/bin/bash

# 设置颜色
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 设置工作目录
PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
cd "$PROJECT_ROOT" || exit 1

echo -e "${BLUE}=== Fabric Realty 测试套件 ===${NC}"
echo -e "${BLUE}工作目录: ${PROJECT_ROOT}${NC}"

# 创建测试结果目录
RESULTS_DIR="$PROJECT_ROOT/tests/results"
mkdir -p "$RESULTS_DIR"

# 运行链码测试
run_chaincode_tests() {
    echo -e "\n${YELLOW}=== 运行链码测试 ===${NC}"
    cd "$PROJECT_ROOT/chaincode" || return 1
    
    # 安装依赖
    echo -e "${YELLOW}安装依赖...${NC}"
    go get github.com/stretchr/testify/assert
    go get github.com/stretchr/testify/mock
    
    # 运行测试
    echo -e "${YELLOW}运行测试...${NC}"
    go test -v ../tests/chaincode/chaincode_test.go > "$RESULTS_DIR/chaincode_test_results.txt" 2>&1
    
    # 检查测试结果
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}链码测试通过!${NC}"
        return 0
    else
        echo -e "${RED}链码测试失败!${NC}"
        echo -e "${YELLOW}查看详细结果: $RESULTS_DIR/chaincode_test_results.txt${NC}"
        return 1
    fi
}

# 运行后端服务测试
run_server_tests() {
    echo -e "\n${YELLOW}=== 运行后端服务测试 ===${NC}"
    cd "$PROJECT_ROOT/application/server" || return 1
    
    # 安装依赖
    echo -e "${YELLOW}安装依赖...${NC}"
    go get github.com/stretchr/testify/assert
    go get github.com/stretchr/testify/mock
    go get github.com/gin-gonic/gin
    
    # 运行 API 测试
    echo -e "${YELLOW}运行 API 测试...${NC}"
    go test -v ../../tests/server/api_test.go > "$RESULTS_DIR/server_api_test_results.txt" 2>&1
    API_TEST_RESULT=$?
    
    # 运行服务层测试
    echo -e "${YELLOW}运行服务层测试...${NC}"
    go test -v ../../tests/server/service_test.go > "$RESULTS_DIR/server_service_test_results.txt" 2>&1
    SERVICE_TEST_RESULT=$?
    
    # 检查测试结果
    if [ $API_TEST_RESULT -eq 0 ] && [ $SERVICE_TEST_RESULT -eq 0 ]; then
        echo -e "${GREEN}后端服务测试通过!${NC}"
        return 0
    else
        echo -e "${RED}后端服务测试失败!${NC}"
        echo -e "${YELLOW}查看 API 测试结果: $RESULTS_DIR/server_api_test_results.txt${NC}"
        echo -e "${YELLOW}查看服务层测试结果: $RESULTS_DIR/server_service_test_results.txt${NC}"
        return 1
    fi
}

# 运行前端测试
run_web_tests() {
    echo -e "\n${YELLOW}=== 运行前端测试 ===${NC}"
    cd "$PROJECT_ROOT/application/web" || return 1
    
    # 安装依赖
    echo -e "${YELLOW}安装依赖...${NC}"
    npm install --silent
    npm install --silent vitest @vue/test-utils
    
    # 创建 vitest.config.js
    echo -e "${YELLOW}创建测试配置...${NC}"
    cat > vitest.config.js << 'EOF'
import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
  },
})
EOF
    
    # 运行测试
    echo -e "${YELLOW}运行测试...${NC}"
    npx vitest run ../../tests/web/components.spec.js > "$RESULTS_DIR/web_test_results.txt" 2>&1
    
    # 检查测试结果
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}前端测试通过!${NC}"
        return 0
    else
        echo -e "${RED}前端测试失败!${NC}"
        echo -e "${YELLOW}查看详细结果: $RESULTS_DIR/web_test_results.txt${NC}"
        return 1
    fi
}

# 运行集成测试
run_integration_tests() {
    echo -e "\n${YELLOW}=== 运行集成测试 ===${NC}"
    cd "$PROJECT_ROOT/tests/integration" || return 1
    
    # 检查是否需要运行集成测试
    if [ "$RUN_INTEGRATION_TESTS" != "true" ]; then
        echo -e "${YELLOW}跳过集成测试。设置 RUN_INTEGRATION_TESTS=true 以运行集成测试。${NC}"
        return 0
    fi
    
    # 检查后端服务是否运行
    echo -e "${YELLOW}检查后端服务...${NC}"
    curl -s http://localhost:8888/api/car-dealer/car/test > /dev/null
    if [ $? -ne 0 ]; then
        echo -e "${RED}后端服务未运行，无法执行集成测试!${NC}"
        echo -e "${YELLOW}请先启动后端服务:${NC}"
        echo -e "${YELLOW}cd $PROJECT_ROOT/application/server && go run main.go${NC}"
        return 1
    fi
    
    # 运行测试
    echo -e "${YELLOW}运行测试...${NC}"
    RUN_INTEGRATION_TESTS=true go test -v . > "$RESULTS_DIR/integration_test_results.txt" 2>&1
    
    # 检查测试结果
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}集成测试通过!${NC}"
        return 0
    else
        echo -e "${RED}集成测试失败!${NC}"
        echo -e "${YELLOW}查看详细结果: $RESULTS_DIR/integration_test_results.txt${NC}"
        return 1
    fi
}

# 运行所有测试
run_all_tests() {
    local chaincode_result=0
    local server_result=0
    local web_result=0
    local integration_result=0
    
    run_chaincode_tests
    chaincode_result=$?
    
    run_server_tests
    server_result=$?
    
    run_web_tests
    web_result=$?
    
    run_integration_tests
    integration_result=$?
    
    # 显示测试结果摘要
    echo -e "\n${BLUE}=== 测试结果摘要 ===${NC}"
    
    if [ $chaincode_result -eq 0 ]; then
        echo -e "${GREEN}✓ 链码测试通过${NC}"
    else
        echo -e "${RED}✗ 链码测试失败${NC}"
    fi
    
    if [ $server_result -eq 0 ]; then
        echo -e "${GREEN}✓ 后端服务测试通过${NC}"
    else
        echo -e "${RED}✗ 后端服务测试失败${NC}"
    fi
    
    if [ $web_result -eq 0 ]; then
        echo -e "${GREEN}✓ 前端测试通过${NC}"
    else
        echo -e "${RED}✗ 前端测试失败${NC}"
    fi
    
    if [ "$RUN_INTEGRATION_TESTS" = "true" ]; then
        if [ $integration_result -eq 0 ]; then
            echo -e "${GREEN}✓ 集成测试通过${NC}"
        else
            echo -e "${RED}✗ 集成测试失败${NC}"
        fi
    else
        echo -e "${YELLOW}? 集成测试已跳过${NC}"
    fi
    
    # 检查是否所有测试都通过
    if [ $chaincode_result -eq 0 ] && [ $server_result -eq 0 ] && [ $web_result -eq 0 ] && ([ "$RUN_INTEGRATION_TESTS" != "true" ] || [ $integration_result -eq 0 ]); then
        echo -e "\n${GREEN}所有测试通过!${NC}"
        return 0
    else
        echo -e "\n${RED}部分测试失败!${NC}"
        return 1
    fi
}

# 显示帮助信息
show_help() {
    echo -e "${BLUE}用法:${NC}"
    echo -e "  $0 [选项]"
    echo -e ""
    echo -e "${BLUE}选项:${NC}"
    echo -e "  all          运行所有测试"
    echo -e "  chaincode    只运行链码测试"
    echo -e "  server       只运行后端服务测试"
    echo -e "  web          只运行前端测试"
    echo -e "  integration  只运行集成测试"
    echo -e "  help         显示此帮助信息"
    echo -e ""
    echo -e "${BLUE}环境变量:${NC}"
    echo -e "  RUN_INTEGRATION_TESTS=true  启用集成测试"
    echo -e ""
    echo -e "${BLUE}示例:${NC}"
    echo -e "  $0 all                      运行所有测试"
    echo -e "  RUN_INTEGRATION_TESTS=true $0 all  运行所有测试，包括集成测试"
    echo -e "  $0 chaincode                只运行链码测试"
}

# 主函数
main() {
    # 如果没有参数，显示帮助信息
    if [ $# -eq 0 ]; then
        show_help
        exit 0
    fi
    
    # 处理参数
    case "$1" in
        all)
            run_all_tests
            ;;
        chaincode)
            run_chaincode_tests
            ;;
        server)
            run_server_tests
            ;;
        web)
            run_web_tests
            ;;
        integration)
            RUN_INTEGRATION_TESTS=true run_integration_tests
            ;;
        help)
            show_help
            ;;
        *)
            echo -e "${RED}未知选项: $1${NC}"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
