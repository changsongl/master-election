package log

import (
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"testing"
)

func TestLogger(t *testing.T) {
	l := New(LevelInfo, nil)
	l.Debugf("this a debug msg")
	l.Infof("this a info msg")
	l.Errorf("this a error msg")

	l.Infof(os.Getenv("USERNAME"))
	dsn := fmt.Sprintf(
		"%s:%s@tcp(127.0.0.1:3306)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("USERNAME"),
		os.Getenv("PASSWORD"),
		os.Getenv("DATABASE"),
	)
	_, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	fmt.Println("err")
	fmt.Println(err)
}
