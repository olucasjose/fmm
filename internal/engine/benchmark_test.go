// Copyright (C) 2026 olucasjose
// SPDX-License-Identifier: GPL-3.0-or-later

package engine

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMeasureSpeedIncludesTimeToFirstByte(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write(make([]byte, 1024))
	}))
	defer server.Close()

	speed, err := measureSpeed(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("measureSpeed retornou erro: %v", err)
	}
	if speed >= 20*1024 {
		t.Errorf("velocidade %.2f B/s não parece incluir a espera pelos cabeçalhos", speed)
	}
}

func TestMeasureSpeedRejectsPartialTransfer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "1024")
		_, _ = fmt.Fprint(w, "partial")
	}))
	defer server.Close()

	if _, err := measureSpeed(context.Background(), server.URL); err == nil {
		t.Fatal("esperava erro para transferência parcial")
	}
}
