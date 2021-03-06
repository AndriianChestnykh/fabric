/*
Copyright IBM Corp. 2016 All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

                 http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package static

import (
	"math/rand"

	ab "github.com/hyperledger/fabric/orderer/atomicbroadcast"
	"github.com/hyperledger/fabric/orderer/common/bootstrap"
	"github.com/hyperledger/fabric/orderer/common/cauthdsl"
	"github.com/hyperledger/fabric/orderer/common/configtx"

	"github.com/golang/protobuf/proto"
)

type bootstrapper struct {
	chainID []byte
}

// New returns a new static bootstrap helper
func New() bootstrap.Helper {
	b := make([]byte, 16)
	rand.Read(b)

	return &bootstrapper{
		chainID: b,
	}
}

// errorlessMarshal prevents poluting this code with many panics, if the genesis block cannot be created, the system cannot start so panic is correct
func errorlessMarshal(thing proto.Message) []byte {
	data, err := proto.Marshal(thing)
	if err != nil {
		panic(err)
	}
	return data
}

func (b *bootstrapper) makeConfigurationEntry(id string, ctype ab.Configuration_ConfigurationType, data []byte, modificationPolicyID string) *ab.ConfigurationEntry {
	configurationBytes := errorlessMarshal(&ab.Configuration{
		ChainID:            b.chainID,
		ID:                 id,
		LastModified:       0,
		Type:               ctype,
		Data:               data,
		ModificationPolicy: modificationPolicyID,
	})
	return &ab.ConfigurationEntry{
		Configuration: configurationBytes,
	}
}

func sigPolicyToPolicy(sigPolicy *ab.SignaturePolicyEnvelope) []byte {
	policy := &ab.Policy{
		Type: &ab.Policy_SignaturePolicy{
			SignaturePolicy: sigPolicy,
		},
	}
	return errorlessMarshal(policy)
}

// GenesisBlock returns the genesis block to be used for bootstrapping
func (b *bootstrapper) GenesisBlock() (*ab.Block, error) {

	// Lock down the default modification policy to prevent any further policy modifications
	lockdownDefaultModificationPolicy := b.makeConfigurationEntry(configtx.DefaultModificationPolicyID, ab.Configuration_Policy, sigPolicyToPolicy(cauthdsl.RejectAllPolicy), configtx.DefaultModificationPolicyID)

	initialConfigTX := errorlessMarshal(&ab.ConfigurationEnvelope{
		Sequence: 0,
		ChainID:  b.chainID,
		Entries: []*ab.ConfigurationEntry{
			lockdownDefaultModificationPolicy,
		},
	})

	return &ab.Block{
		Number:   0,
		PrevHash: []byte("GENESIS"),
		Messages: []*ab.BroadcastMessage{
			&ab.BroadcastMessage{Data: initialConfigTX},
		},
	}, nil

}
