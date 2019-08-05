#### 算子列举

```
type StrategyExpression struct {
	Func       string              `json:"func"`       // 算子, 目前有10个
	Params     []string            `json:"params"`     // 不同算子的参数不同, 按照约定格式填充并解析, params改成[]string
	Operator   string              `json:"operator"`   // 多个阈值的逻辑运算, 支持 与、或
	Thresholds []StrategyThreshold `json:"thresholds"` // 阈值
}
```

|算子|英文名称|参数(加粗)|备注|
|:--|:--|:--|:--|
|持续聚合|duration\_stat| __X__ 秒, 聚合方式__P__|max/min/sun/avg等|
|持续发生|duration\_happen| __X__ 秒, __N__ 次||
|无数据|nodata| __X__ 秒||