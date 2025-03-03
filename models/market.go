package models

// VaultPair 表示配对的vault地址
type VaultPair struct {
	A string `json:"a"`
	B string `json:"b"`
}

// MarketParams 市场参数结构
type MarketParams struct {
	AddressLookupTableAddress string     `json:"addressLookupTableAddress,omitempty"`
	RoutingGroup             int        `json:"routingGroup,omitempty"`
	VaultLpMint             *VaultPair `json:"vaultLpMint,omitempty"`
	VaultToken              *VaultPair `json:"vaultToken,omitempty"`
	SerumAsks               string     `json:"serumAsks,omitempty"`
	SerumBids               string     `json:"serumBids,omitempty"`
	SerumCoinVaultAccount   string     `json:"serumCoinVaultAccount,omitempty"`
	SerumEventQueue         string     `json:"serumEventQueue,omitempty"`
	SerumPcVaultAccount     string     `json:"serumPcVaultAccount,omitempty"`
	SerumVaultSigner        string     `json:"serumVaultSigner,omitempty"`
}

// Market 基础市场数据结构
type Market struct {
	Pubkey string        `json:"pubkey"`
	Owner  string        `json:"owner"`
	Params *MarketParams `json:"params,omitempty"`
}

// MarketData 市场数据输出结构
type MarketData struct {
	Address                  string            `json:"address"`
	Owner                   string            `json:"owner"`
	Params                  map[string]string `json:"params,omitempty"`
	AddressLookupTableAddress string            `json:"addressLookupTableAddress,omitempty"`
} 