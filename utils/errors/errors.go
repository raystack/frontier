package shieldError

import "errors"

var (
	Unauthorized = errors.New("you are not authorized")
	InvalidUUID  = errors.New("invalid syntax of uuid")
)
