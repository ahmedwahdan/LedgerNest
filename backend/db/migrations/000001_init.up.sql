-- Enable PostgreSQL extensions used throughout the schema.

-- uuid_generate_v4() for primary keys
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pg_trgm: trigram similarity for fuzzy merchant name matching
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- citext: case-insensitive text type (used for email columns)
CREATE EXTENSION IF NOT EXISTS "citext";
