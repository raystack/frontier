package shieldError

import "errors"

var (
	Unauthorzied = errors.New("you are not authorized")
	InvalidUUID  = errors.New("invalid syntax of uuid")
)
