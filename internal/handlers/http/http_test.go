package http_handlers

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func ptrUUID(u uuid.UUID) *uuid.UUID { return &u }
func ptrTime(t time.Time) *time.Time { return &t }
func ptrInt(i int) *int              { return &i }
func marshaled(t *testing.T, v interface{}) *bytes.Buffer {
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}
