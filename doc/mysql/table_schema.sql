CREATE TABLE `internal_user_tab` (
  `id` bigint(20) NOT NULL AUTO_INCREMENT COMMENT 'id',
  `user_name` varchar(254) DEFAULT '' COMMENT 'user name',
  `env` varchar(254) DEFAULT '' COMMENT 'env',
  `password_hash` varbinary(1024) DEFAULT '' COMMENT 'password hash',
  `update_time` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update time',
  `create_time` datetime DEFAULT CURRENT_TIMESTAMP COMMENT 'create timestamp',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_user` (`user_name`)
) ENGINE=InnoDB CHARSET=utf8mb4;



CREATE TABLE `slow_query_db_daily_rank` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
  `serial_no` int(10) NOT NULL DEFAULT -1 COMMENT 'rank serial no',
  `cluster_uuid` varchar(30) NOT NULL DEFAULT '' COMMENT 'cluster_uuid',
  `db_name` varchar(256) NOT NULL DEFAULT '' COMMENT 'database name',
  `db_env` varchar(256) NOT NULL DEFAULT '' COMMENT 'database env',
  `rank_order` varchar(128) NOT NULL DEFAULT '' COMMENT 'order by of rank',
  `rank_score` double NOT NULL DEFAULT 0 COMMENT 'score of order',
  `sql_count` int(10) NOT NULL DEFAULT -1 COMMENT 'slow sql count',
  `week_on_week` varchar(60) NOT NULL DEFAULT '' COMMENT 'rank week-on-week',
  `rank_day` Date NOT NULL COMMENT 'rank day',
  `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create timestamp',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uniq_db_rank_day` (`db_name`,`rank_day`),
  KEY `idx_rank_day` (`rank_day`)
) ENGINE=InnoDB CHARSET=utf8mb4;



CREATE TABLE `slow_query_finger_daily_rank` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `serial_no` int(10) NOT NULL DEFAULT -1 COMMENT 'rank serial no',
    `finger_id` varchar(128) NOT NULL DEFAULT '' COMMENT 'finger_id',
    `finger_sql` text NOT NULL COMMENT 'finger_sql',
    `cluster_uuid` varchar(30) NOT NULL DEFAULT '' COMMENT 'cluster_uuid',
    `db_name` varchar(256) NOT NULL DEFAULT '' COMMENT 'database name',
    `db_env` varchar(256) NOT NULL DEFAULT '' COMMENT 'database env',
    `rank_order` varchar(128) NOT NULL DEFAULT '' COMMENT 'order by of rank',
    `rank_score` double NOT NULL DEFAULT 0 COMMENT 'score of order',
    `sql_count` int(10) NOT NULL DEFAULT -1 COMMENT 'slow sql count',
    `week_on_week` varchar(60) NOT NULL DEFAULT '' COMMENT 'rank week-on-week',
    `rank_day` Date NOT NULL COMMENT 'rank day',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create timestamp',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_db_rank_day` (`finger_id`,`db_name`,`rank_day`),
    KEY `idx_rank_day` (`rank_day`)
) ENGINE=InnoDB CHARSET=utf8mb4;


CREATE TABLE `slow_query_daily_rank_email_log` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `task_uuid` varchar(32) NOT NULL DEFAULT '' COMMENT 'task uuid hash code',
    `task_name` varchar(30) NOT NULL DEFAULT '' COMMENT 'daily report name',
    `product_line` text NOT NULL COMMENT 'db productLine',
    `owners`  text NOT NULL COMMENT 'db pic/applicant',
    `leaders` text NOT NULL COMMENT 'productLine leaders',
    `db_env` varchar(20) NOT NULL DEFAULT '' COMMENT 'db env',
    `report_day` Date NOT NULL COMMENT 'report day',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create timestamp',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_task_uuid` (`task_uuid`,`report_day`),
    KEY `idx_report_day` (`report_day`)
) ENGINE=InnoDB CHARSET=utf8mb4;


CREATE TABLE `alert_rule_tab` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `alert_rule_uuid` varchar(255) NOT NULL COMMENT 'uuid by self',
    `channel_uuid` varchar(255) NOT NULL COMMENT 'channel_uuid',
    `strategy_id` varchar(255) NOT NULL DEFAULT '',
    `cmdb` varchar(255) NOT NULL DEFAULT '' COMMENT 'cmdb service',
    `db_env` varchar(20) NOT NULL DEFAULT '' COMMENT 'db env',
    `rule_trigger` varchar(255) NOT NULL DEFAULT '',
    `monitor_rule_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'monitor_rule_id from monitor',
    `alert_display_name` varchar(255) NOT NULL DEFAULT '' COMMENT 'alert_rule_name, self',
    `template_name` varchar(255) NOT NULL DEFAULT '' COMMENT 'template_name',
    `severity` varchar(255) NOT NULL DEFAULT '' COMMENT 'warn,error,critical',
    `prom_ql` text NOT NULL COMMENT 'promQL',
    `expression` varchar(16) NOT NULL DEFAULT '' COMMENT 'expression',
    `expression_value` varchar(16) NOT NULL DEFAULT '0' COMMENT 'threshold by expression',
    `for_range` varchar(255) NOT NULL DEFAULT '' COMMENT '',
    `evaluation_interval` varchar(255) NOT NULL DEFAULT '' COMMENT 'interval for alert',
    `alarm_msg` text NOT NULL COMMENT 'notice message',
    `resolve_msg` text NOT NULL COMMENT 'resolved message',
    `rule_status` varchar(20)  NOT NULL DEFAULT '' COMMENT 'disable,enable',
    `soft_status` varchar(255) NOT NULL,
    `dbs_json` text NOT NULL COMMENT 'db list',
    `creator` varchar(64) NOT NULL DEFAULT '' COMMENT 'creator',
    `modifier` varchar(64) NOT NULL DEFAULT '' COMMENT 'modifier latest',
    `channel_type` varchar(200) NOT NULL COMMENT 'channel:seaTalk/email/phone/MatterMost',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create_time',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update_time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_idx_alert_rule_uuid` (`alert_rule_uuid`),
    KEY `idx_channel_uuid` (`channel_uuid`),
    KEY `idx_cmdb` (`cmdb`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_update_time` (`update_time`),
    KEY `idx_monitor_rule_id` (`monitor_rule_id`),
    KEY `idx_alert_display_name` (`alert_display_name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;


CREATE TABLE `alert_channel_tab` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `channel_uuid` varchar(255) NOT NULL COMMENT 'uuid ,self',
    `channel_name` varchar(255) NOT NULL DEFAULT '' COMMENT 'channel_name',
    `dod_id` int(11) NOT NULL DEFAULT '0' COMMENT 'dod id',
    `users_json` text NOT NULL COMMENT 'use list',
    `channel_interval` varchar(64) NOT NULL,
    `meta_json` mediumtext NOT NULL COMMENT 'meta message',
    `channel_status` varchar(20) NOT NULL DEFAULT '' COMMENT 'disable,enable',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create_time',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update_time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_channel_uuid` (`channel_uuid`),
    KEY `idx_channel_name` (`channel_name`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `alert_strategy_tab` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,
    `strategy_name` varchar(255) NOT NULL DEFAULT '' COMMENT 'strategy',
    `strategy_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'strategy_id',
    `meta_json` mediumtext NOT NULL COMMENT 'meta message',
    `strategy_status` varchar(20) NOT NULL DEFAULT '' COMMENT 'disable,enable',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create_time',
    `update_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'update_time',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_strategy_id` (`strategy_id`),
    KEY `idx_strategy_name` (`strategy_name`),
    KEY `idx_create_time` (`create_time`),
    KEY `idx_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE `slow_query_daily_new_finger_email_log` (
    `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT COMMENT 'id',
    `task_uuid` varchar(32) NOT NULL DEFAULT '' COMMENT 'task uuid hash code',
    `task_name` varchar(30) NOT NULL DEFAULT '' COMMENT 'daily report name',
    `product_line` text NOT NULL COMMENT 'db productLine',
    `owners` text NOT NULL COMMENT 'db productLine owners',
    `leaders` text NOT NULL COMMENT 'db productLine leaders',
    `db_env` varchar(20) NOT NULL DEFAULT '' COMMENT 'db env',
    `new_finger` bigint(20) NOT NULL DEFAULT 0 COMMENT 'new_finger count',
    `new_sql_query` bigint(20) NOT NULL DEFAULT 0 COMMENT 'new_sql_query count',
    `report_day` date NOT NULL COMMENT 'report day',
    `create_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create timestamp',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uniq_task_uuid` (`task_uuid`,`report_day`),
    KEY `idx_report_day` (`report_day`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

