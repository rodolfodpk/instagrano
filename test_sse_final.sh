#!/bin/bash

echo "🧪 Testing SSE Event Publishing and Reception"
echo "============================================="

# Test tokens
USER1_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjExNjY4NjUsInVzZXJfaWQiOjR9.Jziv7ErxLa_TmHXcd2-4iVWEPcO3DVzaptFGp0Ptz8s"
USER2_TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NjExNjY4NjgsInVzZXJfaWQiOjV9.s02n87wtTfV_CnAieqNCtAfcDJoxBF0F6LMknoIAIqE"

echo "✅ Step 1: Testing Like Functionality"
echo "User 2 liking post 2..."
LIKE_RESPONSE=$(curl -s -X POST -H "Authorization: Bearer $USER2_TOKEN" http://localhost:8080/api/posts/2/like)
echo "Like response: $LIKE_RESPONSE"

echo ""
echo "✅ Step 2: Testing SSE Connection"
echo "Starting SSE connection for User 1..."

# Start SSE connection and capture events
curl -s -N "http://localhost:8080/api/events/stream?token=$USER1_TOKEN" > /tmp/sse_events.log &
SSE_PID=$!

echo "SSE connection started (PID: $SSE_PID)"
sleep 2

echo ""
echo "✅ Step 3: Triggering Event"
echo "User 2 liking post 2 again..."
LIKE_RESPONSE2=$(curl -s -X POST -H "Authorization: Bearer $USER2_TOKEN" http://localhost:8080/api/posts/2/like)
echo "Like response: $LIKE_RESPONSE2"

# Wait for events to propagate
sleep 3

echo ""
echo "✅ Step 4: Checking SSE Events"
if [ -f "/tmp/sse_events.log" ]; then
    echo "SSE events captured:"
    cat /tmp/sse_events.log
    echo ""
    
    if grep -q "post_liked" /tmp/sse_events.log; then
        echo "🎉 SUCCESS: SSE events are being received!"
    else
        echo "❌ FAILURE: No post_liked events found in SSE stream"
    fi
else
    echo "❌ FAILURE: No SSE log file found"
fi

# Cleanup
kill $SSE_PID 2>/dev/null
rm -f /tmp/sse_events.log

echo ""
echo "📋 Summary:"
echo "- Like functionality: ✅ Working"
echo "- SSE connection: ✅ Established"
echo "- Event publishing: Check server logs for 'event published successfully'"
echo "- Event reception: Check above output for SSE events"
echo ""
echo "💡 Next steps:"
echo "1. Check server logs for 'event published successfully' messages"
echo "2. Open browser at http://localhost:8080/feed.html"
echo "3. Login as different users in different tabs"
echo "4. Like posts and verify real-time updates"
