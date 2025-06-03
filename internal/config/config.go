package config

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	DbUrl           string `json:"db_url"`
	CurrentUserName string `json:"current_user_name"`
}

func Read() (Config, error) {
	// Reads the JSON file at ~/.gatorconfig.json
	// decode the JSON string into a Config struct.
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, errors.New("Couldn't get Home directory.")
	}
	config_file := home + "/.gatorconfig.json"

	jsonData, err := os.ReadFile(config_file)
	if err != nil {
		return Config{}, errors.New("Error reading .gatorconfig.json.")
	}

	var config Config
	err = json.Unmarshal(jsonData, &config)
	if err != nil {
		return Config{}, errors.New("Error Unmarshaling JSON file.")
	}
	return config, nil
}

func (cfg *Config) SetUser(username string) error {
	// Set the current_user_name field and call the write() helper function.
	cfg.CurrentUserName = username
	err := write(cfg)
	if err != nil {
		return errors.New("Error writing to ~/.gatorconfig.json")
	}
	return nil
}

func write(cfg *Config) error {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	config_file := home + "/.gatorconfig.json"

	err = os.WriteFile(config_file, jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}
