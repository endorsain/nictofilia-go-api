INSERT INTO roles (name, description)
VALUES
    ('admin', 'Administrator role'),
    ('moderator', 'Moderator role'),
    ('regular', 'Regular user role')
ON CONFLICT (name) DO NOTHING;
