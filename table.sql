DROP TABLE IF EXISTS `usages`;
CREATE TABLE IF NOT EXISTS `usages` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT NOT NULL,
    `host_id` VARCHAR(128) NOT NULL COMMENT 'Host ID',
    `os` VARCHAR(128) NOT NULL COMMENT 'OS',
    `hostname` VARCHAR(128) NOT NULL COMMENT 'Host Name',
    `version` VARCHAR(128) NOT NULL COMMENT 'Version',
    `uptime` INT(10) NOT NULL DEFAULT 0 COMMENT 'Uptime',
    `download` INT(10) NOT NULL DEFAULT 0 COMMENT 'Download',
    `upload` INT(10) NOT NULL DEFAULT 0 COMMENT 'Upload',
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Created_at',
    KEY `idx_created_at`(`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
