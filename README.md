# master-election

master-election是进行主从选举的工具包。在分布式场景，我们经常需要多个实例中的某一个实例去单独的运行某一串逻辑。
此包可以帮助你快速上手，并且忽略掉选主的具体细节。

### 支持组件
- Mysql (已完成)
- Redis (待开发)
- PostgreSQL (待开发)
- ETCD (待开发)
- Consul (待开发)
- Mongodb (待开发)

### 安装
````
go get github.com/changsongl/master-election
````

### 示例
```` go
package main

import (
	"fmt"
	"time"

	"github.com/changsongl/master-election"
	"github.com/changsongl/master-election/lock/mysql"
)

func main() {
	m, err := master.New(
		mysql.NewMasterLock(
			&mysql.Config{
				User:      "root",
				Password:  "123456",
				Host:      "127.0.0.1",
				Port:      3306,
				TableName: "master_election_table",
				DBName:    "master_election_database",
				CreateDB:  true,
			},
		),
		master.OptionHeartbeat(time.Second*2),
		master.OptionMasterStartHook(func(epoch uint64) {
			fmt.Printf("master start: epoch %d\n", epoch)
		}),
		master.OptionMasterEndHook(func(epoch uint64) {
			fmt.Printf("master end: epoch %d\n", epoch)
		}),
		// other option available
	)

	if err != nil {
		panic(err)
	}

	go func() {
		err = m.Start()
		if err != nil {
			panic(err)
		}
	}()

	for {
		if m.IsMaster() {
			// do some master logic
		} else {
			// do some slave logic or do nothing
		}
	}

	// m.Stop()
}
````

