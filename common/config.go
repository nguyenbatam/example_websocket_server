package common

type Redis struct {
	Addr     string
	Password string
}

type Config struct {
	Redis        Redis
	JwtSecretKey string
	NumberWorker int
}
