-- Critical: Optimize posts table for cursor pagination
-- Current query uses WHERE (created_at < $2) OR (created_at = $2 AND id < $3) ORDER BY created_at DESC, id DESC
-- Single-column index on created_at forces a sort; composite index does single scan
CREATE INDEX IF NOT EXISTS idx_posts_created_at_id ON posts(created_at DESC, id DESC);

-- Add missing user_id indexes for user-centric queries
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_likes_user_id ON likes(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);

-- Optimize for ORDER BY created_at queries (FindByPostID methods)
CREATE INDEX IF NOT EXISTS idx_likes_post_id_created_at ON likes(post_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_comments_post_id_created_at ON comments(post_id, created_at DESC);
