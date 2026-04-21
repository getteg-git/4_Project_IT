package handlers

import (
	"fmt"
	"time"
)

func randomUsername() string {
	return fmt.Sprintf("user_%d", time.Now().UnixNano())
}
