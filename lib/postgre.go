package lib 

import (
	_ "github.com/lib/pq"
	"database/sql"
	"fmt"
	"context"
)

func PostgreConnector(e string, pt int, u string, p string, d string) (*DBO, error) {
	DSN := "host=%s port=%d user=%s password=%s dbname=%s sslmode=disable"

	dbObj, err := sql.Open("postgres",fmt.Sprintf(DSN,e,pt,u,p,d))
	if err != nil {
		return nil, err
	}

	err = dbObj.Ping()
	if err != nil {
		return nil, err
	}

	return &DBO{
		Client: dbObj,
		Context: context.Background(),
	}, nil
}

func (d *DBO) OnBufferExtention() error {
	_, err := d.Client.Exec("CREATE EXTENSION IF NOT EXISTS pg_buffercache;")
	if err != nil {
		return err
	}

	return nil
}

func (d *DBO) OffBufferExtention() error {
	_, err := d.Client.Exec("DROP EXTENSION IF EXISTS pg_buffercache;")
	if err != nil {
		return err
	}

	return nil
}

func (d *DBO) GetBufferRatio() (*float64, error) {
	Query := `
	SELECT 
		trunc((count(*)::float / (
			select 
				setting::integer 
			from pg_settings 
			where name='shared_buffers')*100)::numeric,2) AS cache_filled
	FROM pg_buffercache b INNER JOIN pg_class c
	ON b.relfilenode = pg_relation_filenode(c.oid);
	`
	var ratio float64 
	err := d.Client.QueryRowContext(d.Context,Query).Scan(&ratio)
	if err != nil {
		return nil, err
	}

	return &ratio, nil

}

