package confutil
 

type Options struct {
	Addr               string
	Password           string
	DB                 int
	DialTimeout        Duration
	ReadTimeout        Duration
	WriteTimeout       Duration
	PoolSize           int
	MinIdleConns       int
	MaxConnAge         Duration
	PoolTimeout        Duration
	IdleTimeout        Duration
	IdleCheckFrequency Duration
}
