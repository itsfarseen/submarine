package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"strings"
	"submarine/decoder"
	decoder_v13 "submarine/decoder/v13"
	decoder_v14 "submarine/decoder/v14"
	. "submarine/rpc"
	"submarine/scale"
	scale_v13 "submarine/scale/v13"
	scale_v14 "submarine/scale/v14"
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

	var exts []decoder.DecodedExtrinsic
	var events []decoder.EventRecord

	metadataReader := scale.NewReader(metadataRaw.Data)

	switch metadataRaw.Version {
	case 14:
		metadata, err := scale_v14.DecodeMetadata(metadataReader)
		if err != nil {
			log.Fatalf("Failed to decode v14 metadata: %s", err)
		}

		for _, ext := range signedBlock.Block.Extrinsics {
			ext = strings.TrimPrefix(ext, "0x")
			extBytes, err := hex.DecodeString(ext)
			if err != nil {
				log.Fatal(err)
			}
			ext_, err := decoder_v14.DecodeExtrinsic(&metadata, extBytes)
			if err != nil {
				log.Fatal(err)
			}
			exts = append(exts, *ext_)
			fmt.Printf("extrinsic: %s: %s\n", ext_.Call.PalletName, ext_.Call.VariantName)
		}

		eventsBytes := client.GetEvents(blockHash)
		events, err = decoder_v14.DecodeEvents(&metadata, eventsBytes)
		if err != nil {
			log.Fatalf("failed to decode v14 event: %s", err)
		}

	case 13:
		metadata, err := scale_v13.DecodeMetadata(metadataReader)
		if err != nil {
			log.Fatalf("Failed to decode v13 metadata: %s", err)
		}

		for _, ext := range signedBlock.Block.Extrinsics {
			ext = strings.TrimPrefix(ext, "0x")
			extBytes, err := hex.DecodeString(ext)
			if err != nil {
				log.Fatal(err)
			}
			ext_, err := decoder_v13.DecodeExtrinsic(&metadata, extBytes)
			if err != nil {
				log.Fatal(err)
			}
			exts = append(exts, *ext_)
			fmt.Printf("extrinsic: %s: %s\n", ext_.Call.PalletName, ext_.Call.VariantName)
		}

		eventsBytes := client.GetEvents(blockHash)
		events, err = decoder_v13.DecodeEvents(&metadata, eventsBytes)
		if err != nil {
			log.Fatalf("failed to decode v13 event: %s", err)
		}
	default:
		log.Fatalf("Unsupported metadata version: %d", metadataRaw.Version)
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
