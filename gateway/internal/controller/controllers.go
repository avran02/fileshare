package controller

import (
	jsoniter "github.com/json-iterator/go"
)

type Controllers struct {
	UsersController UsersController
	FilesController FilesController
	ShareController ShareController
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary
