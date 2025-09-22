package keyservice

// Config holds all necessary configuration for the key service.
type Config struct {
	HTTPListenAddr string
	// ADDED: The secret key for validating JWTs. It will be loaded
	// from the "JWT_SECRET" environment variable.
	JWTSecret string `env:"JWT_SECRET,required"`
}
