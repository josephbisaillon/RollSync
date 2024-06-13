
CREATE TABLE roll_state_data (
                                 id               INT AUTO_INCREMENT PRIMARY KEY,
                                 device_id        VARCHAR(255)     NOT NULL,
                                 roll_state       TINYINT UNSIGNED NOT NULL,
                                 current_face     TINYINT UNSIGNED NOT NULL,
                                 timestamp        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                 battery_level    TINYINT UNSIGNED,
                                 is_charging      BOOLEAN,
                                 led_count        TINYINT UNSIGNED,
                                 design_and_color TINYINT UNSIGNED,
                                 additional_info VARCHAR(255)
);

CREATE INDEX idx_device_timestamp ON roll_state_data(device_id, timestamp);
