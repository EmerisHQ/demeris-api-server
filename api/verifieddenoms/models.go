package verifieddenoms

import "github.com/emerishq/demeris-backend-models/cns"

type VerifiedDenom struct {
	cns.Denom
	ChainName string `json:"chain_name"`
}
type VerifiedDenomsResponse struct {
	VerifiedDenoms []VerifiedDenom `json:"verified_denoms"`
}
