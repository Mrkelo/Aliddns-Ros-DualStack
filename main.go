package main

import (
	middlewares "Aliddns-Ros/log-handler"
	"net"
	"net/http"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/alidns"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func init() {

}

// ConfigInfo 定义域名相关配置信息
type ConfigInfo struct {
	AccessKeyID     string
	AccessKeySecret string
	DomainName      string
	RR              string
	IpAddr          string
}

func main() {
	// 同时将日志写入文件和控制台
	//f, _ := os.Create("Aliddns_Get.log")
	//gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	//gin.SetMode(gin.ReleaseMode)
	//gin.DisableConsoleColor()

	r := gin.Default()
	r.Use(middlewares.Logger())
	r.GET("/aliddns", AddUpdateAliddns)
	r.Run(":8800")
}

func AddUpdateAliddns(c *gin.Context) {

	// 读取获取配置信息
	conf := new(ConfigInfo)
	conf.AccessKeyID = c.Query("AccessKeyID")
	conf.AccessKeySecret = c.Query("AccessKeySecret")
	conf.DomainName = c.Query("DomainName")
	conf.RR = c.Query("RR")
	conf.IpAddr = c.Query("IpAddr")

	// 识别IP类型
	ip := net.ParseIP(conf.IpAddr)
	if ip == nil {
		log.Println("IP地址格式错误：" + conf.IpAddr)
		c.String(http.StatusOK, "iperr")
		return
	}

	recordType := "A"
	if ip.To4() == nil {
		recordType = "AAAA"
	}

	//Info.Print("当前路由公网IP：" + conf.IpAddr)
	//log.SetOutput()
	log.Println("当前路由公网IP：" + conf.IpAddr + " 类型：" + recordType)
	log.Println("进行阿里云登录……")

	// 连接阿里云服务器，获取DNS信息
	// RegionId 填 "cn-hangzhou" 即可，全局通用
	client, err := alidns.NewClientWithAccessKey("cn-hangzhou", conf.AccessKeyID, conf.AccessKeySecret)
	if err != nil {
		log.Println("阿里云登录失败！", err)
		c.String(http.StatusOK, "loginerr")
		return
	}

	log.Println("阿里云登录成功！")
	log.Println("进行域名及IP比对……")

	// 查找现有记录
	request := alidns.CreateDescribeDomainRecordsRequest()
	request.DomainName = conf.DomainName
	request.RRKeyWord = conf.RR
	request.TypeKeyWord = recordType

	response, err := client.DescribeDomainRecords(request)
	if err != nil {
		log.Println("获取解析记录失败！", err)
		c.String(http.StatusOK, "finderr")
		return
	}

	var exsitRecordID string
	// var currentRecordValue string

	// 遍历查找匹配的记录 (虽然用了过滤，但保险起见还是确认一下)
	for _, record := range response.DomainRecords.Record {
		if record.RR == conf.RR && record.Type == recordType {
			if record.Value == conf.IpAddr {
				log.Println("当前配置解析地址与公网IP相同，不需要修改。")
				c.String(http.StatusOK, "same")
				return
			}
			exsitRecordID = record.RecordId
			break
		}
	}

	if len(exsitRecordID) > 0 {
		// 有配置记录，则匹配配置文件，进行更新操作
		updateRequest := alidns.CreateUpdateDomainRecordRequest()
		updateRequest.RecordId = exsitRecordID
		updateRequest.RR = conf.RR
		updateRequest.Type = recordType
		updateRequest.Value = conf.IpAddr

		rsp, err := client.UpdateDomainRecord(updateRequest)
		if err != nil {
			log.Println("修改解析地址信息失败!", err)
			c.String(http.StatusOK, "iperr")
		} else {
			log.Println("修改解析地址信息成功!", rsp)
			c.String(http.StatusOK, "ip")
		}
	} else {
		// 没有找到配置记录，那么就新增一个
		addRequest := alidns.CreateAddDomainRecordRequest()
		addRequest.DomainName = conf.DomainName
		addRequest.RR = conf.RR
		addRequest.Type = recordType
		addRequest.Value = conf.IpAddr

		rsp, err := client.AddDomainRecord(addRequest)
		if err != nil {
			log.Println("添加新域名解析失败！", err)
			c.String(http.StatusOK, "domainerr")
		} else {
			log.Println("添加新域名解析成功！", rsp)
			c.String(http.StatusOK, "domain")
		}
	}
}
