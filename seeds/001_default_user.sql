-- Default user for development (Zitadel sub claim is numeric)
-- Replace the ID below with your actual Zitadel sub ID after login
INSERT INTO users (id, email, name)
VALUES ('374299731825786883', 'default@example.com', 'Default User')
ON CONFLICT (id) DO NOTHING;
