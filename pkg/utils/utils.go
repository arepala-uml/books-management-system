package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/arepala-uml/books-management-system/pkg/config"
	"github.com/arepala-uml/books-management-system/pkg/models"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/gommon/log"
)

func FormatErrorMessage(err validator.FieldError) string {
	// Customize the error messages to make them user-friendly
	log.Info(err)
	switch err.Tag() {
	case "required":
		return err.Field() + " is required"
	case "string":
		return err.Field() + " must be a string"
	case "gte":
		return err.Field() + " must be greater than or equal to " + err.Param()
	case "lte":
		return err.Field() + " must be less than or equal to " + err.Param()
	case "max":
		return err.Field() + " cannot exceed " + err.Param() + " characters"
	case "min":
		return err.Field() + " must be at least " + err.Param() + " characters"
	case "numeric":
		return err.Field() + " must be a valid number" // Handle numeric validation error
	default:
		return err.Field() + " is invalid"
	}
}

func RedisKeyExists(key string) bool {
	rInt, err := config.RedisClient.Exists(context.Background(), key).Result()
	if err != nil {
		log.Error(err)
		return false
	}
	return rInt == 1
}

func ReJSONSet(key string, path string, data models.Book, expiry int) error {
	var message string
	if RedisKeyExists(key) {
		message = fmt.Sprintf("updating data for %s", key)
		if expiry > 0 {
			message = fmt.Sprintf("updating data for %s with expiry of %d seconds", key, expiry)
		}
	} else {
		message = fmt.Sprintf("JSONSet for %s", key)
		if expiry > 0 {
			message = fmt.Sprintf("JSONSet for %s with expiry of %d seconds", key, expiry)
		}
	}
	res, err := config.ReJSONHandler.JSONSet(key, path, data)
	if err != nil {
		log.Fatalf("failed to JSONSet for %s - err %v", key, err)
		return err
	}
	if res.(string) == "OK" {
		if expiry > 0 {
			config.RedisClient.Expire(context.Background(), key, time.Second*time.Duration(expiry))
		}
		log.Info(message)
	} else {
		msg := fmt.Sprintf("failed to JSONSet for %v - res - %v", key, res)
		return errors.New(msg)
	}
	return nil
}

func ReJSONGet(key string, path string) (interface{}, error) {
	data, err := config.ReJSONHandler.JSONGet(key, path)
	if err != nil {
		errMessage := fmt.Sprintf("failed to JSONGet for %v, %v", key, err)
		return "", errors.New(errMessage)
	}
	var bookData interface{}
	err = json.Unmarshal(data.([]byte), &bookData)
	if err != nil {
		errMessage := fmt.Sprintf("failed to unMarshal JSON data for %v, %v", key, err)
		return "", errors.New(errMessage)
	}
	return bookData, nil
}

func ReJSONDel(key string, path string) error {
	res, err := config.ReJSONHandler.JSONDel(key, path)
	if err != nil {
		errMessage := fmt.Sprintf("failed to JSONDel for %v, %v", key, err)
		return errors.New(errMessage)
	}
	if res.(int64) == 1 {
		log.Infof("JSONDel for %s", key)
	} else if res.(int64) == 0 {
		log.Infof("key - %s not found in Redis", key)
	} else {
		log.Warnf("failed to JSONSet for %v - res - %v", key, res)
	}
	return nil
}
