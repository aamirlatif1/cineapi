-- Seed a pre-activated demo user for local development and testing.
-- Email: demo@cineapi.local  Password: pa55word
INSERT INTO users (name, email, password_hash, activated)
VALUES (
    'Demo User',
    'demo@cineapi.local',
    '$2a$12$MPvYyuelOZp/kevcLBO9He2EBLacUO2ETtze0B5U.FtDoovJpS/IO',
    true
)
ON CONFLICT (email) DO NOTHING;
