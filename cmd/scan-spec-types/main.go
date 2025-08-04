package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"submarine/decoder"
	"submarine/decoder/legacy"
	"submarine/rpc"
	"submarine/scale"
)

type BlockInfo struct {
	BlockNumber uint64 `json:"block"`
	BlockHash   string `json:"hash"`
	SpecVersion uint64 `json:"specVersion"`
}

const WS_URL = "ws://37.27.51.25:9944"
const BATCH_SIZE = 5

func main() {
	client, err := rpc.NewRPC(WS_URL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Close()

	log.Println("Connection established.")

	data, err := os.ReadFile("spec-versions.json")
	if err != nil {
		log.Fatalf("read spec-versions.json: %v", err)
	}

	var blockInfos []BlockInfo
	err = json.Unmarshal(data, &blockInfos)
	if err != nil {
		log.Fatalf("unmarshal spec-versions.json: %v", err)
	}

	var chainMetadataList []rpc.ChainMetadata

	for i := 0; i < len(blockInfos); i += BATCH_SIZE {
		m := min(len(blockInfos), i+BATCH_SIZE)
		blockInfosBatch := blockInfos[i:m]

		blockhashes := Map(blockInfosBatch, func(blockInfo BlockInfo) string { return blockInfo.BlockHash })
		log.Printf("Fetching %d metadata %d..%d\n", len(blockhashes), i, m-1)

		res := batchGetMetadata(client, blockhashes)
		ExtendInPlace(&chainMetadataList, res)
	}

	log.Printf("Fetched %d metadata\n", len(chainMetadataList))

	types := make(map[string]struct{})
	for i, spec := range blockInfos {
		chainMetadata := chainMetadataList[i]
		if chainMetadata.Version >= 14 {
			continue
		}

		legacyMetadata, err := parseLegacyMetadata(&chainMetadata)
		if err != nil {
			log.Fatalf("parse legacy metadata for block %s: %v", spec.BlockHash, err)
		}

		specTypes := make(map[string]struct{})
		for _, module := range legacyMetadata.Modules {
			for _, event := range module.Events {
				for _, arg := range event.Args {
					specTypes[arg] = struct{}{}
				}
			}
			for _, call := range module.Calls {
				for _, arg := range call.Args {
					specTypes[arg.Type] = struct{}{}
				}
			}
		}

		sortedSpecTypes := make([]string, 0, len(specTypes))
		for t := range specTypes {
			sortedSpecTypes = append(sortedSpecTypes, t)
		}
		sort.Strings(sortedSpecTypes)

		fmt.Printf("Spec version %d types:\n", spec.SpecVersion)
		for _, t := range sortedSpecTypes {
			fmt.Printf("- %s\n", t)
		}
		fmt.Println()

		for t := range specTypes {
			types[t] = struct{}{}
		}
	}

	sortedTypes := make([]string, 0, len(types))
	for t := range types {
		sortedTypes = append(sortedTypes, t)
	}
	sort.Strings(sortedTypes)

	fmt.Println("All types:")
	for _, t := range sortedTypes {
		fmt.Printf("- %s\n", t)
	}
}

func batchGetMetadata(client *rpc.RPC, blockhashes []string) []rpc.ChainMetadata {
	var getMetadataArgs [][]any
	for _, blockhash := range blockhashes {
		getMetadataArgs = append(getMetadataArgs, []any{blockhash})
	}
	metadataList, err := rpc.SendMany(client, "state_getMetadata", getMetadataArgs, rpc.DecodeChainMetadata)
	if err != nil {
		log.Fatalf("get metadata: %v", err)
	}

	return metadataList
}

func Map[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = fn(v)
	}
	return result
}

func ExtendInPlace[T any](slice1 *[]T, slice2 []T) {
	*slice1 = append(*slice1, slice2...)
}

func parseLegacyMetadata(chainMetadata *rpc.ChainMetadata) (legacy.Metadata, error) {
	var legacyMetadata legacy.Metadata
	var err error

	r := scale.NewReader(chainMetadata.Data)
	metadata, err := decoder.DecodeMetadata(chainMetadata.Version, r)
	if err != nil {
		return legacyMetadata, fmt.Errorf("decode metadata: %w", err)
	}

	legacyMetadata, err = legacy.MakeMetadataFromAny(metadata)
	if err != nil {
		return legacyMetadata, fmt.Errorf("make legacy metadata: %w", err)
	}

	return legacyMetadata, nil
}
