// Copyright 2024 Universidad Carlos III de Madrid
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
)

// ContainsString checks if a string is present in a slice.
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// RemoveString removes a string from a slice.
func RemoveString(slice []string, s string) []string {
	var result []string
	for _, item := range slice {
		if item != s {
			result = append(result, item)
		}
	}
	return result
}

func Int32Ptr(i int32) *int32 { return &i }

func GenerateHash(obj runtime.Object) string {
	// Serializer that handles runtime.Objects specifically for Kubernetes
	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{Yaml: false, Pretty: false, Strict: true})

	// Create a buffer to hold the JSON data
	var b bytes.Buffer

	// Encode the object to JSON; handle runtime objects appropriately
	err := s.Encode(obj, &b)
	if err != nil {
		return ""
	}

	// Compute the SHA-1 hash of the JSON representation
	hash := sha1.Sum(b.Bytes())
	return hex.EncodeToString(hash[:5])
}

func SpecToJson(obj runtime.Object) bytes.Buffer {
	s := json.NewSerializerWithOptions(json.DefaultMetaFactory, nil, nil, json.SerializerOptions{Yaml: false, Pretty: false, Strict: true})

	// Create a buffer to hold the JSON data
	var b bytes.Buffer

	// Encode the object to JSON; handle runtime objects appropriately
	s.Encode(obj, &b)

	return b
}

// GetPortNumberFromNetAttachDef extracts the port number from the network attachment name.
func GetPortNumberFromNetAttachDef(netAttachName string) (string, error) {
	const keyword = "veth"

	// Check if the keyword exists in the netAttachName
	index := strings.Index(netAttachName, keyword)
	if index == -1 {
		return "", fmt.Errorf("keyword '%s' not found in network attachment name", keyword)
	}

	// Extract the port number after the keyword
	portNumber := netAttachName[index+len(keyword):]
	if portNumber == "" {
		return "", fmt.Errorf("port number not found after keyword '%s'", keyword)
	}

	return portNumber, nil
}

// generateDatapathID generates a datapath ID from the switch name
func GenerateDatapathID(switchName string) string {
	// Create a new SHA256 hash object
	hash := sha256.New()

	// Write the switch name to the hash object
	hash.Write([]byte(switchName))

	// Get the hashed bytes
	hashedBytes := hash.Sum(nil)

	// Take the first 8 bytes of the hash to create a 64-bit ID
	dpidBytes := hashedBytes[:8]

	// Convert the bytes to a hexadecimal string
	dpid := hex.EncodeToString(dpidBytes)

	return dpid
}

type BridgeParams struct {
	NodeName     string
	ProviderName string
}

func GetBridgeName(params BridgeParams) string {
	hash := sha256.New()
	hash.Write([]byte(fmt.Sprintf("%s%s", params.NodeName, params.ProviderName)))
	hashedBytes := hash.Sum(nil)
	dpidBytes := hashedBytes[:4]

	// Convert the bytes to a hexadecimal string
	dpid := hex.EncodeToString(dpidBytes)
	return fmt.Sprintf("br-%s", dpid)
}

func GenerateServiceName(overlayName, nodeName string) string {
	hash := fnv.New32() // Using FNV hash for a compact hash, but still 32 bits
	hash.Write([]byte(fmt.Sprintf("%s%s", overlayName, nodeName)))
	sum := hash.Sum32()
	// Encode the hash as a base32 string and take the first 4 characters
	return fmt.Sprintf("l2sm-switch-%04x", sum) // H
}
