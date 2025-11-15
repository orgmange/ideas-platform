CREATE TABLE otps (
    id BIGSERIAL PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    verified BOOLEAN DEFAULT false,
    attempts_left INTEGER DEFAULT 3,
    created_at TIMESTAMP DEFAULT NOW()
);
