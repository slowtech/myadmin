package main

import (
	"os"
	"text/template"
	"github.com/slowtech/myadmin/common"
	"math/rand"
	"time"
	"strconv"
)

const config = `
[client]
socket = {{.DynamicParameters.datadir}}/mysql/3306/data/mysql.sock

[mysql]
no-auto-rehash

[mysqld]
#general
user = mysql
port = {{.DynamicParameters.port}}
basedir = {{.DynamicParameters.basedir}}
datadir = {{.DynamicParameters.datadir}}/mysql/3306/data
socket = {{.DynamicParameters.datadir}}/mysql/3306/data/mysql.sock
pid_file = {{.DynamicParameters.datadir}}/mysql/3306/data/mysql.pid
character_set_server = utf8mb4
transaction_isolation = READ-COMMITTED
sql_mode = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'
log_error = {{.DynamicParameters.datadir}}/mysql/3306/log/mysqld.log
skip-external-locking

#connection
back_log = 2048
max_connections = 500
max_connect_errors = 10000  
interactive_timeout = 1800
wait_timeout = 1800
thread_cache_size = 128
max_allowed_packet = 64M
skip_name_resolve

#session
read_buffer_size = 16M
read_rnd_buffer_size = 16M
sort_buffer_size = 16M
join_buffer_size = 32M


#innodb
innodb_buffer_pool_size = {{.DynamicParameters.innodb_buffer_pool_size}}
innodb_flush_log_at_trx_commit = 1
innodb_io_capacity = {{.DynamicParameters.innodb_io_capacity}}
innodb_io_capacity_max = {{.DynamicParameters.innodb_io_capacity_max}}
innodb_data_file_path = ibdata1:1G:autoextend
innodb_flush_method = O_DIRECT
innodb_log_file_size = {{.DynamicParameters.innodb_log_file_size}}
innodb_purge_threads = 4
innodb_autoinc_lock_mode = 2
innodb_buffer_pool_load_at_startup = 1
innodb_buffer_pool_dump_at_shutdown = 1
innodb_read_io_threads = 8
innodb_write_io_threads = 8
innodb_flush_neighbors = 0
innodb_page_cleaners = 8
innodb_print_all_deadlocks = 1
innodb_file_format = Barracuda
innodb_checksum_algorithm = crc32
innodb_strict_mode = ON
innodb_large_prefix = ON

#replication
server_id = {{.DynamicParameters.server_id}}
log_bin = {{.DynamicParameters.datadir}}/mysql/3306/log/mysql-bin
relay_log = {{.DynamicParameters.datadir}}/mysql/3306/log/
sync_binlog = 1
binlog_format = ROW
master_info_repository = TABLE
relay_log_info_repository = TABLE
relay_log_recovery = ON
log_slave_updates = ON
expire_logs_days = 7
slave-rows-search-algorithms = 'INDEX_SCAN,HASH_SCAN'
skip-slave-start
slave_net_timeout = 60
binlog_error_action = ABORT_SERVER

#semi sync replication
plugin_load = "validate_password.so;semisync_master.so;semisync_slave.so"
rpl_semi_sync_master_enabled = 1
rpl_semi_sync_slave_enabled = 1
rpl_semi_sync_master_timeout = 1000

#GTID
gtid_mode = ON
enforce_gtid_consistency = 1


#slow log 
slow_query_log = ON
long_query_time = 0.5
slow_query_log_file = {{.DynamicParameters.basedir}}/mysql/3306/log/slow.log

#others
open_files_limit = 65535
max_heap_table_size = 32M
tmp_table_size = 32M
table_open_cache = 65535
table_definition_cache = 65535
table_open_cache_instances = 64
event_scheduler = 1
eq_range_index_dive_limit = 200

[mysql-5.6]
#query cache
query_cache_type = 0
query_cache_size = 0

[mysql-5.7]
#query cache
query_cache_type = 0
query_cache_size = 0

#undo tablespace
innodb_undo_tablespaces = 2
innodb_max_undo_log_size = 128M
innodb_undo_log_truncate = 1
{{range  $k, $v := .ExtraParameters_57}}{{ $k }} = {{$v}}{{end}}

#multi-threaded slave
slave-parallel-type = LOGICAL_CLOCK
slave-parallel-workers = 8
slave_preserve_commit_order = 1

#others
log_timestamps = system
innodb_numa_interleave = ON
`

func main() {
	serverId := getServerId()
	var configTemp = template.Must(template.New("configfile").Parse(config))

	type Parameter struct {
		DynamicParameters map[string]string
		ExtraParameters_57 map[string]string
	}
	var parameter Parameter
	totalMem := common.GetTotalMem()
	parameter.DynamicParameters=make(map[string]string)
	parameter.DynamicParameters["basedir"] = "/usr/local/mysql"
	parameter.DynamicParameters["datadir"] = "/data"
	parameter.DynamicParameters["port"] = "3306"
	parameter.DynamicParameters["innodb_buffer_pool_size"] = strconv.Itoa(getInnodbBufferPoolSize(totalMem))+"M"
	parameter.DynamicParameters["server_id"] = serverId
	parameter.DynamicParameters["innodb_io_capacity"]="500"
	parameter.DynamicParameters["innodb_io_capacity_max"] = "1000"
	parameter.DynamicParameters["innodb_log_file_size"] = strconv.Itoa(getInnodbLogFileSize(totalMem))+"M"
	//parameter.ExtraParameters_57=make(map[string]string)
	//parameter.ExtraParameters_57["basedir"] = "/usr/local/mysql"
	configTemp.Execute(os.Stdout, parameter)
	//totalMem := getTotalMem()
	//var innodb_buffer_pool_size,innodb_buffer_pool_instances,innodb_log_file_size string
}

func getServerId() (string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := r.Intn(1000000)
	return strconv.Itoa(randNum)
}


func getInnodbBufferPoolSize(totalMem int) (innodb_buffer_pool_size int) {
	if totalMem < 1024 {
		innodb_buffer_pool_size = 128
	} else if totalMem <= 4*1024 {
		innodb_buffer_pool_size = int(float32(totalMem) * 0.5)
	} else {
		innodb_buffer_pool_size = int(float32(totalMem) * 0.75)
	}
	return
}

func getInnodbLogFileSize(totalMem int) (innodb_log_file_size int) {
	if totalMem < 1024 {
		innodb_log_file_size = 48
	} else if totalMem <= 4*1024  {
		innodb_log_file_size = 128
	} else  if totalMem <= 8*1024 {
		innodb_log_file_size = 512
	} else  if totalMem <= 16*1024 {
		innodb_log_file_size = 1024
	} else {
		innodb_log_file_size = 2048
	}
	return
}

