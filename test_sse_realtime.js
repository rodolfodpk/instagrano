#!/usr/bin/env node

const fetch = require('node-fetch');

// Test configuration
const BASE_URL = 'http://localhost:8080';
const TEST_USER1 = {
    username: 'sse_test_user1',
    email: 'sse_test_user1@example.com',
    password: 'password123'
};
const TEST_USER2 = {
    username: 'sse_test_user2', 
    email: 'sse_test_user2@example.com',
    password: 'password123'
};

let user1Token = null;
let user2Token = null;
let user1PostId = null;

async function registerUser(user) {
    console.log(`Registering user: ${user.username}`);
    const response = await fetch(`${BASE_URL}/api/auth/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(user)
    });
    
    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Registration failed: ${error}`);
    }
    
    const data = await response.json();
    console.log(`✓ User ${user.username} registered with ID: ${data.user.id}`);
    return data.user;
}

async function loginUser(user) {
    console.log(`Logging in user: ${user.username}`);
    const response = await fetch(`${BASE_URL}/api/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            email: user.email,
            password: user.password
        })
    });
    
    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Login failed: ${error}`);
    }
    
    const data = await response.json();
    console.log(`✓ User ${user.username} logged in`);
    return data.token;
}

async function createPost(token, user) {
    console.log(`Creating post for user: ${user.username}`);
    const response = await fetch(`${BASE_URL}/api/posts`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({
            title: `Test post by ${user.username}`,
            caption: `This is a test post created by ${user.username} for SSE testing`,
            media_type: 'image',
            media_url: 'http://localhost:8080/test/image'
        })
    });
    
    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Post creation failed: ${error}`);
    }
    
    const data = await response.json();
    console.log(`✓ Post created with ID: ${data.id}`);
    return data.id;
}

async function likePost(token, postId, user) {
    console.log(`User ${user.username} liking post ${postId}`);
    const response = await fetch(`${BASE_URL}/api/posts/${postId}/like`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Like failed: ${error}`);
    }
    
    const data = await response.json();
    console.log(`✓ Post ${postId} liked by ${user.username}, new count: ${data.likes_count}`);
    return data;
}

async function getFeed(token, user) {
    console.log(`Getting feed for user: ${user.username}`);
    const response = await fetch(`${BASE_URL}/api/feed`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    });
    
    if (!response.ok) {
        const error = await response.text();
        throw new Error(`Feed fetch failed: ${error}`);
    }
    
    const data = await response.json();
    console.log(`✓ Feed retrieved for ${user.username}, ${data.posts.length} posts`);
    return data;
}

// SSE Test using EventSource simulation
class SSETestClient {
    constructor(token, user) {
        this.token = token;
        this.user = user;
        this.events = [];
        this.eventSource = null;
    }
    
    connect() {
        return new Promise((resolve, reject) => {
            console.log(`Connecting SSE for user: ${this.user.username}`);
            
            const url = `${BASE_URL}/api/events/stream?token=${encodeURIComponent(this.token)}`;
            this.eventSource = new (require('eventsource'))(url);
            
            this.eventSource.onopen = () => {
                console.log(`✓ SSE connected for ${this.user.username}`);
                resolve();
            };
            
            this.eventSource.onmessage = (event) => {
                console.log(`SSE message received by ${this.user.username}:`, event.data);
                this.events.push({ type: 'message', data: event.data });
            };
            
            this.eventSource.addEventListener('post_liked', (event) => {
                const data = JSON.parse(event.data);
                console.log(`✓ Post liked event received by ${this.user.username}:`, data);
                this.events.push({ type: 'post_liked', data });
            });
            
            this.eventSource.addEventListener('post_commented', (event) => {
                const data = JSON.parse(event.data);
                console.log(`✓ Post commented event received by ${this.user.username}:`, data);
                this.events.push({ type: 'post_commented', data });
            });
            
            this.eventSource.addEventListener('new_post', (event) => {
                const data = JSON.parse(event.data);
                console.log(`✓ New post event received by ${this.user.username}:`, data);
                this.events.push({ type: 'new_post', data });
            });
            
            this.eventSource.onerror = (error) => {
                console.error(`SSE error for ${this.user.username}:`, error);
                reject(error);
            };
            
            // Timeout after 10 seconds
            setTimeout(() => {
                if (!this.eventSource || this.eventSource.readyState !== EventSource.OPEN) {
                    reject(new Error('SSE connection timeout'));
                }
            }, 10000);
        });
    }
    
    disconnect() {
        if (this.eventSource) {
            this.eventSource.close();
            console.log(`SSE disconnected for ${this.user.username}`);
        }
    }
    
    getEvents(type = null) {
        if (type) {
            return this.events.filter(event => event.type === type);
        }
        return this.events;
    }
}

async function runSSETest() {
    console.log('🚀 Starting SSE Real-time Update Test\n');
    
    try {
        // Step 1: Register and login both users
        console.log('=== Step 1: User Setup ===');
        await registerUser(TEST_USER1);
        await registerUser(TEST_USER2);
        
        user1Token = await loginUser(TEST_USER1);
        user2Token = await loginUser(TEST_USER2);
        
        // Step 2: Create SSE connections for both users
        console.log('\n=== Step 2: SSE Connection Setup ===');
        const user1SSE = new SSETestClient(user1Token, TEST_USER1);
        const user2SSE = new SSETestClient(user2Token, TEST_USER2);
        
        await Promise.all([
            user1SSE.connect(),
            user2SSE.connect()
        ]);
        
        // Wait a moment for connections to stabilize
        await new Promise(resolve => setTimeout(resolve, 1000));
        
        // Step 3: User 1 creates a post
        console.log('\n=== Step 3: Post Creation ===');
        user1PostId = await createPost(user1Token, TEST_USER1);
        
        // Wait for new_post event to propagate
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Step 4: User 2 likes the post
        console.log('\n=== Step 4: Like Action ===');
        const likeResult = await likePost(user2Token, user1PostId, TEST_USER2);
        
        // Wait for post_liked event to propagate
        await new Promise(resolve => setTimeout(resolve, 2000));
        
        // Step 5: Verify SSE events were received
        console.log('\n=== Step 5: Event Verification ===');
        
        const user1Events = user1SSE.getEvents();
        const user2Events = user2SSE.getEvents();
        
        console.log(`User 1 received ${user1Events.length} events:`, user1Events.map(e => e.type));
        console.log(`User 2 received ${user2Events.length} events:`, user2Events.map(e => e.type));
        
        // Check if user 1 received the post_liked event
        const user1LikeEvents = user1SSE.getEvents('post_liked');
        const user2LikeEvents = user2SSE.getEvents('post_liked');
        
        if (user1LikeEvents.length > 0) {
            console.log('✅ SUCCESS: User 1 received post_liked event!');
            console.log('   Event data:', user1LikeEvents[0].data);
        } else {
            console.log('❌ FAILURE: User 1 did not receive post_liked event');
        }
        
        if (user2LikeEvents.length > 0) {
            console.log('✅ SUCCESS: User 2 received post_liked event!');
            console.log('   Event data:', user2LikeEvents[0].data);
        } else {
            console.log('❌ FAILURE: User 2 did not receive post_liked event');
        }
        
        // Step 6: Verify feed updates
        console.log('\n=== Step 6: Feed Verification ===');
        const user1Feed = await getFeed(user1Token, TEST_USER1);
        const user2Feed = await getFeed(user2Token, TEST_USER2);
        
        const user1Post = user1Feed.posts.find(p => p.id === user1PostId);
        const user2Post = user2Feed.posts.find(p => p.id === user1PostId);
        
        if (user1Post && user1Post.likes_count === likeResult.likes_count) {
            console.log('✅ SUCCESS: User 1 feed shows correct like count!');
        } else {
            console.log('❌ FAILURE: User 1 feed like count mismatch');
        }
        
        if (user2Post && user2Post.likes_count === likeResult.likes_count) {
            console.log('✅ SUCCESS: User 2 feed shows correct like count!');
        } else {
            console.log('❌ FAILURE: User 2 feed like count mismatch');
        }
        
        // Cleanup
        user1SSE.disconnect();
        user2SSE.disconnect();
        
        console.log('\n🎉 SSE Real-time Update Test Completed!');
        
    } catch (error) {
        console.error('❌ Test failed:', error.message);
        process.exit(1);
    }
}

// Run the test
runSSETest();
