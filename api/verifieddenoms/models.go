package verifieddenoms

import "github.com/allinbits/demeris-backend-models/cns"

type verifiedDenom struct {
	cns.Denom
	ChainName string `json:"chain_name"`
}
type verifiedDenomsResponse struct {
	VerifiedDenoms []verifiedDenom `json:"verified_denoms"`
}
