#!/bin/sh
################ Version Info ##################
# Create Date: 2024-01-09
# Author:      BianJian
# Mail:        jian.bian@shoppe.com
# Version:     0.3 beta
# Attention:   wrk benchmark 压测 for slow query openapi
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
  sleep 5
  echo "benchmark to openapi /rds/smart/v1/openapi/live/slowquery/query_db_statistic start"
  wrk -c$1 -t$2 -d$3s --timeout 3s --latency -s ./scripts/openapi/query_db_statistic.lua http://space.shopee.io/rds/smart/v1/openapi/live/slowquery/query_db_statistic >> $log_dir/benchmark_openapi.log
  echo "benchmark to openapi /rds/smart/v1/openapi/live/slowquery/query_db_statistic finish"
  sleep 5
  echo "finish all print openapi benchmark report ..."
  cat $log_dir/benchmark_openapi.log
}

main() {
if [ $# -ne 3 ];then
	errorMsg
fi
rm -f $log_dir/benchmark_openapi.log
benchmark $1 $2 $3
}

# 需要把隐号加上，不然传入的参数就不能有空格
main "$@"
