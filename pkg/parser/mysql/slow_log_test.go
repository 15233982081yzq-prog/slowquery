package mysql

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	slowPattern = map[string]string{
		"GREEDYMULTILINE": `(.|\n)*`,
		"METRICSPACE":     `([ #\n]*)`,
		"EXPLAIN":         `(# explain:.*\n|#\\s*\n)*`,
		"SLOW": "^# User@Host: %{USER:defaultUser}(\\[%{USER:user}\\])?%{METRICSPACE}@%{METRICSPACE}%{HOSTNAME:clientHost}?%{METRICSPACE}\\[%{IP:clientIP}?\\]%{METRICSPACE}(Id:%{SPACE}%{NUMBER:connectionId}%{METRICSPACE})?(Thread_id:%{SPACE}%{NUMBER:connectionId}%{METRICSPACE})?" +
			"(Schema:%{SPACE}%{WORD:currentDB}?%{METRICSPACE})?(Last_errno: %{NUMBER:lastErrno}%{METRICSPACE})?(Killed: %{NUMBER:killed}%{METRICSPACE})?(QC_hit: %{WORD:queryCacheHit}%{METRICSPACE})?" +
			"(Query_time: %{NUMBER:queryTime}%{METRICSPACE})?(Lock_time: %{NUMBER:lockTime}%{METRICSPACE})?(Rows_sent: %{NUMBER:numRows}%{METRICSPACE})?(Rows_examined: %{NUMBER:examinedRows}%{METRICSPACE})?(Rows_affected: %{NUMBER:affectedRows}%{METRICSPACE})?" +
			"(Thread_id: %{NUMBER:connectionId}%{METRICSPACE})?(Errno: %{NUMBER:lastErrno}%{METRICSPACE})?(Killed: %{NUMBER:killed}%{METRICSPACE})?(Bytes_received: %{NUMBER:bytesReceived}%{METRICSPACE})?(Bytes_sent: %{NUMBER:bytesSent}%{METRICSPACE})?(Read_first: %{NUMBER:readFirst}%{METRICSPACE})?(Read_last: %{NUMBER:readLast}%{METRICSPACE})?(Read_key: %{NUMBER:readKey}%{METRICSPACE})?(Read_next: %{NUMBER:readNext}%{METRICSPACE})?(Read_prev: %{NUMBER:readPrev}%{METRICSPACE})?(Read_rnd: %{NUMBER:readRnd}%{METRICSPACE})?(Read_rnd_next: %{NUMBER:readRndNext}%{METRICSPACE})?(Sort_merge_passes: %{NUMBER:sortMergePasses}%{METRICSPACE})?(Sort_range_count: %{NUMBER:sortRangeCount}%{METRICSPACE})?(Sort_rows: %{NUMBER:sortRows}%{METRICSPACE})?(Sort_scan_count: %{NUMBER:sortScanCount}%{METRICSPACE})?(Created_tmp_disk_tables: %{NUMBER:tmpDiskTables}%{METRICSPACE})?(Created_tmp_tables: %{NUMBER:tmpTables}%{METRICSPACE})?(Tmp_tables: %{NUMBER:tmpTables}%{METRICSPACE})?(Tmp_disk_tables: %{NUMBER:tmpDiskTables}%{METRICSPACE})?(Tmp_table_sizes: %{NUMBER:tmpTableSizes}%{METRICSPACE})?(Start: %{TIMESTAMP_ISO8601:start}%{METRICSPACE})?(End: %{TIMESTAMP_ISO8601:end}%{METRICSPACE})?(InnoDB_trx_id: %{WORD:TrxId}%{METRICSPACE})?(QC_Hit: %{WORD:queryCacheHit}%{METRICSPACE})?(Full_scan: %{WORD:fullScan}%{METRICSPACE})?(Full_join: %{WORD:fullJoin}%{METRICSPACE})?(Tmp_table: %{WORD:tmp_table}%{METRICSPACE})?(Tmp_table_on_disk: %{WORD:tmp_table_on_disk}%{METRICSPACE})?(Filesort: %{WORD:filesort}%{METRICSPACE})?(Filesort_on_disk: %{WORD:filesort_on_disk}%{METRICSPACE})?(Merge_passes: %{NUMBER:merge_passes}%{METRICSPACE})?(Priority_queue: %{WORD:priority_queue}%{METRICSPACE})?(No InnoDB statistics available for this query%{METRICSPACE})?(InnoDB_IO_r_ops: %{NUMBER:innodb.io_r_ops}%{METRICSPACE})?(InnoDB_IO_r_bytes: %{NUMBER:innodb.io_r_bytes}%{METRICSPACE})?(InnoDB_IO_r_wait: %{NUMBER:innodb.io_r_wait.sec}%{METRICSPACE})?(InnoDB_rec_lock_wait: %{NUMBER:innodb.rec_lock_wait.sec}%{METRICSPACE})?(InnoDB_queue_wait: %{NUMBER:innodb.queue_wait.sec}%{METRICSPACE})?(InnoDB_pages_distinct: %{NUMBER:innodb.pages_distinct}%{METRICSPACE})?(Log_slow_rate_type: %{WORD:log_slow_rate_type}%{METRICSPACE})?(Log_slow_rate_limit: %{NUMBER:log_slow_rate_limit}%{METRICSPACE})?" +
			"%{EXPLAIN:explain}?(use %{WORD:currentDB};\n)?SET timestamp=%{NUMBER:timestamp};\n" +
			"%{GREEDYMULTILINE:query}",
	}

	slowLog    = "# User@Host: shopee_video_supply[shopee_video_supply] @  [10.174.235.159]  Id: 6296389\n# Schema: shopee_luckyvideo_supply_db  Last_errno: 0  Killed: 0\n# Query_time: 0.467243  Lock_time: 0.000058  Rows_sent: 2000  Rows_examined: 1337466  Rows_affected: 0\n# Bytes_sent: 103881\nSET timestamp=1683613152;\nselect id, user_id, user_type, ctime, utime from user_info_tab where id > 3094757 and user_type = 1 order by id asc limit 2000;"
	slowParser *SlowLogParser
	err        error
)

func TestNewParser(t *testing.T) {
	assert.NoError(t, err)
	assert.NotNil(t, slowParser)
}

func TestNewParserError(t *testing.T) {
	_, err := NewSlowLogParser(nil)
	assert.Error(t, err)
}

func TestParser_ParserSlowLog(t *testing.T) {
	infos, err := slowParser.ParserSlowLog(slowLog)
	assert.NoError(t, err)
	assert.NotEmpty(t, infos)
	fmt.Printf("slow log parser infos:%v \n", infos)
}

func init() {
	slowParser, err = NewSlowLogParser(slowPattern)
}
