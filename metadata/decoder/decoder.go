package decoder

import (
	"fmt"
	scale_v10 "submarine/metadata/generated/v10"
	scale_v11 "submarine/metadata/generated/v11"
	scale_v12 "submarine/metadata/generated/v12"
	scale_v14 "submarine/metadata/generated/v14"
	scale_v9 "submarine/metadata/generated/v9"
	"submarine/scale"
)

func DecodeMetadata(version uint, r *scale.Reader) (any, error) {
	switch version {
	case 9:
		meta, err := scale_v9.DecodeMetadata(r)
		if err != nil {
			return nil, fmt.Errorf("v9: %w", err)
		}
		return &meta, nil
	case 10:
		meta, err := scale_v10.DecodeMetadata(r)
		if err != nil {
			return nil, fmt.Errorf("v10: %w", err)
		}
		return &meta, nil
	case 11:
		meta, err := scale_v11.DecodeMetadata(r)
		if err != nil {
			return nil, fmt.Errorf("v11: %w", err)
		}
		return &meta, nil
	case 12, 13:
		meta, err := scale_v12.DecodeMetadata(r)
		if err != nil {
			return nil, fmt.Errorf("v12: %w", err)
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
