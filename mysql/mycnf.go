package mysql

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/slowtech/myadmin/common"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

const config = `
[client]
socket = {{.datadir}}/mysql/{{.port}}/data/mysql.sock

[mysql]
no-auto-rehash

[mysqld]
#general
user = mysql
port = {{.port}}
basedir = {{.basedir}}
datadir = {{.datadir}}/mysql/{{.port}}/data
socket = {{.datadir}}/mysql/{{.port}}/data/mysql.sock
pid_file = {{.datadir}}/mysql/{{.port}}/data/mysql.pid
character_set_server = utf8mb4
transaction_isolation = READ-COMMITTED
sql_mode = 'ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES,NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION'
log_error = {{.datadir}}/mysql/{{.port}}/log/mysqld.log
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
read_buffer_size = {{.read_buffer_size}}
read_rnd_buffer_size = {{.read_rnd_buffer_size}}
sort_buffer_size = {{.sort_buffer_size}}
join_buffer_size = {{.join_buffer_size}}

#innodb
innodb_buffer_pool_size = {{.innodb_buffer_pool_size}}
innodb_log_file_size = {{.innodb_log_file_size}}
innodb_flush_log_at_trx_commit = 1{{if or (.mysqld57) (.mysqld80)}}
innodb_undo_tablespaces = 2
innodb_max_undo_log_size = 1024M
innodb_undo_log_truncate = 1
innodb_page_cleaners = 8
{{end}}innodb_io_capacity = {{.innodb_io_capacity}}
innodb_io_capacity_max = {{.innodb_io_capacity_max}}
innodb_data_file_path = ibdata1:1G:autoextend
innodb_flush_method = O_DIRECT
innodb_purge_threads = 4
innodb_autoinc_lock_mode = 2
innodb_buffer_pool_load_at_startup = 1
innodb_buffer_pool_dump_at_shutdown = 1
innodb_read_io_threads = 8
innodb_write_io_threads = 8
innodb_flush_neighbors = {{.innodb_flush_neighbors}}
innodb_checksum_algorithm = crc32
innodb_strict_mode = ON{{if or (.mysqld56) (.mysqld57)}}
innodb_file_format = Barracuda
innodb_large_prefix = ON{{end}}


#replication
server_id = {{.server_id}}
log_bin = {{.datadir}}/mysql/{{.port}}/log/mysql-bin
relay_log = {{.datadir}}/mysql/{{.port}}/log/relay-bin
sync_binlog = 1
binlog_format = ROW
master_info_repository = TABLE
relay_log_info_repository = TABLE
relay_log_recovery = ON
log_slave_updates = ON
expire_logs_days = 7
slave_rows_search_algorithms = 'INDEX_SCAN,HASH_SCAN'
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

{{if or (.mysqld57) (.mysqld80)}}
#multi-threaded slave
slave-parallel-type = LOGICAL_CLOCK
slave-parallel-workers = 8
slave_preserve_commit_order = 1
{{end}}

#slow log 
slow_query_log = ON
long_query_time = 0.5
slow_query_log_file = {{.datadir}}/mysql/{{.port}}/log/slow.log

#general log
general_log = OFF
general_log_file = {{.datadir}}/mysql/{{.port}}/log/general.log

{{if or (.mysqld56) (.mysqld57)}}
#query cache
query_cache_type = 0
query_cache_size = 0
{{end}}

#others
open_files_limit = 65535
max_heap_table_size = 32M
tmp_table_size = 32M
table_open_cache = 65535
table_definition_cache = 65535
table_open_cache_instances = 64
event_scheduler = 1
eq_range_index_dive_limit = 200
loose-innodb_numa_interleave = ON
log_bin_trust_function_creators = 1{{if or (.mysqld57) (.mysqld80)}}
log_timestamps = system
{{end}}
`

//{{range  $k, $v := .ExtraVariables_57}}{{ $k }} = {{$v}}{{end}}

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

	//type Variable struct {
	//	DynamicVariables  map[string]interface{}
	//	//ExtraVariables_57 map[string]interface{}
	//}

	dynamicvariables:= make(map[string]interface{})
	dynamicvariables["basedir"] = args["basedir"]
	dynamicvariables["datadir"] = args["datadir"]
	dynamicvariables["port"] = args["port"]
	dynamicvariables["innodb_buffer_pool_size"] = strconv.Itoa(getInnodbBufferPoolSize(totalMem)) + "M"
	dynamicvariables["server_id"] = serverId
	dynamicvariables["innodb_flush_neighbors"] = "0"
	dynamicvariables["innodb_io_capacity"] = "1000"
	dynamicvariables["innodb_io_capacity_max"] = "2500"
	if args["mysqld_version"] == "5.6" {
		dynamicvariables["mysqld56"] = true
	} else if args["mysqld_version"] == "5.7" {
		dynamicvariables["mysqld57"] = true
	} else {
		dynamicvariables["mysqld80"] = true
	}
	if args["ssd"] == "0" {
		dynamicvariables["innodb_flush_neighbors"] = "1"
		dynamicvariables["innodb_io_capacity"] = "200"
		dynamicvariables["innodb_io_capacity_max"] = "500"
	}

	//Assume read_rnd_buffer_size==sort_buffer_size==join_buffer_size==read_buffer_size*2
	read_buffer_size := getReadBufferSize(totalMem)
	dynamicvariables["read_buffer_size"] = strconv.Itoa(read_buffer_size) + "M"
	dynamicvariables["read_rnd_buffer_size"] = strconv.Itoa(read_buffer_size*2) + "M"
	dynamicvariables["sort_buffer_size"] = strconv.Itoa(read_buffer_size*2) + "M"
	dynamicvariables["join_buffer_size"] = strconv.Itoa(read_buffer_size*2) + "M"
	dynamicvariables["innodb_log_file_size"] = strconv.Itoa(getInnodbLogFileSize(totalMem)) + "M"
	//variable.ExtraVariables_57=make(map[string]string)
	//variable.ExtraVariables_57["basedir"] = "/usr/local/mysql"
	b := bytes.NewBuffer(make([]byte, 0))
	w := bufio.NewWriter(b)
	mycnfTemplate.Execute(w, dynamicvariables)
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

//func main() {
//
//	mycnf_args := make(map[string]string)
//	mycnf_args["basedir"] = "/usr/local/mysql"
//	mycnf_args["datadir"] = "/data"
//	mycnf_args["port"] = strconv.Itoa(3306)
//	mycnf_args["memory"] = "10G"
//	mycnf_args["mysqld_version"] = "5.6"
//	mycnf_args["ssd"] = "0"
//
//	mycnf := GenerateMyCnf(mycnf_args)
//	fmt.Println(mycnf)
//
//}
