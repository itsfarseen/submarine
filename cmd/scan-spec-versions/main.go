package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strconv"
	"submarine/rpc"
)

type SpecVersionInfo struct {
	BlockNumber uint64
	BlockHash   string
	SpecVersion uint64
}

func main() {
	var parallelRequests = flag.Int("parallel-requests", 250, "Number of parallel requests to make.")
	flag.Parse()

	client, err := rpc.NewRPC("ws://37.27.51.25:9944")
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer client.Close()

	log.Println("Connection established.")

	var latestHeader rpc.BlockHeader
	err = client.Send("chain_getHeader", nil).As(&latestHeader)
	if err != nil {
		log.Fatalf("get latest header: %v", err)
	}
	latestBlockNumber, _ := hexToDecimal(latestHeader.Number)
	log.Printf("Latest Block Number: %d", latestBlockNumber)

	initialSpecVersions, err := getSpecVersions(client, []int{0, int(latestBlockNumber)})
	if err != nil {
		log.Fatalf("initial spec versions: %v", err)
	}

	startBlock := initialSpecVersions[0]
	endBlock := initialSpecVersions[1]

	specVersionNumbers := make([]int, 0)
	sf := SpecFinder{client, *parallelRequests, make(map[int]SpecVersionInfo)}

	err = sf.findSpecVersions(startBlock, endBlock)
	if err != nil {
		log.Fatalf("find spec versions: %v", err)
	}

	for v := range sf.specs {
		specVersionNumbers = append(specVersionNumbers, v)
	}
	slices.Sort(specVersionNumbers)

	var output string

	output += "["
	for i, k := range specVersionNumbers {
		if i > 0 {
			output += ","
		}
		output += "\n"

		info := sf.specs[k]
		output += fmt.Sprintf(`  { "specVersion": %d, "block": %d, "hash": "%s" }`, k, info.BlockNumber, info.BlockHash)
	}
	output += "\n]"

	fmt.Println(output)
	os.WriteFile("spec-versions.json", []byte(output), 0)
}

func getBlockHashes(client *rpc.RPC, blockNumbers []int) ([]string, error) {
	args := make([][]any, len(blockNumbers))
	for i, n := range blockNumbers {
		args[i] = []any{n}
	}

	return rpc.SendMany(
		client,
		"chain_getBlockHash",
		args,
		func(pr *rpc.PendingRequest) (string, error) {
			return pr.AsString()
		},
	)
}

func getRuntimeVersions(client *rpc.RPC, blockHashes []string) ([]rpc.RuntimeVersion, error) {
	args := make([][]any, len(blockHashes))
	for i, n := range blockHashes {
		args[i] = []any{n}
	}

	return rpc.SendMany(
		client,
		"state_getRuntimeVersion",
		args,
		func(pr *rpc.PendingRequest) (rpc.RuntimeVersion, error) {
			var x rpc.RuntimeVersion
			err := pr.As(&x)
			return x, err
		},
	)
}

func getSpecVersions(client *rpc.RPC, blockNumbers []int) ([]SpecVersionInfo, error) {
	blockHashes, err := getBlockHashes(client, blockNumbers)
	if err != nil {
		return nil, err
	}
	runtimeVersions, err := getRuntimeVersions(client, blockHashes)
	if err != nil {
		return nil, err
	}

	specVersions := make([]SpecVersionInfo, len(blockNumbers))
	for i, blockNumber := range blockNumbers {
		specVersions[i] = SpecVersionInfo{
			BlockNumber: uint64(blockNumber),
			BlockHash:   blockHashes[i],
			SpecVersion: uint64(runtimeVersions[i].SpecVersion),
		}
	}
	return specVersions, nil
}

type SpecFinder struct {
	client           *rpc.RPC
	parallelRequests int
	specs            map[int]SpecVersionInfo
}

func (sf *SpecFinder) add(specVersion SpecVersionInfo) {
	sf.specs[int(specVersion.SpecVersion)] = specVersion
}

func (sf *SpecFinder) findSpecVersions(
	startBlock, endBlock SpecVersionInfo,
) error {
	if (endBlock.SpecVersion-startBlock.SpecVersion) <= 1 || (endBlock.BlockNumber-startBlock.BlockNumber) <= 1 {
		sf.add(startBlock)
		sf.add(endBlock)
		return nil
	}

	fmt.Printf(
		"finding %d(%d)..%d(%d)\n",
		startBlock.BlockNumber, startBlock.SpecVersion, endBlock.BlockNumber, endBlock.SpecVersion,
	)

	interval := (endBlock.BlockNumber - startBlock.BlockNumber) / uint64(sf.parallelRequests)
	if interval == 0 {
		interval = 1
	}

	blockNumbers := make([]uint64, 0)
	for i := startBlock.BlockNumber + interval; i < endBlock.BlockNumber; i += interval {
		blockNumbers = append(blockNumbers, i)
	}

	var blockHashesArgsList [][]any
	for _, n := range blockNumbers {
		blockHashesArgsList = append(blockHashesArgsList, []any{n})
	}

	blockHashes, err := rpc.SendMany(
		sf.client,
		"chain_getBlockHash",
		blockHashesArgsList,
		func(p *rpc.PendingRequest) (string, error) {
			return p.AsString()
		},
	)
	if err != nil {
		return err
	}

	var runtimeVersionsArgsList [][]any
	for _, hash := range blockHashes {
		runtimeVersionsArgsList = append(runtimeVersionsArgsList, []any{hash})
	}

	runtimeVersions, err := rpc.SendMany(
		sf.client,
		"state_getRuntimeVersion",
		runtimeVersionsArgsList,
		func(pr *rpc.PendingRequest) (rpc.RuntimeVersion, error) {
			var x rpc.RuntimeVersion
			err := pr.As(&x)
			return x, err
		},
	)
	if err != nil {
		return err
	}

	blocks := make([]SpecVersionInfo, 0)
	blocks = append(blocks, startBlock)

	for i := range blockNumbers {
		block := SpecVersionInfo{
			BlockNumber: blockNumbers[i],
			BlockHash:   blockHashes[i],
			SpecVersion: uint64(runtimeVersions[i].SpecVersion),
		}
		blocks = append(blocks, block)
	}

	blocks = append(blocks, endBlock)

	for i := range blocks {
		if i == 0 {
			continue
		}

		err := sf.findSpecVersions(blocks[i-1], blocks[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func hexToDecimal(hex string) (uint64, error) {
	if len(hex) > 2 && hex[:2] == "0x" {
		hex = hex[2:]
	}
	return strconv.ParseUint(hex, 16, 64)
}
