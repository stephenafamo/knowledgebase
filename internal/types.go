package internal

type Config struct {
	TESTING bool `env:"TESTING"`

	PORT     int    `env:"PORT,required"`
	ROOT_DIR string `env:"ROOT_DIR,required"`

	SENTRY_DSN string `env:"SENTRY_DSN"`
}
