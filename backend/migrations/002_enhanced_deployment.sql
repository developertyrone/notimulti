-- Migration: Phase 1 → Phase 2 (002-enhanced-deployment)
-- Description: Add Phase 2 features to existing notification_logs table
-- Created: 2025-11-08
-- Author: Developer Tyrone

-- =============================================================================
-- PHASE 2 ENHANCEMENTS
-- =============================================================================

-- Add is_test column for provider testing feature (User Story 2)
-- Allows filtering out test notifications from production history
ALTER TABLE notification_logs ADD COLUMN is_test INTEGER DEFAULT 0 NOT NULL;

-- Create index on is_test for efficient filtering
CREATE INDEX IF NOT EXISTS idx_notification_logs_is_test 
ON notification_logs(is_test);

-- Create composite index for common query patterns (notification history with filters)
CREATE INDEX IF NOT EXISTS idx_notification_logs_provider_status 
ON notification_logs(provider_id, status);

CREATE INDEX IF NOT EXISTS idx_notification_logs_created_status 
ON notification_logs(created_at, status);

CREATE INDEX IF NOT EXISTS idx_notification_logs_provider_created 
ON notification_logs(provider_id, created_at DESC);

-- Create index for pagination cursor queries
CREATE INDEX IF NOT EXISTS idx_notification_logs_id_created 
ON notification_logs(id, created_at DESC);

-- =============================================================================
-- DATA MIGRATION (if upgrading from Phase 1)
-- =============================================================================

-- Mark all existing notifications as non-test (production)
-- This is safe because Phase 1 had no test feature
UPDATE notification_logs SET is_test = 0 WHERE is_test IS NULL;

-- =============================================================================
-- VERIFICATION QUERIES
-- =============================================================================

-- Verify indexes were created
-- SELECT name FROM sqlite_master WHERE type='index' AND tbl_name='notification_logs';

-- Verify is_test column exists and has correct default
-- SELECT sql FROM sqlite_master WHERE type='table' AND name='notification_logs';

-- Check for any NULL is_test values (should be none)
-- SELECT COUNT(*) FROM notification_logs WHERE is_test IS NULL;

-- =============================================================================
-- ROLLBACK (if needed)
-- =============================================================================

-- WARNING: Rollback will drop Phase 2 features
-- Uncomment and run these statements to rollback:

-- DROP INDEX IF EXISTS idx_notification_logs_is_test;
-- DROP INDEX IF EXISTS idx_notification_logs_provider_status;
-- DROP INDEX IF EXISTS idx_notification_logs_created_status;
-- DROP INDEX IF EXISTS idx_notification_logs_provider_created;
-- DROP INDEX IF EXISTS idx_notification_logs_id_created;

-- Note: SQLite does not support DROP COLUMN, so to fully rollback:
-- 1. Create new table without is_test column
-- 2. Copy data (excluding is_test)
-- 3. Drop old table
-- 4. Rename new table
--
-- CREATE TABLE notification_logs_new AS 
--   SELECT id, provider_id, provider_type, recipient, message, subject, metadata,
--          priority, status, error_message, attempts, created_at, delivered_at
--   FROM notification_logs;
-- DROP TABLE notification_logs;
-- ALTER TABLE notification_logs_new RENAME TO notification_logs;
-- -- Then recreate original indexes

-- =============================================================================
-- MIGRATION COMPLETE
-- =============================================================================

-- Phase 2 migration complete
-- New features available:
--   - is_test column for filtering test notifications
--   - Optimized indexes for history queries
--   - Better query performance with composite indexes
