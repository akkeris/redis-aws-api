package main

import "github.com/go-martini/martini"
import "github.com/martini-contrib/render"
import "github.com/martini-contrib/binding"
import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/aws/session"
import "github.com/aws/aws-sdk-go/service/elasticache"
import "fmt"
import "strconv"
import "database/sql"
import _ "github.com/lib/pq"
import "os"


type provisionspec struct {
	Plan        string `json:"plan"`
	Billingcode string `json:"billingcode"`
}

type tagspec struct {
    Resource string `json:"resource"`
    Name     string `json:"name"`
    Value    string `json:"value"`
}

func tag(spec tagspec, berr binding.Errors, r render.Render) {
    if berr != nil {
        fmt.Println(berr)
        errorout := make(map[string]interface{})
        errorout["error"] = berr
        r.JSON(500, errorout)
        return
    }
    svc := elasticache.New(session.New(&aws.Config{
        Region: aws.String(os.Getenv("REGION")),
    }))
    region := os.Getenv("REGION")
    accountnumber := os.Getenv("ACCOUNTNUMBER")
    name := spec.Resource

    arnname := "arn:aws:elasticache:" + region + ":" + accountnumber + ":cluster:" + name

    params := &elasticache.AddTagsToResourceInput{
        ResourceName: aws.String(arnname),
        Tags: []*elasticache.Tag{ // Required
            {
                Key:   aws.String(spec.Name),
                Value: aws.String(spec.Value),
            },
        },
    }
    _, err := svc.AddTagsToResource(params)

    if err != nil {
        fmt.Println(err.Error())
                errorout := make(map[string]interface{})
                errorout["error"] = berr
                r.JSON(500, errorout)
        return
    }

        r.JSON(200, map[string]interface{}{"response": "tag added"})
}



func provision(spec provisionspec, err binding.Errors, r render.Render) {
	plan := spec.Plan
	billingcode := spec.Billingcode

	brokerdb := os.Getenv("BROKERDB")
	uri := brokerdb
	db, dberr := sql.Open("postgres", uri)
	if dberr != nil {
		fmt.Println(dberr)
                toreturn := dberr.Error()
                r.JSON(500, map[string]interface{}{"error": toreturn})
                return
	}
	var name string
	dberr = db.QueryRow("select name from provision where plan='" + plan + "' and claimed='no' and make_date=(select min(make_date) from provision where plan='" + plan + "' and claimed='no')").Scan(&name)

	if dberr != nil {
		fmt.Println(dberr)
                toreturn := dberr.Error()
                r.JSON(500, map[string]interface{}{"error": toreturn})
                return
	}

        available := isAvailable(name)
        if(available) {
	stmt, dberr := db.Prepare("update provision set claimed=$1 where name=$2")

	if dberr != nil {
		fmt.Println(dberr)
                toreturn := dberr.Error()
                r.JSON(500, map[string]interface{}{"error": toreturn})
                return
	}
	_, dberr = stmt.Exec("yes", name)
	if dberr != nil {
		fmt.Println(dberr)
                toreturn := dberr.Error()
                r.JSON(500, map[string]interface{}{"error": toreturn})
                return
	}

	region := os.Getenv("REGION")
	svc := elasticache.New(session.New(&aws.Config{
		Region: aws.String(region),
	}))
	accountnumber := os.Getenv("ACCOUNTNUMBER")
	arnname := "arn:aws:elasticache:" + region + ":" + accountnumber + ":cluster:" + name

	params := &elasticache.AddTagsToResourceInput{
		ResourceName: aws.String(arnname),
		Tags: []*elasticache.Tag{ // Required
			{
				Key:   aws.String("billingcode"),
				Value: aws.String(billingcode),
			},
		},
	}
	_, awserr := svc.AddTagsToResource(params)

	if awserr != nil {
		fmt.Println(awserr.Error())
                toreturn := awserr.Error()
                r.JSON(500, map[string]interface{}{"error": toreturn})
		return
	}

	eparams := &elasticache.DescribeCacheClustersInput{
		CacheClusterId:    aws.String(name),
		MaxRecords:        aws.Int64(20),
		ShowCacheNodeInfo: aws.Bool(true),
	}
	eresp, awserr := svc.DescribeCacheClusters(eparams)
	if awserr != nil {
		toreturn := awserr.Error()
		r.JSON(500, map[string]interface{}{"error": toreturn})
		return
	}
	endpointhost := *eresp.CacheClusters[0].CacheNodes[0].Endpoint.Address
	endpointport := *eresp.CacheClusters[0].CacheNodes[0].Endpoint.Port
	stringport := strconv.FormatInt(endpointport, 10)
	r.JSON(200, map[string]string{"REDIS_URL": "redis://" + endpointhost + ":" + stringport})
        return
        }
        if (!available){
               r.JSON(503, map[string]string{"REDIS_URL": ""}) 
               return
        }

}

func main() {
	region := os.Getenv("REGION")
	svc := elasticache.New(session.New(&aws.Config{
		Region: aws.String(region),
	}))

	m := martini.Classic()
	m.Use(render.Renderer())

	m.Post("/v1/redis/instance", binding.Json(provisionspec{}), provision)

	m.Delete("/v1/redis/instance/:name", func(params martini.Params, r render.Render) {
		name := params["name"]
		dparams := &elasticache.DeleteCacheClusterInput{
			CacheClusterId: aws.String(name), // Required
		}
		dresp, derr := svc.DeleteCacheCluster(dparams)

		if derr != nil {
			fmt.Println(derr.Error())
			errorout := make(map[string]interface{})
			errorout["error"] = derr.Error()
			r.JSON(500, errorout)
			return
		}
		brokerdb := os.Getenv("BROKERDB")
		uri := brokerdb
		db, dberr := sql.Open("postgres", uri)
		if dberr != nil {
			fmt.Println(dberr)
                        toreturn := dberr.Error()
                        r.JSON(500, map[string]interface{}{"error": toreturn})
                        return
		}

		stmt, err := db.Prepare("delete from provision where name=$1")
                if err != nil {
                        errorout := make(map[string]interface{})
                        errorout["error"] = err.Error()
                        r.JSON(500, errorout)
                        return
                }
		res, err := stmt.Exec(name)
                if err != nil {
                        errorout := make(map[string]interface{})
                        errorout["error"] = err.Error()
                        r.JSON(500, errorout)
                        return
                }
		_, err = res.RowsAffected()
		if err != nil {
			errorout := make(map[string]interface{})
			errorout["error"] = err.Error()
			r.JSON(500, errorout)
			return
		}
		r.JSON(200, dresp)
	})
	m.Get("/v1/redis/plans", func(r render.Render) {
		plans := make(map[string]interface{})
		plans["small"] = "Small - 1x CPU - 0.6 GB "
		plans["medium"] = "Medium - 2x CPU - 3.2 GB"
		plans["large"] = "Large - 2x CPU 6 GB"
		r.JSON(200, plans)
	})

	m.Get("/v1/redis/url/:name", func(params martini.Params, r render.Render) {
		name := params["name"]
		eparams := &elasticache.DescribeCacheClustersInput{
			CacheClusterId:    aws.String(name),
			MaxRecords:        aws.Int64(20),
			ShowCacheNodeInfo: aws.Bool(true),
		}
		resp, err := svc.DescribeCacheClusters(eparams)
		if err != nil {
			toreturn := err.Error()
			r.JSON(200, map[string]interface{}{"error": toreturn})
			return
		}
		endpointhost := *resp.CacheClusters[0].CacheNodes[0].Endpoint.Address
		endpointport := *resp.CacheClusters[0].CacheNodes[0].Endpoint.Port
		stringport := strconv.FormatInt(endpointport, 10)
		r.JSON(200, map[string]string{"REDIS_URL": "redis://" + endpointhost + ":" + stringport})
	})
    m.Post("/v1/tag", binding.Json(tagspec{}), tag)
	m.Run()

}


func isAvailable(name string) bool {
    fmt.Println("Checking if " + name + " is available ...")
    var toreturn bool

    region := os.Getenv("REGION")

    svc := elasticache.New(session.New(&aws.Config{
        Region: aws.String(region),
    }))

    params := &elasticache.DescribeCacheClustersInput{
        CacheClusterId:    aws.String(name),
        MaxRecords:        aws.Int64(20),
        ShowCacheNodeInfo: aws.Bool(true),
    }
    resp, err := svc.DescribeCacheClusters(params)

    if err != nil {
        fmt.Println(err.Error())
        return false
    }


    status := *resp.CacheClusters[0].CacheClusterStatus
    if status == "available" {
        toreturn = true
    }
    if status != "available" {
        toreturn = false
    }
    return toreturn

}

