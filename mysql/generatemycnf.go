package mysql

import (
	"text/template"
	"github.com/slowtech/myadmin/common"
	"math/rand"
	"time"
	"strconv"
	"bytes"
	"bufio"
	"fmt"
	"regexp"
	"strings"
	"os"
)

const config = `
[client]
socket = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/data/mysql.sock

[mysql]
no-auto-rehash

[mysqld]
#general
user = mysql
port = {{.DynamicVariables.port}}
basedir = {{.DynamicVariables.basedir}}
datadir = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/data
socket = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/data/mysql.sock
pid_file = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/data/mysql.pid
character_set_server = utf8mb4
transaction_isolation = READ-COMMITTED
sql_mode = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'
log_error = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/log/mysqld.log
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
read_buffer_size = {{.DynamicVariables.read_buffer_size}}
read_rnd_buffer_size = {{.DynamicVariables.read_rnd_buffer_size}}
sort_buffer_size = {{.DynamicVariables.sort_buffer_size}}
join_buffer_size = {{.DynamicVariables.join_buffer_size}}


#innodb
innodb_buffer_pool_size = {{.DynamicVariables.innodb_buffer_pool_size}}
innodb_flush_log_at_trx_commit = 1
innodb_io_capacity = {{.DynamicVariables.innodb_io_capacity}}
innodb_io_capacity_max = {{.DynamicVariables.innodb_io_capacity_max}}
innodb_data_file_path = ibdata1:1G:autoextend
innodb_flush_method = O_DIRECT
innodb_log_file_size = {{.DynamicVariables.innodb_log_file_size}}
innodb_purge_threads = 4
innodb_autoinc_lock_mode = 2
innodb_buffer_pool_load_at_startup = 1
innodb_buffer_pool_dump_at_shutdown = 1
innodb_read_io_threads = 8
innodb_write_io_threads = 8
innodb_flush_neighbors = {{.DynamicVariables.innodb_flush_neighbors}}
innodb_print_all_deadlocks = 1
innodb_file_format = Barracuda
innodb_checksum_algorithm = crc32
innodb_strict_mode = ON
innodb_large_prefix = ON

#replication
server_id = {{.DynamicVariables.server_id}}
log_bin = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/log/mysql-bin
relay_log = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/log/relay-bin
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
slow_query_log_file = {{.DynamicVariables.datadir}}/mysql/{{.DynamicVariables.port}}/log/slow.log

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
innodb_max_undo_log_size = 1024M
innodb_undo_log_truncate = 1
{{range  $k, $v := .ExtraVariables_57}}{{ $k }} = {{$v}}{{end}}

#multi-threaded slave
slave-parallel-type = LOGICAL_CLOCK
slave-parallel-workers = 8
slave_preserve_commit_order = 1

#others
innodb_page_cleaners = 8
log_timestamps = system
innodb_numa_interleave = ON
`

func GenerateMyCnf(args map[string]string) (string) {
	serverId := getServerId()

	var totalMem int
	inputMem := args["memory"]
	if inputMem == "" {
		totalMem = common.GetTotalMem()
	} else {
		totalMem = formatMem(inputMem)
	}
	var mycnfTemplate = template.Must(template.New("mycnf").Parse(config))

	type Variable struct {
		DynamicVariables  map[string]string
		ExtraVariables_57 map[string]string
	}
	var variable Variable
	variable.DynamicVariables = make(map[string]string)
	variable.DynamicVariables["basedir"] = args["basedir"]
	variable.DynamicVariables["datadir"] = args["datadir"]
	variable.DynamicVariables["port"] = args["port"]
	variable.DynamicVariables["innodb_buffer_pool_size"] = strconv.Itoa(getInnodbBufferPoolSize(totalMem)) + "M"
	variable.DynamicVariables["server_id"] = serverId
	variable.DynamicVariables["innodb_flush_neighbors"] = "0"
	variable.DynamicVariables["innodb_io_capacity"] = "1000"
	variable.DynamicVariables["innodb_io_capacity_max"] = "2500"
	if args["ssd"] == "0" {
		variable.DynamicVariables["innodb_flush_neighbors"] = "1"
		variable.DynamicVariables["innodb_io_capacity"] = "200"
		variable.DynamicVariables["innodb_io_capacity_max"] = "500"
	}

	//Assume read_rnd_buffer_size==sort_buffer_size==join_buffer_size==read_buffer_size*2
	read_buffer_size := getReadBufferSize(totalMem)
	variable.DynamicVariables["read_buffer_size"] = strconv.Itoa(read_buffer_size) + "M"
	variable.DynamicVariables["read_rnd_buffer_size"] = strconv.Itoa(read_buffer_size*2) + "M"
	variable.DynamicVariables["sort_buffer_size"] = strconv.Itoa(read_buffer_size*2) + "M"
	variable.DynamicVariables["join_buffer_size"] = strconv.Itoa(read_buffer_size*2) + "M"
	variable.DynamicVariables["innodb_log_file_size"] = strconv.Itoa(getInnodbLogFileSize(totalMem)) + "M"
	//variable.ExtraVariables_57=make(map[string]string)
	//variable.ExtraVariables_57["basedir"] = "/usr/local/mysql"
	b := bytes.NewBuffer(make([]byte, 0))
	w := bufio.NewWriter(b)
	mycnfTemplate.Execute(w, variable)
	w.Flush()

	return b.String()
	//totalMem := getTotalMem()
	//var innodb_buffer_pool_size,innodb_buffer_pool_instances,innodb_log_file_size string
}

func getServerId() (string) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	randNum := r.Intn(1000000)
	return strconv.Itoa(randNum)
}

func getReadBufferSize(totalMem int) (read_buffer_size int) {
	innodb_buffer_pool_size := getInnodbBufferPoolSize(totalMem)
	freeSize := totalMem - innodb_buffer_pool_size
	//Assume read_rnd_buffer_size==sort_buffer_size==join_buffer_size==read_buffer_size*2
	//and max_connections=500
	if freeSize <= (2+4+4+4)*500 {
		read_buffer_size = 2
	} else if freeSize <= (4+8+8+8)*500 {
		read_buffer_size = 4
	} else if freeSize <= (8+16+16+16)*500 {
		read_buffer_size = 8
	} else {
		read_buffer_size = 16
	}
	return
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
	} else if totalMem <= 4*1024 {
		innodb_log_file_size = 128
	} else if totalMem <= 8*1024 {
		innodb_log_file_size = 512
	} else if totalMem <= 16*1024 {
		innodb_log_file_size = 1024
	} else {
		innodb_log_file_size = 2048
	}
	return
}

func formatMem(inputMem string) (totalMem int) {
	matched, _ := regexp.MatchString(`^(?i)\d+[M|G]B?$`, inputMem)
	if ! matched {
		fmt.Println(`Valid units for --memory are "M","G"`)
		os.Exit(1)
	}
	inputMemLower := strings.ToLower(inputMem)
	if strings.Contains(inputMemLower, "m") {
		inputMemLower = strings.Split(inputMemLower, "m")[0]

	} else if strings.Contains(inputMemLower, "g") {
		inputMemLower = strings.Split(inputMemLower, "g")[0]
		temp, _ := strconv.Atoi(inputMemLower)
		inputMemLower = strconv.Itoa(temp * 1024)
	}
	totalMem, _ = strconv.Atoi(inputMemLower)
	return
}

