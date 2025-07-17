// MongoDB initialization script for User Management System

// Switch to user_logs database
db = db.getSiblingDB('user_logs');

// Create user_logs collection with validation schema
db.createCollection('user_logs', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['event', 'data', 'timestamp'],
      properties: {
        user_id: {
          bsonType: ['string', 'null'],
          description: 'User ID - UUID format or null for system events'
        },
        event: {
          bsonType: 'string',
          enum: [
            'USER_CREATED',
            'USER_UPDATED', 
            'USER_DELETED',
            'USER_LOGIN',
            'ADMIN_LOGIN',
            'ADMIN_LOGOUT',
            'LOGIN_SUCCESS',
            'LOGIN_FAILED',
            'TOKEN_REFRESH',
            'SYSTEM_ERROR',
            'VALIDATION_ERROR'
          ],
          description: 'Event type - must be one of the predefined values'
        },
        data: {
          bsonType: 'object',
          required: ['action'],
          properties: {
            action: {
              bsonType: 'string',
              description: 'Action description'
            },
            details: {
              bsonType: 'object',
              description: 'Additional event details'
            },
            old_values: {
              bsonType: ['object', 'null'],
              description: 'Previous values for update operations'
            },
            new_values: {
              bsonType: ['object', 'null'],
              description: 'New values for update operations'
            },
            error: {
              bsonType: ['string', 'null'],
              description: 'Error message for error events'
            },
            duration: {
              bsonType: ['number', 'null'],
              description: 'Operation duration in milliseconds'
            },
            status_code: {
              bsonType: ['number', 'null'],
              description: 'HTTP status code'
            }
          }
        },
        timestamp: {
          bsonType: 'date',
          description: 'Event timestamp'
        },
        ip_address: {
          bsonType: ['string', 'null'],
          description: 'Client IP address'
        },
        user_agent: {
          bsonType: ['string', 'null'],
          description: 'Client user agent'
        }
      }
    }
  }
});

// Create indexes for optimal query performance
db.user_logs.createIndex({ 'user_id': 1 }, { name: 'idx_user_id' });
db.user_logs.createIndex({ 'event': 1 }, { name: 'idx_event' });
db.user_logs.createIndex({ 'timestamp': -1 }, { name: 'idx_timestamp' });
db.user_logs.createIndex({ 'user_id': 1, 'timestamp': -1 }, { name: 'idx_user_timestamp' });
db.user_logs.createIndex({ 'ip_address': 1 }, { name: 'idx_ip_address', sparse: true });
db.user_logs.createIndex({ 'data.action': 1 }, { name: 'idx_action' });

// Create compound indexes for common query patterns
db.user_logs.createIndex({ 'event': 1, 'timestamp': -1 }, { name: 'idx_event_timestamp' });
db.user_logs.createIndex({ 'user_id': 1, 'event': 1 }, { name: 'idx_user_event' });

// Create text index for search functionality
db.user_logs.createIndex({
  'event': 'text',
  'data.action': 'text',
  'data.error': 'text'
}, {
  name: 'idx_text_search',
  weights: {
    'event': 10,
    'data.action': 5,
    'data.error': 1
  }
});

// Create TTL index for automatic log cleanup (optional - 1 year retention)
// Uncomment the following line to enable automatic log cleanup after 1 year
// db.user_logs.createIndex({ 'timestamp': 1 }, { expireAfterSeconds: 31536000, name: 'idx_ttl' });

// Insert a sample log entry to verify everything works
db.user_logs.insertOne({
  event: 'SYSTEM_ERROR',
  data: {
    action: 'DATABASE_INIT',
    details: {
      message: 'MongoDB initialized successfully',
      version: '6.0',
      environment: 'development'
    }
  },
  timestamp: new Date(),
  ip_address: 'localhost',
  user_agent: 'MongoDB Init Script'
});

print('✅ MongoDB user_logs database initialized successfully');
print('📊 Collections created: user_logs');
print('🔍 Indexes created: 8 indexes for optimal performance');
print('📝 Sample log entry inserted');

// Show database stats
print('\n📈 Database Statistics:');
printjson(db.stats());

print('\n📋 Collection Statistics:');
printjson(db.user_logs.stats()); 