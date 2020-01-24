package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

//定时任务
type Job struct {
	Name     string `json:name`     //任务名
	Command  string `json:command`  //shell命令
	CronExpr string `json:cronexpr` //cron表达式
}

//任务调度计划
type JobSchedulerPlan struct {
	Job      *Job                 //任务信息
	Expr     *cronexpr.Expression //解析好的cronexpr表达式
	NextTime time.Time            //下次调度时间
}

//变化事件
type JobEvent struct {
	EventType int  //SAVE DELETE
	Job       *Job //事件的信息
}

//http接口应答
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

//任务执行状态
type JobExecuteInfo struct {
	Job      *Job      //任务信息
	PlanTime time.Time //理论上的调度时间
	RealTime time.Time //实际调度时间(在理论调度时间上有微小的误差)
}

//任务执行结果
type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo //执行的状态
	OutPut      []byte          //脚本的输出
	Err         error           //脚本的错误原因
	StartTime   time.Time       //启动时间
	EndTime     time.Time       //结束时间
}

//构建应答
func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	//定义response
	var (
		response Response
	)
	//赋值
	response.Errno = errno
	response.Msg = msg
	response.Data = data
	//序列化为json
	resp, err = json.Marshal(response)
	return
}

//反序列化job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)
	job = &Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

//从etcd的key中提取任务名
// /cron/jobs/job10 ->job10
func ExtractJobName(jobKey string) string {
	//将JOB_DIR从string中删除
	return strings.Trim(jobKey, JOB_DIR)
}

//创建任务事件
//任务变化事件有2种：1）更新任务 2）删除任务
func BuildJobEvent(eventType int, job *Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

//构造任务执行计划
func BuildJobSchedulrPlan(job *Job) (jobSchedulePlan *JobSchedulerPlan, err error) {
	//cron 表达式的解析
	var (
		expr *cronexpr.Expression
	)
	//解析job的cron表达式，并检查是否合法
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	//生成任务调度计划对象
	jobSchedulePlan = &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

//构造执行状态信息
func BuildJobExecuteInfo(jobSchedulerPlan *JobSchedulerPlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobSchedulerPlan.Job,
		PlanTime: jobSchedulerPlan.NextTime, //计划调度时间
		RealTime: time.Now(),                //时间调度的时间
	}
	return
}
