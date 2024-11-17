package config

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "github.com/spf13/viper"
)

const (
    AppName = "disc-cuer"
    AppVersion = "0.2"
)

var (
    GnuHelloEmail string
    GnuDbUrl      string
    CacheLocation string
)

func init() () {
    // Set default values
    viper.SetDefault("gnuHelloEmail", "")
    viper.SetDefault("gnuDbUrl", "https://gnudb.gnudb.org")
    viper.SetDefault("cacheLocation", getDefaultCacheFolder())

    // Load from configuration file, environment variables, and CLI flags
    viper.SetConfigName("config")  // name of config file (without extension)
    viper.SetConfigType("yaml")    // config file format
    viper.AddConfigPath(filepath.Join("/etc", AppName))  // Global configuration path
    if home, err := os.UserHomeDir(); err == nil {
        viper.AddConfigPath(filepath.Join(home, ".config", AppName)) // User config path
    }

    // Environment variable support
    viper.SetEnvPrefix(strings.ReplaceAll(AppName, "-", "_")) // environment variables start with CD_CUER
    viper.AutomaticEnv()

    // Load configuration from file if present
    err := viper.ReadInConfig()
    if err != nil {
        // File not found is acceptable, only raise errors for other issues
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
            os.Exit(1)
        }
    }

    if GnuHelloEmail = viper.GetString("gnuHelloEmail"); GnuHelloEmail == "" {
        fmt.Fprintf(os.Stderr, "gnuHelloEmail is required in config.yaml or via environment variable to use gnuDB\n")
    }
    CacheLocation = viper.GetString("cacheLocation")
    GnuDbUrl = viper.GetString("gnuDbUrl")
}

func getDefaultCacheFolder() string {
    home, err := os.UserHomeDir()
    if err != nil {
        return filepath.Join("var", "cache", AppName)
    }
    return filepath.Join(home, ".cache", AppName)
}
