package configfile

var ConfString = `
debug: false
commit: 0
port: 3000

database:
    type: postgres
    port: 5432
    host: 127.0.0.1
    user: monkeyking
    password: kingofmonkeys
    dbname: monkeys

redis:
    url: redis://127.0.0.1:6379 
    index: 0
    pool_max_idle: 10000
    pool_max_active: 10000
    pool_idle_time_out: 300
`
