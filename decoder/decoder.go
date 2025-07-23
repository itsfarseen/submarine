package decoder

import (
	"fmt"
	. "submarine/decoder/models"
	v13 "submarine/decoder/v13"
	v14 "submarine/decoder/v14"
	"submarine/scale"
	scale_v13 "submarine/scale/v13"
	scale_v14 "submarine/scale/v14"
)

func DecodeMetadata(version uint, r *scale.Reader) (any, error) {
	switch version {
	case 13:
		meta, err := scale_v13.DecodeMetadata(r)
		if err != nil {
			return nil, fmt.Errorf("v13: %w", err)
		}
		return &meta, nil
	case 14:
		meta, err := scale_v14.DecodeMetadata(r)
		if err != nil {
			return nil, err
		}
		return &meta, nil
	default:
		return nil, fmt.Errorf("unsupported metadata version: %d", version)
	}
}

func DecodeExtrinsic(metadata any, extrinsicBytes []byte) (*DecodedExtrinsic, error) {
	switch meta := metadata.(type) {
	case *scale_v14.Metadata:
		return v14.DecodeExtrinsic(meta, extrinsicBytes)
	case *scale_v13.Metadata:
		return v13.DecodeExtrinsic(meta, extrinsicBytes)
	default:
		return nil, fmt.Errorf("unsupported metadata type for extrinsic decoding: %T", metadata)
	}
}

func DecodeEvents(metadata any, eventBytes []byte) ([]EventRecord, error) {
	switch meta := metadata.(type) {
	case *scale_v14.Metadata:
		return v14.DecodeEvents(meta, eventBytes)
	case *scale_v13.Metadata:
		return v13.DecodeEvents(meta, eventBytes)
	default:
		return nil, fmt.Errorf("unsupported metadata type for event decoding: %T", metadata)
	}
}
