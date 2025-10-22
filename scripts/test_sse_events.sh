#!/bin/bash

echo "🧪 Testing SSE Event Reception"
echo "=============================="

# Test tokens
USER1_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjExNjY4NjUsInVzZXJfaWQiOjR9.Jziv7ErxLa_TmHXcd2-4iVWEPcO3DVzaptFGp0Ptz8s"
USER2_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjExNjY4NjgsInVzZXJfaWQiOjV9.s02n87wtTfV_CnAieqNCtAfcDJoxBF0F6LMknoIAIqE"

echo "Starting SSE connection for User 1 (ID: 4)..."
echo "SSE URL: http://localhost:8080/api/events/stream?token=${USER1_TOKEN:0:20}..."

# Start SSE connection and capture events
timeout 10s curl -s -N "http://localhost:8080/api/events/stream?token=$USER1_TOKEN" > /tmp/sse_events.log &
SSE_PID=$!

echo "SSE connection started (PID: $SSE_PID)"
sleep 2

echo ""
echo "Testing like action..."
echo "User 2 (ID: 5) liking post 2..."

# Perform like action
LIKE_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $USER2_TOKEN" http://localhost:8080/api/posts/2/like)
echo "Like response: $LIKE_RESPONSE"

# Wait for events to propagate
sleep 3

# Check captured events
echo ""
echo "SSE Events Received:"
echo "==================="
if [ -f "/tmp/sse_events.log" ]; then
    cat /tmp/sse_events.log
else
    echo "No SSE events captured"
fi

# Cleanup
kill $SSE_PID 2>/dev/null
rm -f /tmp/sse_events.log

echo ""
echo "Test completed!"
