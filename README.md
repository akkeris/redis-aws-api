## Synopsis

Docker image which runs an http server with REST interface for provisioning of redis clusters on AWS ElastiCache

## Details

Listens on Port 3000
Supports the following

1. GET /v1/redis/plans
2. POST /v1/redis/instance/ with JSON data of plan and billingcode
3. DELETE /v1/redis/intance/:name
4. GET /v1/redis/url/:name


## Dependencies

1. "github.com/go-martini/martini"
2. "github.com/martini-contrib/render"
3. "github.com/martini-contrib/binding"
4. "github.com/aws/aws-sdk-go/aws"
5. "github.com/aws/aws-sdk-go/aws/session"
6. "github.com/aws/aws-sdk-go/service/elasticache"
7. "fmt"
8. "strconv"
9. "database/sql"
10. "github.com/lib/pq"
11. "os"



## Requirements
go

aws creds

## Runtime Environment Variables
1. ACCOUNTNUMBER
2. BROKERDB
3. REGION


## Examples
`curl hostname:3000/v1/redis/plans`

returns:
`{
  "large": "Large",
  "medium": "Medium",
  "small": "Small"
}`

`curl -X POST -d '{"plan":"small","billingcode":"gwp"}' hostname:3000/v1/redis/instance`

returns
`{"REDIS_URL":"redis://oct-f592509d.fuetsj.0001.usw2.cache.amazonaws.com:6379"}`

`curl hostname"3000/v1/redis/url/oct-f592509d`

returns `{"REDIS_URL":"redis://oct-f592509d.fuetsj.0001.usw2.cache.amazonaws.com:6379"}`


`curl -X DELETE oct-redisbroker.octanner.io/v1/redis/instance/oct-f592509d` 

returns
`
{"CacheCluster":{"AutoMinorVersionUpgrade":true,"CacheClusterCreateTime":"2016-05-20T04:28:03.758Z","CacheClusterId":"oct-f592509d","CacheClusterStatus":"deleting","CacheNodeType":"cache.t2.micro","CacheNodes":null,"CacheParameterGroup":{"CacheNodeIdsToReboot":null,"CacheParameterGroupName":"redis-28-small","ParameterApplyStatus":"in-sync"},"CacheSecurityGroups":null,"CacheSubnetGroupName":"redis-subnet-group","ClientDownloadLandingPage":"https://console.aws.amazon.com/elasticache/home#client-download:","ConfigurationEndpoint":null,"Engine":"redis","EngineVersion":"2.8.24","NotificationConfiguration":null,"NumCacheNodes":1,"PendingModifiedValues":{"CacheNodeIdsToRemove":null,"CacheNodeType":null,"EngineVersion":null,"NumCacheNodes":null},"PreferredAvailabilityZone":"us-west-2b","PreferredMaintenanceWindow":"sun:11:30-sun:12:30","ReplicationGroupId":null,"SecurityGroups":[{"SecurityGroupId":"sg-e3013b84","Status":"active"}],"SnapshotRetentionLimit":null,"SnapshotWindow":null}}
`


