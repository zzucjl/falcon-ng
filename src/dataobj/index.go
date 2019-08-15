package dataobj

type IndexModel struct {
	Endpoint  string            `json:"endpoint"`
	Metric    string            `json:"metric"`
	DsType    string            `json:"dsType"`
	Step      int               `json:"step"`
	Tags      map[string]string `json:"tags"`
	Timestamp int64             `json:"ts"`
}
