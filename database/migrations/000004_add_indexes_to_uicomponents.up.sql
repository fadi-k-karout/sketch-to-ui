-- Essential indexes for UIComponents store performance

-- 1. Primary composite index for user queries (most important)
-- Covers: WHERE user_id = ? AND archived_at IS NULL ORDER BY created_at DESC
CREATE INDEX idx_uicomponents_user_active_created ON uicomponents(user_id, archived_at, created_at DESC);

-- 2. Composite index for ID lookups with archive filter
-- Covers: WHERE id = ? AND archived_at IS NULL
CREATE INDEX idx_uicomponents_id_archived ON uicomponents(id, archived_at);

-- 3. Index for archive status queries
-- Covers: WHERE archived_at IS NULL ORDER BY created_at DESC
CREATE INDEX idx_uicomponents_archived_created ON uicomponents(archived_at, created_at DESC);

-- 4. Foreign key index (if not automatically created)
-- Covers: Foreign key constraint performance
CREATE INDEX idx_uicomponents_user_id ON uicomponents(user_id);

