// Copyright (C) 2026 Jad Madi. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package main

import (
	"context"
	"errors"
	"testing"
)

func TestIsNormalLiveStop(t *testing.T) {
	for _, err := range []error{
		nil,
		context.Canceled,
		errors.New("program was killed: program was interrupted"),
		errors.New("program was interrupted"),
	} {
		if !isNormalLiveStop(err) {
			t.Fatalf("%v should be a normal stop", err)
		}
	}
	if isNormalLiveStop(errors.New("renderer failed")) {
		t.Fatal("a renderer failure must not be a normal stop")
	}
}
