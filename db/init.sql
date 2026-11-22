CREATE TABLE IF NOT EXISTS measurements (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    vin VARCHAR(32) NOT NULL,
    component VARCHAR(64) NOT NULL,
    measured_value DOUBLE NOT NULL,
    is_valid BOOLEAN NOT NULL,
    checked_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_vin_checked_at (vin, checked_at)
);
