package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"submarine/decoder"
	decoder_models "submarine/decoder/models"
	. "submarine/rpc"
	"submarine/scale"
)

func main() {
	client, err := NewRPC("ws://37.27.51.25:9944")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	log.Println("Connection established. Querying system info...")

	// Chain info

	chainReq := client.Send("system_chain", nil)
	nameReq := client.Send("system_name", nil)
	versionReq := client.Send("system_version", nil)

	chainName, err := chainReq.AsString()
	if err != nil {
		log.Fatalf("Failed to get chain name: %v", err)
	}
	log.Printf("✅ Chain Name: %s", chainName)

	nodeName, err := nameReq.AsString()
	if err != nil {
		log.Fatalf("Failed to get node name: %v", err)
	}
	log.Printf("✅ Node Name: %s", nodeName)

	nodeVersion, err := versionReq.AsString()
	if err != nil {
		log.Fatalf("Failed to get node version: %v", err)
	}
	log.Printf("✅ Node Version: %s\n", nodeVersion)

	// Latest blocks info

	log.Println("Querying block numbers...")
	latestHeaderReq := client.Send("chain_getHeader", nil)
	finalizedHashReq := client.Send("chain_getFinalizedHead", nil)

	latestHeader, err := latestHeaderReq.AsBlockHeader()
	if err != nil {
		log.Fatalf("Failed to get latest header: %v", err)
	}
	latestBlockNumber, _ := hexToDecimal(latestHeader.Number)
	log.Printf("✅ Latest Block Number: %d", latestBlockNumber)

	finalizedHash, err := finalizedHashReq.AsString()
	if err != nil {
		log.Fatalf("Failed to get finalized hash: %v", err)
	}

	finalizedHeaderReq := client.Send("chain_getHeader", []any{finalizedHash})
	finalizedHeader, err := finalizedHeaderReq.AsBlockHeader()
	if err != nil {
		log.Fatalf("Failed to get finalized header: %v", err)
	}
	finalizedBlockNumber, _ := hexToDecimal(finalizedHeader.Number)
	log.Printf("✅ Finalized Block Number: %d\n", finalizedBlockNumber)

	// Block contents

	blockNumberToQuery := uint64(7_000_000)
	log.Printf("Querying block #%d...", blockNumberToQuery)

	blockHashReq := client.Send("chain_getBlockHash", []any{blockNumberToQuery})
	blockHash, err := blockHashReq.AsString()
	if err != nil {
		log.Fatalf("Failed to get block hash: %v", err)
	}
	log.Printf("✅ Hash for block #%d: %s", blockNumberToQuery, blockHash)

	signedBlockReq := client.Send("chain_getBlock", []any{blockHash})
	signedBlock, err := signedBlockReq.AsSignedBlock()
	if err != nil {
		log.Fatalf("Failed to get signed block: %v", err)
	}
	log.Printf("✅ Block data for hash %s:", blockHash)

	metadataRaw := client.GetMetadata(blockHash)
	log.Printf("Metadata Version: %d", metadataRaw.Version)

	metadataReader := scale.NewReader(metadataRaw.Data)
	metadata, err := decoder.DecodeMetadata(metadataRaw.Version, metadataReader)
	if err != nil {
		log.Fatalf("Failed to decode metadata: %s", err)
	}

	exts := make([]decoder_models.DecodedExtrinsic, 0, 10)

	for _, ext := range signedBlock.Block.Extrinsics {
		ext = strings.TrimPrefix(ext, "0x")
		extBytes, err := hex.DecodeString(ext)
		if err != nil {
			log.Fatal(err)
		}
		ext_, err := decoder.DecodeExtrinsic(metadata, extBytes)
		if err != nil {
			log.Fatal(err)
		}
		exts = append(exts, *ext_)
		fmt.Printf("extrinsic: %s: %s\n", ext_.Call.PalletName, ext_.Call.VariantName)
	}

	eventsBytes := client.GetEvents(blockHash)

	events, err := decoder.DecodeEvents(metadata, eventsBytes)
	if err != nil {
		log.Fatalf("failed to decode event: %s", err)
	}

	for _, event := range events {
		var ctx string
		if event.Phase.IsApplyExtrinsic {
			ext := exts[event.Phase.AsApplyExtrinsic]
			ctx = fmt.Sprintf("(ext) %s: %s", ext.Call.PalletName, ext.Call.VariantName)
		} else if event.Phase.IsInitialization {
			ctx = "init"
		} else if event.Phase.IsFinalization {
			ctx = "fin"
		} else {
			ctx = "unknown"
		}
		fmt.Printf("event (%s): %s: %s\n", ctx, event.Event.PalletName, event.Event.EventName)
	}
}

// hexToDecimal converts a hex string (e.g., "0x123") to a decimal number.
func hexToDecimal(hex string) (uint64, error) {
	if len(hex) > 2 && hex[:2] == "0x" {
		hex = hex[2:]
	}
	return strconv.ParseUint(hex, 16, 64)
}
