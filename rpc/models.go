package rpc

type BlockHeader struct {
	ParentHash     string `json:"parentHash"`
	Number         string `json:"number"`
	StateRoot      string `json:"stateRoot"`
	ExtrinsicsRoot string `json:"extrinsicsRoot"`
	Digest         struct {
		Logs []string `json:"logs"`
	} `json:"digest"`
}

type SignedBlock struct {
	Block Block `json:"block"`
}

type Block struct {
	Header     BlockHeader `json:"header"`
	Extrinsics []string    `json:"extrinsics"`
}
