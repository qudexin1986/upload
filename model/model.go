package model

type Files struct {
	Id      int64  `xorm:"BIGINT(20)"`
	Name    string `xorm:"not null default '0' VARCHAR(200)"`
	Addr    string `xorm:"not null default '0' VARCHAR(400)"`
	Type    string `xorm:"not null default '0' VARCHAR(10)"`
	Addtime int64  `xorm:"not null default 0 INT(11)"`
	Hash    string `xorm:"not null default '0' VARCHAR(100)"`
	Size    int    `xorm:"not null default 0 INT(11)"`
	Status  int    `xorm:"not null default 1 INT(11)"`
}

type ShowFile struct {
	Id      int64
	Name    string
	Addr    string
	Type    string
	Size    int
	Addtime string
}

