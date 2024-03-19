# Overview

---

Buffer-jumper는 MySQL / Postgresql을 사용하는 Database에 대한 웜업이 필요한 경우 자동으로 웜업을 수행하도록 하기 위해 개발된 웜업 툴입니다. 

> 일반적으로 웜업이 필요한 시점은 재부팅 또는 aws의 경우 인스턴스 타입 교체 등이며 자세한 버퍼 Clear 시점에 대해서는 별도로 기술합니다.

## 제약사항
 - DB Admin 계정으로 수행해야만 정상적으로 수행이 가능합니다.

# Configure

---

```yaml
Param:  
  - ConfigId : testDB1.employees
    Conf:
      Engine : mysql
      Endpoint: test-dba-am3.kr
      Port: 3306
      User: adm
      Pass: 
      Database: employees
      TableList:
        - employees
        - dept_emp
      Queries:
        - | 
          select 
          *
          from employees
        - | 
          select
          *
          from dept_emp
```

- ConfigId : 단순 Alias 입니다.
    - Conf
        - Endpoint: DB 접속 Host
        - Port : DB 접속 Port
        - User: 접속 계정
        - Pass : 접속 계정 패스워드
        - Database : 조회 대상 데이터베이스 명
        - TableList(Array)
            - 조회할 테이블 목록
        - Queries(Array)
            - 수행할 쿼리 목록
            - 작성시 Yaml Multi line 규칙을 따릅니다.

## 참고사항

- Table List와 Queries 목록이 모두 작성된 경우 모두 수행합니다.
- 단, Buffer Page Usage Ratio가 80%를 넘는 경우 Warm Up은 중단됩니다.
- Table List와 Queries중 하나만 수행할 경우 Configure에서 변수를 제거하거나, 배열 구분자를 제거해야합니다.
- 여러 데이터베이스를 순차 실행하려면 ConfigId를 기준으로 추가 작성하면 가능합니다.
- 동시에 실행하는 경우 여러 프로세스를 실행하여 수행해야 합니다.

> TableList는 테이블에 대한 Full Scan Count를 수행하는 목록입니다.

> Queries는 작성된 쿼리를 기준으로 1회씩 수행합니다.

# Usage Command

---

```yaml
./innopump -conf=./conf.yml
```

## Example

### Using Table List

```bash
INFO[0000] Target Identifier : test-dba-am3 
INFO[0000] AS-IS:                                       
INFO[0000] InnoDB Buffer Pool Size : 8403288064B(7.00 GB) 
INFO[0000] InnoDB Buffer Page Size : 16KB               
INFO[0000] innoDB Page Stat :                           
INFO[0000] ================================             
INFO[0000] Total page : 512896                          
INFO[0000] Used page : 1041                             
INFO[0000] Free page : 511761                           
INFO[0000] ================================             
INFO[0000] Buffer Page Usage Rate : 0.20%               
INFO[0000] Table Information: (4)                       
INFO[0000] 1 : sbtest2 () - 2886243                     
INFO[0000] 2 : sbtest4 () - 2883669                     
INFO[0000] 3 : sbtest1 () - 2051529                     
INFO[0000] 4 : sbtest3 () - 831373                      
INFO[0016] sbtest2 Count : 3000000 (16.361930416s)      
INFO[0032] sbtest4 Count : 3000000 (16.035962583s)      
INFO[0046] sbtest1 Count : 3000000 (13.264197917s)      
INFO[0059] sbtest3 Count : 3000000 (13.113851208s)      
INFO[0059] To-be:                                       
INFO[0059] InnoDB Buffer Pool Size : 8403288064B(7.00 GB) 
INFO[0059] InnoDB Buffer Page Size : 16KB               
INFO[0059] innoDB Page Stat :                           
INFO[0059] ================================             
INFO[0059] Total page : 512896                          
INFO[0059] Used page : 165578                           
INFO[0059] Free page : 347224                           
INFO[0059] ================================             
INFO[0059] Buffer Page Usage Rate : 32.28% 
```

### Using Queries

```bash
INFO[0000] Target Identifier : test-dba-am3
INFO[0000] AS-IS:                                       
INFO[0000] InnoDB Buffer Pool Size : 8403288064B(7.00 GB) 
INFO[0000] InnoDB Buffer Page Size : 16KB               
INFO[0000] innoDB Page Stat :                           
INFO[0000] ================================             
INFO[0000] Total page : 512896                          
INFO[0000] Used page : 165578                           
INFO[0000] Free page : 347224                           
INFO[0000] ================================             
INFO[0000] Buffer Page Usage Rate : 32.28%              
INFO[0000] Query:                                       
INFO[0000] select 
*
from employees                     
INFO[0000] Execute Time : 445.839958ms                  
INFO[0001] Query:                                       
INFO[0001] select
*
from dept_emp                       
INFO[0001] Execute Time : 15.405791ms                   
INFO[0001] To-be:                                       
INFO[0001] InnoDB Buffer Pool Size : 8403288064B(7.00 GB) 
INFO[0001] InnoDB Buffer Page Size : 16KB               
INFO[0001] innoDB Page Stat :                           
INFO[0001] ================================             
INFO[0001] Total page : 512896                          
INFO[0001] Used page : 165784                           
INFO[0001] Free page : 347018                           
INFO[0001] ================================             
INFO[0001] Buffer Page Usage Rate : 32.32%
```
### Using Queries

```bash
INFO[0000] Target r1 Engine : postgres                  
INFO[0000] Target Identifier : d 
INFO[0000] Turn on Buffer Cache Extension               
INFO[0001] PostgreSQL Shared Buffer Usage Ratio : 1.76% 
INFO[0002] T1 Count : 13470 (1.529343375s) 
INFO[0003] T2 Count : 18216 (833.190834ms) 
INFO[0004] T3 Count : 21868 (405.679667ms) 
INFO[0005] T4 Count : 12131 (350.704208ms)    
INFO[0005] PostgreSQL Shared Buffer Usage Ratio : 1.76% 
INFO[0005] Turn off Buffer Cache Extension   

```

### 순차 실행 Config Example

```yaml
Param:  
  - ConfigId : testDB1.employees
    Conf:
      Endpoint: 
      Port: 3306
      User: adm
      Pass: 
      Database: employees
      TableList:
      Queries:
        - | 
          select 
          *
          from employees
        - | 
          select
          *
          from dept_emp

  - ConfigId : testDB1.sysbench
    Conf:
      Endpoint: 
      Port: 3306
      User: adm
      Pass: 
      Database: sysbench
      TableList:
        - sbtest1
        - sbtest2
        - sbtest3
        - sbtest4
```