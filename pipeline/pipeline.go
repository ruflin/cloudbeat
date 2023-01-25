// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package pipeline

import (
	"context"
	"github.com/elastic/cloudbeat/resources/fetching"
	"github.com/elastic/elastic-agent-autodiscover/kubernetes"
	"github.com/elastic/elastic-agent-libs/logp"
)

const (
	chBuffer = 10
)

func Step[In any, Out any](log *logp.Logger, inputChannel chan In, fn func(context.Context, In) (Out, error)) chan Out {
	outputCh := make(chan Out, chBuffer)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(outputCh)
		defer cancel()

		for s := range inputChannel {

			val, err := fn(ctx, s)
			if err != nil {
				log.Error(err)
				continue
			}
			outputCh <- val
		}
	}()

	return outputCh
}

func ExtractAssets[In any, Out any](log *logp.Logger, inputChannel chan fetching.ResourceInfo, fn func(context.Context, In) (Out, error)) (chan fetching.ResourceInfo, chan fetching.ResourceInfo) {
	outputCh := make(chan fetching.ResourceInfo, chBuffer)
	assetsCh := make(chan fetching.ResourceInfo, chBuffer)
	_, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(outputCh)
		defer cancel()

		for s := range inputChannel {
			// Check type of data to decide if it should be collected as asset
			data := s.GetData()
			_, ok := data.(*kubernetes.Node)
			if ok {
				// Create copy and send it to asset channel
				copyS := s
				assetsCh <- copyS
			}

			// Pipe through all data
			outputCh <- s
		}
	}()

	return outputCh, assetsCh
}
