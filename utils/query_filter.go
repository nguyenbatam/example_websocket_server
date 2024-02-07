package utils

import (
	"github.com/gin-gonic/gin"
	"strconv"
)

func QueryInt64Context(c *gin.Context, query string, _default int64) (int64, error) {
	text := c.Query(query)
	value, err := strconv.ParseInt(text, 10, 64)
	if err != nil {
		return _default, err
	}
	return value, nil
}
func QueryIntContext(c *gin.Context, query string, _default int) (int, error) {
	text := c.Query(query)
	value, err := strconv.Atoi(text)
	if err != nil {
		return _default, err
	}
	return value, nil
}

func ParamUint64Context(c *gin.Context, query string, _default uint64) (uint64, error) {
	text := c.Param(query)
	value, err := strconv.ParseUint(text, 10, 64)
	if err != nil {
		return _default, err
	}
	return value, nil
}
