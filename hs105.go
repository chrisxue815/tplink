package tplink

// TP-Link HS105 smart plug
type HS105 struct {
	// hs105 has same feature as hs100
	HS100
}

func NewHS105(ip string) *HS100 {
	return &HS100{ip: ip}
}
