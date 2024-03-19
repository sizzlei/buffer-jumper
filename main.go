package main 

import (
	loader "github.com/sizzlei/confloader"
	"flag"
	log "github.com/sirupsen/logrus"
	// "fmt"
	"buffer-jumper/lib"
	"time"
	"strings"
)

func main() {
	var confPath string
	flag.StringVar(&confPath,"conf","","buffer-jumper Configure File Path")
	flag.Parse()

	// Configure Load
	conf, err := loader.FileLoader(confPath)
	if err != nil {
		log.Fatal(err)
	}

	// Configure Key List
	confList := conf.Conflist()
	cntLimit := int64(3000000)
	for _, t := range confList {
		dbc := conf.Keyload(t)

		log.Infof("Target %s Engine : %s",t,dbc["Engine"].(string))
		switch strings.ToLower(dbc["Engine"].(string)) {
		case "mysql":
			// Connect
			client, err := lib.MySQLConnector(
				dbc["Endpoint"].(string),
				dbc["Port"].(int),
				dbc["User"].(string),
				dbc["Pass"].(string),
				dbc["Database"].(string),
			)
			if err != nil {
				log.Fatal(err)
			}
			defer client.Client.Close()
		
			// Database Engine Check
			err = client.VersionChecker()
			if err != nil {
				log.Error(err)
			}

			// Get Buffer Stat
			log.Infof("Target Identifier : %s",dbc["Endpoint"].(string))
			bufferStat, err := client.BufferStatus()
			if err != nil {
				log.Error(err)
			}

			log.Info("AS-IS:")
			PrintBufferStat(bufferStat)
			bufferRate := bufferStat.BufferPageRate()
			log.Infof("Buffer Page Usage Rate : %.2f%%",bufferRate)
		
			// Using Table(Execute Query)
			if dbc["TableList"] != nil {
				tables := loader.InterfaceToSlice(dbc["TableList"])
				tableList,err := client.GetTable(dbc["Database"].(string),tables)
				if err != nil {
					log.Fatal(err)
				}
			
				log.Infof("Table Information: (%d)",len(*tableList))
				for i, t := range *tableList {
					log.Infof("%d : %s (%s) - %d",(i+1), t.TableName,t.TableComment, t.TableRows)
			
				}
				
				for _, t := range *tableList {
					if t.TableRows < cntLimit {
						start  := time.Now()
			
						cnt, err := client.BufferWarmingUp(t.TableName,cntLimit)
						if err != nil {
							log.Fatal(err)
						}
				
						end := time.Since(start)
						log.Infof("%s Count : %d (%s)",t.TableName, *cnt,end)
					} else {
						log.Warningf("Table %s Passing, Because Row Count Over %d",t.TableName, cntLimit)
					}
			
					bufferStat, err := client.BufferStatus()
					if err != nil {
						log.Error(err)
					}
			
					bufferRate := bufferStat.BufferPageRate()
			
					if bufferRate > 80 {
						log.Infof("Buffer Page Usage Rate : %.2f%%",bufferRate)
						break
					}
				}
			}

			if dbc["Queries"] != nil {
				queries := loader.InterfaceToSlice(dbc["Queries"])

				for _, q := range  queries {
					start  := time.Now()
					err := client.ExecuteQuery(q)
					if err != nil {
						log.Error(err)
					}
					end := time.Since(start)

					log.Info("Query:")
					log.Infof("%s",q)
					log.Infof("Execute Time : %s",end)

					bufferStat, err := client.BufferStatus()
					if err != nil {
						log.Error(err)
					}
			
					bufferRate := bufferStat.BufferPageRate()
			
					if bufferRate > 80 {
						log.Infof("Buffer Page Usage Rate : %.2f%%",bufferRate)
						break
					}
				}
			}
			
			// buffer result
			bufferStat, err = client.BufferStatus()
			if err != nil {
				log.Error(err)
			}
			log.Info("To-be:")
			PrintBufferStat(bufferStat)
			bufferRate = bufferStat.BufferPageRate()
			log.Infof("Buffer Page Usage Rate : %.2f%%",bufferRate)
			log.Info("")
		case "postgresql":
			// Connect
			pClient, err := lib.PostgreConnector(
				dbc["Endpoint"].(string),
				dbc["Port"].(int),
				dbc["User"].(string),
				dbc["Pass"].(string),
				dbc["Database"].(string),
			)
			if err != nil {
				log.Fatal(err)
			}
			defer pClient.Client.Close()

			log.Infof("Target Identifier : %s",dbc["Endpoint"].(string))
			err = pClient.OnBufferExtention()
			if err != nil {
				log.Fatal(err)
			}
			log.Info("Turn on Buffer Cache Extension")
			bufferRatio, err := pClient.GetBufferRatio()
			if err != nil {
				log.Error(err)
			}
			PostBufferStat(bufferRatio)

			// Using Table(Execute Count Query)
			if dbc["TableList"] != nil {
				tables := loader.InterfaceToSlice(dbc["TableList"])
				for _, t := range tables {
					start  := time.Now()
					cnt, err := pClient.BufferWarmingUp(t, cntLimit)
					if err != nil {
						log.Error(err)
					}
					end := time.Since(start)

					log.Infof("%s Count : %d (%s)",t, *cnt,end)

					bufferRatio, err = pClient.GetBufferRatio()
					if err != nil {
						log.Error(err)
					}

					if *bufferRatio > 80 {
						log.Infof("Buffer Page Usage Rate : %.2f%%",*bufferRatio)
						break
					}
				}
				
			}

			if dbc["Queries"] != nil {
				queries := loader.InterfaceToSlice(dbc["Queries"])

				for _, q := range  queries {
					start  := time.Now()
					err := pClient.ExecuteQuery(q)
					if err != nil {
						log.Error(err)
					}
					end := time.Since(start)

					log.Info("Query:")
					log.Infof("%s",q)
					log.Infof("Execute Time : %s",end)
				}
			}

			bufferRatio, err = pClient.GetBufferRatio()
			if err != nil {
				log.Error(err)
			}
			PostBufferStat(bufferRatio)

			err = pClient.OnBufferExtention()
			if err != nil {
				log.Fatal(err)
			}
			log.Info("Turn off Buffer Cache Extension")
			log.Info("")
		default:
			log.Errorf("Invalid Engine %s",dbc["Engine"].(string))
		}	
	}	
}

func PrintBufferStat(b *lib.Bufferpool) {
	log.Infof("InnoDB Buffer Pool Size : %dB(%.2f GB)",b.BufferByteSize,b.BufferGBSize)
	log.Infof("InnoDB Buffer Page Size : %dKB",b.PageSize)
	log.Info("innoDB Page Stat : ")
	log.Info("================================")
	log.Infof("Total page : %d",b.TotalPage)
	log.Infof("Used page : %d",b.UsePage)
	log.Infof("Free page : %d",b.FreePage)
	log.Info("================================")
}

func PostBufferStat(b *float64) {
	log.Info("================================")
	log.Infof("PostgreSQL Shared Buffer Usage Ratio : %.2f%%",*b)
	log.Info("================================")
}