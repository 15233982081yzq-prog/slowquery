#!/bin/sh
################ Version Info ##################
# Create Date: 2023-08-14
# Author:      BianJian
# Mail:        jian.bian@shoppe.com
# Version:     0.1 beta
# Attention:   wrk benchmark 压测
################################################
# 加载环境变量
# 如果脚本放到crontab中执行，会缺少环境变量，所以需要添加以下3行
. /etc/profile
. /etc/bashrc

# 脚本所在目录即脚本名称
script_dir=$( cd "$( dirname "$0"  )" && pwd )
script_name=$(basename ${0})
# 日志目录
log_dir="${script_dir}/log"
  [ ! -d ${log_dir} ] && {
    mkdir -p ${log_dir}
  }

errorMsg(){
  echo "USAGE:$0 arg1:connetion arg2:threads arg3:duration"
  exit 2
}


benchmark() {
  echo "benchmark to /rds/smart/v1/api/non-live/slowquery/query_list start"
  wrk -c$1 -t$2 -d$3s --timeout 20s --latency -s ./scripts/platform/query_list.lua http://10.129.120.132:30083/rds/smart/v1/api/non-live/slowquery/query_list >> $log_dir/benchmark.log
  echo "benchmark /rds/smart/v1/api/non-live/slowquery/query_list finish sleep 5s to next"
  sleep 5
  echo "benchmark to /rds/smart/v1/api/non-live/platform/database/hosts start"
  wrk -c$1 -t$2 -d$3s --timeout 20s --latency -s ./scripts/platform/database_hosts.lua http://10.129.120.132:30083/rds/smart/v1/api/non-live/platform/database/hosts >> $log_dir/benchmark.log
  echo "benchmark /rds/smart/v1/api/non-live/platform/database/hosts finish sleep 5s to next"
  sleep 5
  echo "benchmark to /rds/smart/v1/api/non-live/slowquery/query_detail start"
  wrk -c$1 -t$2 -d$3s --timeout 20s --latency -s ./scripts/platform/query_detail.lua http://10.129.120.132:30083/rds/smart/v1/api/non-live/slowquery/query_detail >> $log_dir/benchmark.log
  echo "benchmark /rds/smart/v1/api/non-live/slowquery/query_detail finish sleep 5s to next"
  sleep 5
  echo "benchmark get platform metrics start"
  curl http://10.129.120.132:30083/metrics >> $log_dir/benchmark.log
  echo "benchmark get platform metrics finish"
  sleep 5
  echo "finish all print benchmark report ..."
  cat $log_dir/benchmark.log
}

main() {
if [ $# -ne 3 ];then
	errorMsg
fi
rm -f $log_dir/benchmark_platform.log
benchmark $1 $2 $3
}

# 需要把隐号加上，不然传入的参数就不能有空格
main "$@"
