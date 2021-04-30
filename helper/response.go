/*
  Copyright (c) 2021, Refactory.id
  All rights reserved.
*/
package helper

import "strings"

/**
 * Response is used for static shape json return
 */
type Response struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Errors  interface{} `json:"errors"`
	Data    interface{} `json:"data"`
	Embed   interface{} `json:"embed"`
}

/**
 * EmptyObj is used when data doesn't want to be null on json
 */
type EmptyObj struct{}

/**
 * Create new Response and returns Response's data
 * @param status boolean
 * @param message string
 * @param data interface{}
 */
func BuildResponse(status bool, message string, data interface{}) Response {
	res := Response{
		Status:  status,
		Message: message,
		Errors:  nil,
		Data:    data,
	}
	return res
}

/**
 * Create new Response and returns Response's data
 * @param status boolean
 * @param message string
 * @param data interface{}
 * @param embed interface{}
 */
func BuildResponseEmbed(status bool, message string, data interface{}, embed interface{}) Response {
	res := Response{
		Status:  status,
		Message: message,
		Errors:  nil,
		Data:    data,
		Embed:   embed,
	}
	return res
}

/**
 * Create new Response and returns Errors's data
 * @param status boolean
 * @param message string
 * @param data interface{}
 */
func BuildErrorResponse(message string, errs string, data interface{}) Response {
	splittedError := strings.Split(errs, "\n")
	res := Response{
		Status:  false,
		Message: message,
		Errors:  splittedError,
		Data:    data,
	}
	return res
}
