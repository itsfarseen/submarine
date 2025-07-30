package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"submarine/decoder"
	"submarine/rpc"
	"submarine/scale"
	"submarine/scale/gen/v10"
	"submarine/scale/gen/v11"
	"submarine/scale/gen/v12"
	"submarine/scale/gen/v9"
)

func main() {
	client, err := rpc.NewRPC("ws://37.27.51.25:9944")
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	log.Println("Connection established.")

	latestHeader, err := client.Send("chain_getHeader", nil).AsBlockHeader()
	if err != nil {
		log.Fatalf("Failed to get latest header: %v", err)
	}
	latestBlockNumber, _ := hexToDecimal(latestHeader.Number)
	log.Printf("Latest Block Number: %d", latestBlockNumber)

	v0, err := getMetadataVersion(client, 0)
	if err != nil {
		log.Fatalf("Failed to get metadata version for block 0: %v", err)
	}

	vLatest, err := getMetadataVersion(client, latestBlockNumber)
	if err != nil {
		log.Fatalf("Failed to get metadata version for block %d: %v", latestBlockNumber, err)
	}

	fmt.Printf("Metadata versions from %d to %d\n", v0, vLatest)

	versionBlocks := make(map[uint]uint64)

	for v := v0; v <= vLatest; v++ {
		log.Printf("Searching for block with metadata version %d", v)
		block, err := findBlockForVersion(client, v, 0, latestBlockNumber)
		if err != nil {
			log.Printf("Could not find block for version %d: %v", v, err)
			continue
		}
		versionBlocks[v] = block
	}

	fmt.Println("Found blocks for metadata versions:")
	for v, b := range versionBlocks {
		fmt.Printf("Version %d: Block %d\n", v, b)
	}

	for v := v0; v <= 13; v++ {
		block, ok := versionBlocks[v]
		if !ok {
			continue
		}
		types, err := getEventTypes(client, v, block)
		if err != nil {
			log.Printf("Could not get event types for version %d at block %d: %v", v, block, err)
			continue
		}
		fmt.Printf("Version %d Event Types: %v\n", v, types)
	}
}

func hexToDecimal(hex string) (uint64, error) {
	if len(hex) > 2 && hex[:2] == "0x" {
		hex = hex[2:]
	}
	return strconv.ParseUint(hex, 16, 64)
}

func getMetadataVersion(client *rpc.RPC, blockNumber uint64) (uint, error) {
	blockHash, err := client.Send("chain_getBlockHash", []any{blockNumber}).AsString()
	if err != nil {
		return 0, fmt.Errorf("failed to get block hash for block %d: %w", blockNumber, err)
	}
	if blockHash == "0x0000000000000000000000000000000000000000000000000000000000000000" {
		return 0, fmt.Errorf("block %d not found", blockNumber)
	}
	metadata := client.GetMetadata(blockHash)
	return metadata.Version, nil
}

func findBlockForVersion(client *rpc.RPC, version uint, start, end uint64) (uint64, error) {
	var foundBlock uint64
	low, high := start, end

	for low <= high {
		mid := low + (high-low)/2
		if mid > end {
			break
		}
		v, err := getMetadataVersion(client, mid)
		if err != nil {
			// This can happen if a block is skipped or not available, we can try to adjust.
			// For simplicity, we'll treat it as an error for now.
			return 0, err
		}

		if v < version {
			low = mid + 1
		} else if v >= version {
			foundBlock = mid
			high = mid - 1
		}
	}

	if foundBlock > 0 {
		// verify it is the correct version
		v, err := getMetadataVersion(client, foundBlock)
		if err != nil {
			return 0, err
		}
		if v == version {
			return foundBlock, nil
		}
	}

	return 0, fmt.Errorf("not found")
}

func getEventTypes(client *rpc.RPC, version uint, blockNumber uint64) ([]string, error) {
	blockHash, err := client.Send("chain_getBlockHash", []any{blockNumber}).AsString()
	if err != nil {
		return nil, fmt.Errorf("failed to get block hash for block %d: %w", blockNumber, err)
	}

	metadataRaw := client.GetMetadata(blockHash)
	metadataReader := scale.NewReader(metadataRaw.Data)
	decoded, err := decoder.DecodeMetadata(metadataRaw.Version, metadataReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode metadata: %w", err)
	}

	types := make(map[string]struct{})

	switch meta := decoded.(type) {
	case *v9.Metadata:
		for _, mod := range meta.Modules {
			if mod.Events != nil {
				for _, ev := range *mod.Events {
					for _, arg := range ev.Args {
						types[arg] = struct{}{}
					}
				}
			}
		}
	case *v10.Metadata:
		for _, mod := range meta.Modules {
			if mod.Events != nil {
				for _, ev := range *mod.Events {
					for _, arg := range ev.Args {
						types[arg] = struct{}{}
					}
				}
			}
		}
	case *v11.Metadata:
		for _, mod := range meta.Modules {
			if mod.Events != nil {
				for _, ev := range *mod.Events {
					for _, arg := range ev.Args {
						types[arg] = struct{}{}
					}
				}
			}
		}
	case *v12.Metadata:
		for _, mod := range meta.Modules {
			if mod.Events != nil {
				for _, ev := range *mod.Events {
					for _, arg := range ev.Args {
						types[arg] = struct{}{}
					}
				}
			}
		}
	default:
		return nil, fmt.Errorf("unsupported metadata type for event type extraction: %T", decoded)
	}

	var sortedTypes []string
	for t := range types {
		sortedTypes = append(sortedTypes, t)
	}
	sort.Strings(sortedTypes)

	return sortedTypes, nil
}
