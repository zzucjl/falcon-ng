package schema

import "regexp"

/*
Name		- 监控策略名
FilePath	- 文件路径
TimeFormat	- 时间格式
Pattern		- 表达式
Interval	- 采集周期
Tags		- Tags
Func		- 采集方式（max/min/avg/cnt）
Degree		- 精度位数
MeasurementType	- Web保留字段，固定为LOG
*CollectType	- 采集方式（FLOW / COSTTIME）
*Aggregate	- 聚合方式
*Unit		- 单位
*Commit		- 备注
*/

type Strategy struct {
	ID              int64                     `json:"id"`
	Name            string                    `json:"name"`
	FilePath        string                    `json:"file_path"`
	TimeFormat      string                    `json:"time_format"`
	Pattern         string                    `json:"pattern"`
	Exclude         string                    `json:"-"`
	MeasurementType string                    `json:"measurement_type"`
	Interval        int64                     `json:"interval"`
	Tags            map[string]string         `json:"tags"`
	Func            string                    `json:"func"`
	Degree          int64                     `json:"degree"`
	Unit            string                    `json:"unit"`
	Comment         string                    `json:"comment"`
	Creator         string                    `json:"creator"`
	SrvUpdated      string                    `json:"updated"`
	LocalUpdated    int64                     `json:"-"`
	TimeReg         *regexp.Regexp            `json:"-"`
	PatternReg      *regexp.Regexp            `json:"-"`
	ExcludeReg      *regexp.Regexp            `json:"-"`
	TagRegs         map[string]*regexp.Regexp `json:"-"`
	ParseSucc       bool                      `json:"parse_succ"`
}

type LimitResp struct {
	CpuNum int `json:"cpu_num"`
	MemMB  int `json:"mem_mb"`
}

func DeepCopyStrategy(p *Strategy) *Strategy {
	s := Strategy{}
	s.ID = p.ID
	s.Name = p.Name
	s.FilePath = p.FilePath
	s.TimeFormat = p.TimeFormat
	s.Pattern = p.Pattern
	s.MeasurementType = p.MeasurementType
	s.Interval = p.Interval
	s.Tags = DeepCopyStringMap(p.Tags)
	s.Func = p.Func
	s.Degree = p.Degree
	s.Unit = p.Unit
	s.Comment = p.Comment
	s.Creator = p.Creator
	s.SrvUpdated = p.SrvUpdated
	s.LocalUpdated = p.LocalUpdated

	return &s
}

func DeepCopyStringMap(p map[string]string) map[string]string {
	r := make(map[string]string, len(p))
	for k, v := range p {
		r[k] = v
	}
	return r
}
