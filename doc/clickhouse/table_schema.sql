-------------------------------------------------------------- slow_query_log -------------------------------------------------------------------------
CREATE TABLE szinfra_clouddba_slow_query.slow_query_log_local ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `finger_id` FixedString(64),
    `finger_sql` String DEFAULT '',
    `slow_sql` String DEFAULT '',
    `hint` String DEFAULT '',
    `cluster_uuid` String DEFAULT '',
    `instance_host` String DEFAULT '',
    `instance_port` UInt32 DEFAULT 0,
    `database_name` String DEFAULT '',
    `environment` String DEFAULT '',
    `query_time` Float64 DEFAULT 0.,
    `lock_time` Float64 DEFAULT 0.,
    `examine_rows` UInt64 DEFAULT 0,
    `num_rows` UInt64 DEFAULT 0,
    `affect_rows` UInt64 DEFAULT 0,
    `bytes_sent` UInt64 DEFAULT 0,
    `client_host` String DEFAULT '',
    `connection_id` UInt64 DEFAULT 0,
    `default_user` String DEFAULT '',
    `client_user` String DEFAULT '',
    `log_time` DateTime DEFAULT now(),
    `create_time` DateTime DEFAULT now(),
    `killed` UInt8 DEFAULT 0,
    `lastErrno` Int32 DEFAULT 0
)ENGINE = ReplicatedMergeTree('/clickhouse/tables/{layer}-{shard}/szinfra_clouddba_slow_query.slow_query_log_local', '{replica}')
PARTITION BY (toYYYYMMDD(log_time), environment)
PRIMARY KEY (database_name, environment, log_time, instance_host, finger_id, query_time, lock_time)
ORDER BY (database_name, environment, log_time, instance_host, finger_id, query_time, lock_time)
TTL log_time + toIntervalDay(7)
SETTINGS I_insist_create_table_without_one_partition_key_even_with_poor_performance = 1, index_granularity = 8192;



CREATE TABLE szinfra_clouddba_slow_query.slow_query_log_all_rand ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `finger_id` FixedString(64),
    `finger_sql` String DEFAULT '',
    `slow_sql` String DEFAULT '',
    `hint` String DEFAULT '',
    `cluster_uuid` String DEFAULT '',
    `instance_host` String DEFAULT '',
    `instance_port` UInt32 DEFAULT 0,
    `database_name` String DEFAULT '',
    `environment` String DEFAULT '',
    `query_time` Float64 DEFAULT 0.,
    `lock_time` Float64 DEFAULT 0.,
    `examine_rows` UInt64 DEFAULT 0,
    `num_rows` UInt64 DEFAULT 0,
    `affect_rows` UInt64 DEFAULT 0,
    `bytes_sent` UInt64 DEFAULT 0,
    `client_host` String DEFAULT '',
    `connection_id` UInt64 DEFAULT 0,
    `default_user` String DEFAULT '',
    `client_user` String DEFAULT '',
    `log_time` DateTime DEFAULT now(),
    `create_time` DateTime DEFAULT now(),
    `killed` UInt8 DEFAULT 0,
    `lastErrno` Int32 DEFAULT 0
)ENGINE = Distributed('cluster_szinfra_szinfra_clouddba_online', 'szinfra_clouddba_slow_query', 'slow_query_log_local', rand64());

------------------------------------------------------- alert_operation_log ------------------------------------------------------------------------------

CREATE TABLE szinfra_clouddba_slow_query.alert_operation_log_local ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `operator` String DEFAULT '',
    `action_id` String DEFAULT '',
    `action_type` String DEFAULT '',
    `action_name` String DEFAULT '',
    `environment` String DEFAULT '',
    `create_time` DateTime DEFAULT now(),
    `old_setting` String DEFAULT '',
    `new_setting` String DEFAULT ''
)ENGINE = ReplicatedMergeTree('/clickhouse/tables/{layer}-{shard}/szinfra_clouddba_slow_query.alert_operation_log_local', '{replica}')
PARTITION BY (toYYYYMMDD(create_time), environment)
PRIMARY KEY (create_time, operator, environment)
ORDER BY (create_time, operator, environment)
TTL create_time + toIntervalDay(30)
SETTINGS I_insist_create_table_without_one_partition_key_even_with_poor_performance = 1, index_granularity = 8192;



CREATE TABLE szinfra_clouddba_slow_query.alert_operation_log_all_rand ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `operator` String DEFAULT '',
    `action_id` String DEFAULT '',
    `action_type` String DEFAULT '',
    `action_name` String DEFAULT '',
    `environment` String DEFAULT '',
    `create_time` DateTime DEFAULT now(),
    `old_setting` String DEFAULT '',
    `new_setting` String DEFAULT ''
)ENGINE = Distributed('cluster_szinfra_szinfra_clouddba_online', 'szinfra_clouddba_slow_query', 'alert_operation_log_local', rand64());

------------------------------------------------------- alert_message_mute ------------------------------------------------------------------------------

CREATE TABLE szinfra_clouddba_slow_query.alert_message_mute_local ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `monitor_mute_id` String DEFAULT '',
    `environment` String DEFAULT '',
    `mute_title` String DEFAULT '',
    `rule_uuid` String DEFAULT '',
    `monitor_alert_id` String DEFAULT '',
    `status` String DEFAULT '',
    `mute_filter` String DEFAULT '',
    `start_time` UInt64 DEFAULT 0,
    `end_time` UInt64 DEFAULT 0,
    `creator` String DEFAULT '',
    `create_time` DateTime DEFAULT now(),
    `update_time` DateTime DEFAULT now()
)ENGINE = ReplicatedMergeTree('/clickhouse/tables/{layer}-{shard}/szinfra_clouddba_slow_query.alert_message_mute_local', '{replica}')
PARTITION BY (toYYYYMMDD(create_time), environment)
PRIMARY KEY (create_time, monitor_mute_id, monitor_alert_id, start_time, end_time)
ORDER BY (create_time, monitor_mute_id, monitor_alert_id, start_time, end_time)
TTL create_time + toIntervalDay(30)
SETTINGS I_insist_create_table_without_one_partition_key_even_with_poor_performance = 1, index_granularity = 8192;


CREATE TABLE szinfra_clouddba_slow_query.alert_message_mute_all_rand ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `monitor_mute_id` String DEFAULT '',
    `environment` String DEFAULT '',
    `mute_title` String DEFAULT '',
    `rule_uuid` String DEFAULT '',
    `monitor_alert_id` String DEFAULT '',
    `status` String DEFAULT '',
    `mute_filter` String DEFAULT '',
    `start_time` UInt64 DEFAULT 0,
    `end_time` UInt64 DEFAULT 0,
    `creator` String DEFAULT '',
    `create_time` DateTime DEFAULT now(),
    `update_time` DateTime DEFAULT now()
)ENGINE = Distributed('cluster_szinfra_szinfra_clouddba_online', 'szinfra_clouddba_slow_query', 'alert_message_mute_local', rand64());


------------------------------------------------------- alert_message_log ------------------------------------------------------------------------------


CREATE TABLE szinfra_clouddba_slow_query.alert_message_log_local ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `monitor_alert_id` String DEFAULT '',
    `monitor_rule_id` String DEFAULT '',
    `alert_rule_uuid` String DEFAULT '',
    `channel_uuid` String DEFAULT '',
    `alert_rule_name` String DEFAULT '',
    `alert_strategy` String DEFAULT '',
    `alert_status` String DEFAULT '',
    `cmdb` String DEFAULT '',
    `database_name` String DEFAULT '',
    `environment` String DEFAULT '',
    `severity` String DEFAULT '',
    `message` String DEFAULT '',
    `ack_by` String DEFAULT '',
    `label_info` String DEFAULT '',
    `template_name` String DEFAULT '',
    `alert_count` UInt64 DEFAULT 0,
    `start_time` UInt64 DEFAULT 0,
    `resolve_time` UInt64 DEFAULT 0,
    `last_alert_time` DateTime DEFAULT now(),
    `create_time` DateTime DEFAULT now(),
    `update_time` DateTime DEFAULT now()
)ENGINE = ReplicatedMergeTree('/clickhouse/tables/{layer}-{shard}/szinfra_clouddba_slow_query.alert_message_log_local', '{replica}')
PARTITION BY (toYYYYMMDD(create_time), environment)
PRIMARY KEY (create_time, cmdb, database_name, environment)
ORDER BY (create_time, cmdb, database_name, environment)
TTL create_time + toIntervalDay(30)
SETTINGS I_insist_create_table_without_one_partition_key_even_with_poor_performance = 1, index_granularity = 8192;



CREATE TABLE szinfra_clouddba_slow_query.alert_message_log_all_rand ON CLUSTER cluster_szinfra_szinfra_clouddba_online
(
    `monitor_alert_id` String DEFAULT '',
    `monitor_rule_id` String DEFAULT '',
    `alert_rule_uuid` String DEFAULT '',
    `channel_uuid` String DEFAULT '',
    `alert_rule_name` String DEFAULT '',
    `alert_strategy` String DEFAULT '',
    `alert_status` String DEFAULT '',
    `cmdb` String DEFAULT '',
    `database_name` String DEFAULT '',
    `environment` String DEFAULT '',
    `severity` String DEFAULT '',
    `message` String DEFAULT '',
    `ack_by` String DEFAULT '',
    `label_info` String DEFAULT '',
    `template_name` String DEFAULT '',
    `alert_count` UInt64 DEFAULT 0,
    `start_time` UInt64 DEFAULT 0,
    `resolve_time` UInt64 DEFAULT 0,
    `last_alert_time` DateTime DEFAULT now(),
    `create_time` DateTime DEFAULT now(),
    `update_time` DateTime DEFAULT now()
)ENGINE = Distributed('cluster_szinfra_szinfra_clouddba_online', 'szinfra_clouddba_slow_query', 'alert_message_log_local', rand64());