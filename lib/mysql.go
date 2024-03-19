package lib

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"context"
	"strings"
)

type DBO struct {
	Client 			*sql.DB
	Context 		context.Context
	Version 		string
	VersionDiv 		int
}

type Bufferpool struct{
	BufferByteSize	int64
	BufferGBSize	float64
	PageSize 		int64
	TotalPage 		int64 
	UsePage			int64
	FreePage		int64
}

type TableInfo struct {
	TableName 		string 
	TableComment 	string
	TableRows		int64
}

func MySQLConnector(e string, pt int, u string, p string, d string) (*DBO, error) {
	DSN := "%s:%s@tcp(%s:%d)/%s"
	// Create DB Object
	dbObj, err := sql.Open("mysql",fmt.Sprintf(DSN,u,p,e,pt,d))
	if err != nil {
		return nil,err
	}

	// Connection Check
	err = dbObj.Ping()
	if err != nil {
		return nil, err
	}

	return &DBO{
		Client: dbObj,
		Context: context.Background(),
	}, nil
}

func (d *DBO) VersionChecker() error {
	err := d.Client.QueryRowContext(d.Context,"SELECT SUBSTRING_INDEX(VERSION(),'.',1),version()").Scan(&d.VersionDiv,&d.Version)
	if err != nil {
		return err
	}

	return nil
}

func (d *DBO) BufferStatus() (*Bufferpool, error) {
	var bfp Bufferpool 
	err := d.Client.QueryRowContext(d.Context,"SELECT @@innodb_buffer_pool_size").Scan(&bfp.BufferByteSize)
	if err != nil {
		return nil, err
	}

	bfp.BufferGBSize = float64(bfp.BufferByteSize /1024/1024/1024)

	GetStatus := `
		SHOW GLOBAL STATUS WHERE VARIABLE_NAME IN (
			'innodb_buffer_pool_pages_total',
			'innodb_buffer_pool_pages_data',
			'innodb_buffer_pool_pages_free'
		)
	`
	data, err := d.Client.QueryContext(d.Context,GetStatus)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	for data.Next() {
		var variable string 
		var values int64
		err := data.Scan(&variable,&values)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(variable) {
		case "innodb_buffer_pool_pages_total":
			bfp.TotalPage = values
		case "innodb_buffer_pool_pages_data":
			bfp.UsePage = values
		case "innodb_buffer_pool_pages_free":
			bfp.FreePage = values
		}
	}

	bfp.PageSize = bfp.BufferByteSize / bfp.TotalPage / 1024

	return &bfp, nil
}

func (d *DBO) GetTable(database string, tables []string)(*[]TableInfo, error) {
	GetQuery := `
		SELECT
			table_name,
			table_comment,
			table_rows
		FROM information_schema.tables WHERE table_schema = ? and table_name IN(%s)
		ORDER BY table_rows DESC
	`

	InCondition := InCondition(tables)

	data, err := d.Client.QueryContext(d.Context,fmt.Sprintf(GetQuery,InCondition),database)
	if err != nil {
		return nil, err
	}
	defer data.Close()

	var CheckTable []TableInfo
	for data.Next() {
		var tbl TableInfo
		err := data.Scan(
			&tbl.TableName,
			&tbl.TableComment,
			&tbl.TableRows,
		)
		if err != nil {
			return nil, err
		}

		CheckTable = append(CheckTable,tbl)
	}

	return &CheckTable, nil

}

func (d *DBO) BufferWarmingUp(table string, limit int64) (*int64, error) {
	Query := "SELECT * FROM %s limit %d"
	var cnt int64 
	data, err := d.Client.QueryContext(d.Context, fmt.Sprintf(Query,table,limit))
	if err != nil {
		return nil, err
	}
	defer data.Close()

	cnt = 0
	for data.Next() {
		cnt++
	}


	return &cnt, nil
}

func InCondition(s []string) string {
	var NewS []string
	for _, v := range s {
		NewS = append(NewS,fmt.Sprintf("'%s'",v))
	}

	return strings.Join(NewS,",")
}

func (b *Bufferpool) BufferPageRate() float64 {
	return (float64(b.UsePage) / float64(b.TotalPage))*100
}

func (d *DBO) ExecuteQuery(q string) error{
	_, err := d.Client.QueryContext(d.Context,q)
	if err != nil {
		return err
	}

	return nil

}