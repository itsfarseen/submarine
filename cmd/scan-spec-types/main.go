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

type SpecVersionInfo struct {
	BlockNumber uint64 `json:"block"`
	BlockHash   string `json:"hash"`
	SpecVersion uint64 `json:"specVersion"`
}

func main() {
	client, err := rpc.NewRPC("ws://37.27.51.25:9944")
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Close()

	log.Println("Connection established.")

	data, err := os.ReadFile("spec-versions.json")
	if err != nil {
		log.Fatalf("read spec-versions.json: %v", err)
	}

	var specVersions []SpecVersionInfo
	err = json.Unmarshal(data, &specVersions)
	if err != nil {
		log.Fatalf("unmarshal spec-versions.json: %v", err)
	}

	allTypes := make(map[string]struct{})
	for _, spec := range specVersions {
		chainMetadata := client.GetMetadata(spec.BlockHash)
		if chainMetadata.Version >= 14 {
			continue
		}

		r := scale.NewReader(chainMetadata.Data)
		metadata, err := decoder.DecodeMetadata(chainMetadata.Version, r)
		if err != nil {
			log.Fatalf("decode metadata for block %s: %v", spec.BlockHash, err)
		}

		legacyMetadata, err := legacy.MakeMetadataFromAny(metadata)
		if err != nil {
			log.Fatalf("make legacy metadata for block %s: %v", spec.BlockHash, err)
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
			allTypes[t] = struct{}{}
		}
	}

	sortedAllTypes := make([]string, 0, len(allTypes))
	for t := range allTypes {
		sortedAllTypes = append(sortedAllTypes, t)
	}
	sort.Strings(sortedAllTypes)

	fmt.Println("All types:")
	for _, t := range sortedAllTypes {
		fmt.Printf("- %s\n", t)
	}
}
