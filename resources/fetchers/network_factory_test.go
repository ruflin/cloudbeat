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

package fetchers

import (
	"github.com/elastic/cloudbeat/resources/providers/awslib/ec2"
	"testing"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/elastic/cloudbeat/resources/providers/awslib"
	agentconfig "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var awsConfig = `
name: aws-iam
access_key_id: key
secret_access_key: secret
session_token: session
default_region: us1-east
`

func TestNetworkFactory_Create(t *testing.T) {
	identity := &awslib.MockIdentityProviderGetter{}
	identity.EXPECT().GetIdentity(mock.Anything).Return(&awslib.Identity{
		Account: awssdk.String("test-account"),
	}, nil)

	mockCrossRegion := &awslib.MockCrossRegionFactory[ec2.ElasticCompute]{}
	mockCrossRegion.On(
		"NewMultiRegionClients",
		mock.Anything,
		mock.Anything,
		mock.Anything,
		mock.Anything,
	).Return(&awslib.MockCrossRegionFetcher[ec2.ElasticCompute]{})

	f := &EC2NetworkFactory{
		CrossRegionFactory: mockCrossRegion,
		IdentityProvider: func(cfg awssdk.Config) awslib.IdentityProviderGetter {
			return identity
		},
	}
	cfg, err := agentconfig.NewConfigFrom(awsConfig)
	assert.NoError(t, err)
	fetcher, err := f.Create(logp.NewLogger("network-factory-test"), cfg, nil)
	assert.NoError(t, err)
	assert.NotNil(t, fetcher)
	nacl, ok := fetcher.(*NetworkFetcher)
	assert.True(t, ok)
	assert.Equal(t, nacl.cfg.AwsConfig.AccessKeyID, "key")
	assert.Equal(t, nacl.cfg.AwsConfig.SecretAccessKey, "secret")
}
