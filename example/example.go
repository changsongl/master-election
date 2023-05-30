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
