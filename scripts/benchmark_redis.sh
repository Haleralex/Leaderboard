#!/bin/bash

set -e

echo "🚀 Starting Redis Performance Benchmark"
echo "========================================"
echo ""

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "📋 Checking infrastructure..."
if ! docker ps | grep -q postgres; then
    echo -e "${RED}❌ PostgreSQL container not running${NC}"
    echo "Run: docker-compose up -d"
    exit 1
fi

REDIS_RUNNING=false
if docker ps | grep -q redis; then
    REDIS_RUNNING=true
    echo -e "${GREEN}✅ Redis is running${NC}"
else
    echo -e "${YELLOW}⚠️  Redis is not running${NC}"
fi

echo ""
echo "================================"
echo "📊 Running benchmarks WITH Redis"
echo "================================"
echo ""

if [ "$REDIS_RUNNING" = false ]; then
    echo "Starting Redis..."
    docker-compose up -d redis
    sleep 2
fi

go test -bench=. -benchmem -benchtime=5s -timeout=30m \
    ./internal/service \
    > bench_with_redis.txt 2>&1

echo -e "${GREEN}✅ Benchmark with Redis completed${NC}"
echo ""

echo "=================================="
echo "📊 Running benchmarks WITHOUT Redis"
echo "=================================="
echo ""

echo "Stopping Redis..."
docker-compose stop redis
sleep 2

go test -bench=BenchmarkGetLeaderboardNoRedis -benchmem -benchtime=5s -timeout=30m \
    ./internal/service \
    > bench_without_redis.txt 2>&1

echo -e "${GREEN}✅ Benchmark without Redis completed${NC}"
echo ""

echo "Restarting Redis..."
docker-compose up -d redis
sleep 2

echo ""
echo "======================================="
echo "📈 Benchmark Results"
echo "======================================="
echo ""

echo -e "${YELLOW}=== WITH REDIS ===${NC}"
grep "^Benchmark" bench_with_redis.txt | head -20
echo ""

echo -e "${YELLOW}=== WITHOUT REDIS ===${NC}"
grep "^Benchmark" bench_without_redis.txt
echo ""

if command -v benchstat &> /dev/null; then
    echo ""
    echo "======================================="
    echo "📊 Detailed Comparison (benchstat)"
    echo "======================================="
    echo ""
    benchstat bench_without_redis.txt bench_with_redis.txt
else
    echo ""
    echo -e "${YELLOW}💡 Install benchstat for detailed comparison:${NC}"
    echo "   go install golang.org/x/perf/cmd/benchstat@latest"
fi

echo ""
echo "======================================="
echo "🔍 Quick Analysis"
echo "======================================="
echo ""

WITH_REDIS=$(grep "BenchmarkGetLeaderboardWithRedis" bench_with_redis.txt | awk '{print $3}')
WITHOUT_REDIS=$(grep "BenchmarkGetLeaderboardNoRedis" bench_without_redis.txt | awk '{print $3}')

if [ -n "$WITH_REDIS" ] && [ -n "$WITHOUT_REDIS" ]; then
    echo "GetLeaderboard performance:"
    echo "  • WITH Redis:    $WITH_REDIS ns/op"
    echo "  • WITHOUT Redis: $WITHOUT_REDIS ns/op"
    
    SPEEDUP=$(echo "scale=2; $WITHOUT_REDIS / $WITH_REDIS" | bc 2>/dev/null || echo "N/A")
    if [ "$SPEEDUP" != "N/A" ]; then
        echo -e "  • ${GREEN}Speedup: ${SPEEDUP}x faster with Redis${NC}"
    fi
else
    echo "Could not extract benchmark metrics"
fi

echo ""
echo "======================================="
echo "📁 Full results saved to:"
echo "======================================="
echo "  • bench_with_redis.txt"
echo "  • bench_without_redis.txt"
echo ""

echo -e "${GREEN}✅ Benchmark complete!${NC}"
