package utils

import (
	"fmt"
	"time"
)

// GenerateID 生成唯一ID
func GenerateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
