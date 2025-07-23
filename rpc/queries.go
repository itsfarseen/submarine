package rpc

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"submarine/decoder"
	"submarine/scale"
)

func (client *RPC) GetMetadata(blockHash string) scale.MetadataV14 {
	log.Println("Querying runtime metadata...")

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

	// The version is the 5th byte (index 4) after the 4-byte magic number ('meta').
	if len(metadataBytes) < 5 {
		log.Println("Metadata is too short to contain a version number.")
	} else {
		version := metadataBytes[4]
		log.Printf("✅ Successfully retrieved metadata (V%d).", version)
	}

	err = os.WriteFile("metadata.dump", metadataBytes, 0644)
	if err != nil {
		panic(err)
	}

	reader := scale.NewReader(metadataBytes[5:])
	meta, err := scale.DecodeMetadataV14(reader)
	if err != nil {
		log.Printf("pos: %d", reader.Pos())
		log.Fatal(err)
	}

	json, err := json.MarshalIndent(meta, "", "   ")
	err = os.WriteFile("metadata.json", json, 0644)
	if err != nil {
		panic(err)
	}

	return meta
}

func (client *RPC) GetEvents(blockHash string) []byte {
	resp := client.Send("state_getStorage", []any{decoder.SYSTEM_EVENT_KEY, blockHash})
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
