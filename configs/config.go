package configs

import "github.com/joho/godotenv"

func LoadConfig() error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	return nil
}
