CREATE TABLE authboss_user (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255),
    confirmed_at TIMESTAMP,
    locked_at TIMESTAMP,
    recover_token VARCHAR(255),
    recover_expiry TIMESTAMP,
    last_attempt_at TIMESTAMP,
    failed_attempts INTEGER,
    totp_recovery_codes TEXT[]
);