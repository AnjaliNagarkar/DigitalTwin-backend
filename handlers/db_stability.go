package handlers

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func isInvalidDBConnection(err error) bool {
	if err == nil {
		return false
	}

	message := strings.ToLower(err.Error())
	return errors.Is(err, sql.ErrConnDone) ||
		strings.Contains(message, "invalid connection") ||
		strings.Contains(message, "wsarecv") ||
		strings.Contains(message, "connection forcibly closed")
}

func ensureDBReady(c *gin.Context, db *sql.DB, endpoint string) (context.Context, context.CancelFunc, bool) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)

	if err := db.PingContext(ctx); err == nil {
		return ctx, cancel, true
	} else {
		log.Printf("[DB] ping failed before %s: %v", endpoint, err)
		if retryErr := db.PingContext(ctx); retryErr == nil {
			return ctx, cancel, true
		} else {
			log.Printf("[DB] ping retry failed before %s: %v", endpoint, retryErr)
		}
	}

	cancel()
	return nil, nil, false
}
