// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

// Copyright 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copied from https://raw.githubusercontent.com/gardener/gardener/ae5cb871291bb46e5113c5a738c7458fe6141a81/pkg/healthz/informer.go

package healthz

import (
	"context"
	"errors"
	"net/http"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

type cacheSyncWaiter interface {
	WaitForCacheSync(ctx context.Context) bool
}

// NewCacheSyncHealthz returns a new healthz.Checker that will pass only if all informers in the given cacheSyncWaiter sync.
func NewCacheSyncHealthz(cacheSyncWaiter cacheSyncWaiter) healthz.Checker {
	return func(_ *http.Request) error {
		// cache.Cache.WaitForCacheSync is racy for closed context, so use context with 1ms timeout instead.
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		if !cacheSyncWaiter.WaitForCacheSync(ctx) {
			return errors.New("informers not synced yet")
		}
		return nil
	}
}
