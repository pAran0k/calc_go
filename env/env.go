package env

import (
	"os"
	"strconv"
)

type Config struct {
	ComputingPower       int
	TimeAdditionMS       int
	TimeSubtractionMS    int
	TimeMultiplicationMS int
	TimeDivisionMS       int
	OrchestratorAddr     string
}

func LoadConfig() Config {
	return Config{
		ComputingPower:       getEnvInt("COMPUTING_POWER", 1),
		TimeAdditionMS:       getEnvInt("TIME_ADDITION_MS", 100),
		TimeSubtractionMS:    getEnvInt("TIME_SUBTRACTION_MS", 100),
		TimeMultiplicationMS: getEnvInt("TIME_MULTIPLICATIONS_MS", 100),
		TimeDivisionMS:       getEnvInt("TIME_DIVISIONS_MS", 100),
		OrchestratorAddr:     getEnvString("ORCHESTRATOR_ADDR", ":8080"),
	}
}

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
