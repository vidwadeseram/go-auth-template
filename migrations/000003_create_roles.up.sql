CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles (id, name, description)
VALUES
    ('11111111-1111-1111-1111-111111111111', 'super_admin', 'Super administrator'),
    ('22222222-2222-2222-2222-222222222222', 'admin', 'Administrator'),
    ('33333333-3333-3333-3333-333333333333', 'user', 'Standard user')
ON CONFLICT (name) DO NOTHING;
