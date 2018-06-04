package main

import (
        "os"
        "text/template"
	"fmt"
        "github.com/slowtech/myadmin/common"
	"strings"
	"math/rand"
	"time"
	"strconv"
)

const config = `
[client]
socket = {{.datadir}}/mysql/3306/data/mysql.sock

[mysql]
no-auto-rehash

[mysqld]
#general
user = mysql
port = {{.port}}
socket = {{.datadir}}/mysql/3306/data/mysql.sock
pid_file = {{.datadir}}/mysql/3306/data/mysql.pid
character_set_server = utf8mb4
basedir = {{.basedir}}
datadir = {{.datadir}}/mysql/3306/data
log_error = {{.datadir}}/mysql/3306/log/mysqld.log
sql_mode = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'

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

#query cache
query_cache_type = 0
query_cache_size = 0

#innodb
innodb_buffer_pool_size = {{.innodb_buffer_pool_size}}
innodb_buffer_pool_instances = {{.innodb_buffer_pool_instances}}
innodb_flush_log_at_trx_commit = 1
innodb_io_capacity = {{.innodb_io_capacity}}
innodb_io_capacity_max = {{.innodb_io_capacity_max}}
innodb_data_file_path = ibdata1:1G:autoextend
innodb_flush_method = O_DIRECT
innodb_log_file_size = {{.innodb_log_file_size}}
innodb_purge_threads = 4
#innodb_undo_tablespaces = 3
#innodb_max_undo_log_size = 2G
#innodb_undo_log_truncate = 1
innodb_autoinc_lock_mode = 2
innodb_buffer_pool_load_at_startup = 1
innodb_buffer_pool_dump_at_shutdown = 1
innodb_read_io_threads = 8
innodb_write_io_threads = 8
innodb_flush_neighbors = 0
innodb_page_cleaners = 8
innodb_print_all_deadlocks = 1

#replication
server_id = {{.server_id}}
log_bin = {{.datadir}}/mysql/3306/log/mysql-bin
relay_log = {{.datadir}}/mysql/3306/log/
sync_binlog = 1
binlog_format = ROW
master_info_repository = TABLE
relay_log_info_repository = TABLE
relay_log_recovery = ON
log_slave_updates = ON
expire_logs_days = 7
slave-rows-search-algorithms = 'INDEX_SCAN,HASH_SCAN'
skip-slave-start

#semi sync replication
plugin_load = "validate_password.so;semisync_master.so;semisync_slave.so"
rpl_semi_sync_master_enabled = 1
rpl_semi_sync_slave_enabled = 1
rpl_semi_sync_master_timeout = 1000

#GTID
gtid_mode = ON
enforce_gtid_consistency = 1

#multi-threaded slave
slave-parallel-type = LOGICAL_CLOCK
slave-parallel-workers = {{.slave_parallel_workers}}
slave_preserve_commit_order=1

#slow log 
slow_query_log = ON
long_query_time = 0.5
slow_query_log_file = {{.datadir}}/mysql/3306/log/slow.log

#others
open_files_limit = 65535
max_heap_table_size = 32M
tmp_table_size = 32M
table_open_cache = 4096
table_definition_cache = 4096
table_open_cache_instances = 64
log_timestamps = system
event_scheduler = 1
innodb_numa_interleave = ON
`

func main(){
        serverId := getServerId()
 	fmt.Println(serverId)
        var configTemp = template.Must(template.New("configfile").Parse(config))
        var dynamicParameters = make(map[string]string)
        dynamicParameters["basedir"]="/usr/local/mysql"
        dynamicParameters["datadir"]="/data"
        dynamicParameters["port"]="3306"
        dynamicParameters["innodb_buffer_pool_size"]="10G"
        dynamicParameters["innodb_buffer_pool_instances"]="8"
        dynamicParameters["server_id"]=serverId
//        dynamicParameters["innodb_io_capacity"]="500"
        dynamicParameters["innodb_io_capacity_max"]="1000"
        dynamicParameters["innodb_log_file_size"]="1G"
//        dynamicParameters["slave-parallel-workers"]="8"
        
        
        configTemp.Execute(os.Stdout,dynamicParameters)
        //totalMem := getTotalMem()
	//var innodb_buffer_pool_size,innodb_buffer_pool_instances,innodb_log_file_size string
}

func getServerId()(string){
     	r := rand.New(rand.NewSource(time.Now().UnixNano()))
        randNum := r.Intn(1000000)
 	return strconv.Itoa(randNum)
}

func getInnodbBufferPoolSize(totalMem int)(innodb_buffer_pool_size int,innodb_buffer_pool_instances int) {
	if totalMem < 1024 {
                innodb_buffer_pool_size = 128
                innodb_buffer_pool_instances = 1
        } else if totalMem <= 4*1024 {
                innodb_buffer_pool_size = int(float32(totalMem)*0.5)
                innodb_buffer_pool_instances = 2
        } else {
		innodb_buffer_pool_size = int(float32(totalMem)*0.75) 	
	}	
	return 
}

func getInnodbLogFileSize(totalMem int) (innodb_log_file_size int) {
	if totalMem 
	
}

func getTotalMem()(int) {
	getMemoryCmd := `grep "MemTotal" /proc/meminfo`
	memoryResult,err := common.Run_cmd(getMemoryCmd)
	if err != nil {
        	fmt.Println(err)
	}
	totalMem := strings.Fields(memoryResult)[1]
        totalMemInt,_:=strconv.Atoi(totalMem)
        return totalMemInt/1024
}
