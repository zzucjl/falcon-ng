package query

var (
	defaultSeriesMaxConn     = 10
	defaultSeriesMaxIdle     = 10
	defaultSeriesConnTimeout = 1000
	defaultSeriesCallTimeout = 2000
)

type SeriesQueryOption struct {
	Addrs            []string `yaml:"addrs"`            // 直连下游的地址
	MaxConn          int      `yaml:"maxConn"`          //
	MaxIdle          int      `yaml:"maxIdle"`          //
	ConnTimeout      int      `yaml:"connTimeout"`      // 连接超时
	CallTimeout      int      `yaml:"callTimeout"`      // 请求超时
	IndexAddrs       []string `yaml:"indexAddrs"`       // 直连下游index的地址
	IndexCallTimeout int      `yaml:"indexCallTimeout"` // 请求超时
}

func NewSeriesQueryOption(query []string, index []string) SeriesQueryOption {
	return SeriesQueryOption{
		Addrs:            query,
		MaxConn:          defaultSeriesMaxConn,
		MaxIdle:          defaultSeriesMaxIdle,
		ConnTimeout:      defaultSeriesConnTimeout,
		CallTimeout:      defaultSeriesCallTimeout,
		IndexAddrs:       index,
		IndexCallTimeout: defaultSeriesCallTimeout,
	}
}
