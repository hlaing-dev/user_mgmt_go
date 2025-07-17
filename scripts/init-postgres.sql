-- PostgreSQL initialization script for User Management System

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create users table (GORM will handle this, but we can prepare extensions)
-- This file is mainly for setting up extensions and default data

-- Create database if not exists (handled by Docker environment)

-- Enable necessary extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";  -- For additional crypto functions
CREATE EXTENSION IF NOT EXISTS "citext";    -- For case-insensitive text

-- Set timezone
SET timezone = 'UTC';

-- Create a function to update updated_at timestamps
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- This trigger will be applied by GORM migrations
-- But having the function ready helps

-- Create indexes that might be useful (GORM will create these too)
-- This is mainly for documentation and manual setup if needed

-- Performance settings for development
-- These are safe for development containers
ALTER SYSTEM SET shared_preload_libraries = 'pg_stat_statements';
ALTER SYSTEM SET max_connections = 100;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '1GB';
ALTER SYSTEM SET work_mem = '4MB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '16MB';
ALTER SYSTEM SET default_statistics_target = 100;

-- Reload configuration
SELECT pg_reload_conf();

-- Create development user with necessary permissions
-- (This is already handled by Docker environment variables) 