package config

// Config The general configuration for the file converter application.
type Config struct {
	// Port specifies the network port on which the application will listen for incoming connections. Defaults to `8080`.
	Port int

	// Address specifies the network address on which the application will listen for incoming connections. Defaults to `0.0.0.0`.
	Address string

	// MaxFileSizeBytes specifies the maximum allowed file size in bytes. Defaults to `25 * 1024 * 1024` (25 MB).
	MaxFileSizeBytes int64
}
