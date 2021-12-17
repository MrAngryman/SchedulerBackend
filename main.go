package main

import (
	//"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	//"github.com/sirupsen/logrus"
	"log"
)

var PodCountMap map[string]int
var schedulerList []SchedulerListItem
//var nodeNameArr =[]string{"tljl1","tljl2","tljl3","tljl4","tljl5","tljl6","tljl7","tljl8"}
var nodeNameArr =[]string{"pitboss1","pitboss2","pitboss3","pitboss4","pitboss5"}
const MAXSPREADSCORE=100

type Score struct {
	CpuScore float64
	MemScore float64
	DiskScore float64
	DelayScore float64
}

type SchedulerListItem struct {
	ScoreMap map[string]Score `json:"ScoreMap"`
	PodName string  `json:"PodName"`
	PrioritizeScoreMap map[string]float64 `json:"PrioritizeScoreMap"`
	SpreadScoreMap map[string]float64 `json:"SpreadScore"`
	NodeArr []string `json:"NodeArr"`
}

func main() {

	schedulerList=make([]SchedulerListItem,0)
	PodCountMap=make(map[string]int)
	ResetPodCountMap(PodCountMap)
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		BodyLimit:             100 * 1024 * 1024,
	})

	//请求url := "http://127.0.0.1:30083/taskplan" 可通过curl发送请求
	app.Get("/resetPodCountMap", ResetMapHandler)

 	app.Post("/SchedulerListItem",SchedulerListItemHandler)
	// 启动服务
	app.Get("/getSchedulerList",getSchedulerListHandler)
	log.Fatal(app.Listen(fmt.Sprintf(":%d", 8889)))

}


func SchedulerListItemHandler(ctx *fiber.Ctx) error {
	//获取来自调度器的数据
	item := new(SchedulerListItem)
	err := ctx.BodyParser(item)
	if err !=nil {
		ctx.Status(400)
		return  ctx.SendString("Json parse err")
	}
	//item.NodeArr:=  item.ScoreMap.
	item.NodeArr=[]string{}
	for k,_:=range item.ScoreMap{
		if k!=""{
			//fmt.Println("k",k)
			//fmt.Println("k len",len(k))
			item.NodeArr=append(item.NodeArr,k)
		}
	}
	//计算prioritize分数
	item.PrioritizeScoreMap=getPrioritizeScore(item.ScoreMap)
	//计算Spread分数
	item.SpreadScoreMap=getSpreadScore(PodCountMap,item.NodeArr)
	//err:=ctx.JSON(subareajson)

	fmt.Println(item.PodName)
	fmt.Println(PodCountMap)
	fmt.Println(item.NodeArr)
	fmt.Println(item.ScoreMap)
	fmt.Println(item.SpreadScoreMap)
	fmt.Println(item.PrioritizeScoreMap)
	//预测放置结果
	putNode:=getPutNode(*item)
	fmt.Println("放置结果",putNode)
	PodCountMap[putNode]=PodCountMap[putNode]+1
	fmt.Println(PodCountMap)
	schedulerList=append(schedulerList,*item)
	return  ctx.SendStatus(200)
}

func getSpreadScore(countMap map[string]int,nodeArr []string) map[string]float64{
	spreadScoreMap:=make(map[string]float64)
	var maxCount int=0
	//var maxDelay float64
	for  _,v:=range nodeArr {
		count:=countMap[v]
		if count>maxCount {
			maxCount=count
		}
	}

	if maxCount==0 {
		for _,v:=range nodeArr {
			spreadScoreMap[v]=MAXSPREADSCORE
		}
	}else {
		for _,v:=range nodeArr {
			spreadScoreMap[v]=float64(MAXSPREADSCORE)-float64(MAXSPREADSCORE*countMap[v])/float64(maxCount)
		}
	}
	return  spreadScoreMap
}

func getPrioritizeScore (ScoreMap map[string]Score) map[string]float64{
	prioritizeScore:=make(map[string]float64)
	for k,v :=range ScoreMap {
		prioritizeScore[k]=v.CpuScore+v.MemScore+v.DiskScore+v.DelayScore
	}
	return prioritizeScore
}

func getPutNode (item SchedulerListItem) string {
	var max = float64(0)
	var putnode =""
	for _,nodeName := range item.NodeArr {
		score :=item.PrioritizeScoreMap[nodeName]+item.SpreadScoreMap[nodeName]
		if max < score {
			max =score
			putnode=nodeName
		}
	}
	return  putnode
}

//请求处理函数
func ResetMapHandler(ctx *fiber.Ctx) error {
	ResetPodCountMap(PodCountMap)
	return ctx.SendStatus(200)
}
func ResetPodCountMap(countMap map[string]int ) {
	for _,name :=range nodeNameArr {
		countMap[name]=0
	}
}


func getSchedulerListHandler(ctx *fiber.Ctx) error {

	return ctx.JSON(schedulerList)
}
