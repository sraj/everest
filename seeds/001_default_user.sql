-- Default user for development
INSERT INTO users (id, email, name)
VALUES ('00000000-0000-0000-0000-000000000001', 'default@example.com', 'Default User')
ON CONFLICT (id) DO NOTHING;
