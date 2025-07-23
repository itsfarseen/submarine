package rpc

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"
)

const SYSTEM_EVENT_KEY = "0x26aa394eea5630e07c48ae0c9558cef780d41e5e16056765bc8461851072c9d7"

type ChainMetadata struct {
	Version uint
	Data    []byte
}

func (client *RPC) GetMetadata(blockHash string) ChainMetadata {
	metadataReq := client.Send("state_getMetadata", []any{blockHash})

	metadataHex, err := metadataReq.AsString()
	if err != nil {
		log.Fatalf("Failed to get metadata: %v", err)
	}

	// Remove the '0x' prefix and decode the hex string into bytes.
	cleanHex := strings.TrimPrefix(metadataHex, "0x")
	metadataBytes, err := hex.DecodeString(cleanHex)
	if err != nil {
		log.Fatalf("Failed to decode metadata hex: %v", err)
	}

	var version uint = 0

	// The version is the 5th byte (index 4) after the 4-byte magic number ('meta').
	if len(metadataBytes) < 5 {
		log.Fatalf("Metadata is too short to contain a version number.")
	} else {
		version = uint(metadataBytes[4])
	}

	data := metadataBytes[5:]

	return ChainMetadata{
		Version: version,
		Data:    data,
	}
}

func (client *RPC) GetEvents(blockHash string) []byte {
	resp := client.Send("state_getStorage", []any{SYSTEM_EVENT_KEY, blockHash})
	eventsHex, err := resp.AsString()
	if err != nil {
		log.Fatalf("Failed to get events: %s", err)
	}
	fmt.Printf("Events length: %d\n", len(eventsHex))
	eventsBytes, err := hex.DecodeString(eventsHex[2:])
	if err != nil {
		log.Fatalf("decode events hex: %s", err)
	}
	return eventsBytes
}
