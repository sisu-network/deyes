package types

type Account struct {
	Summary  *AccountSummary  `json:"summary"`
	Token    *AccountToken    `json:"token"`
	Sequence *AccountSequence `json:"sequence"`
	Keys     *AccountKeys     `json:"keys"`
	Dpos     *Dpos            `json:"dpos"`
}

type AccountSummary struct {
	Address          string `json:"address"`
	Balance          string `json:"balance"`
	Username         string `json:"username"`
	PublicKey        string `json:"publicKey"`
	IsDelegate       bool   `json:"isDelegate"`
	IsMultisignature bool   `json:"isMultisignature"`
}

type AccountToken struct {
	Balance string `json:"balance"`
}

type AccountSequence struct {
	Nonce string `json:"nonce"`
}

type AccountKeys struct {
	NumberOfSignatures uint64   `json:"numberOfSignatures"`
	MandatoryKeys      []string `json:"mandatoryKeys"`
	OptionalKeys       []string `json:"optionalKeys"`
}

type Dpos struct {
	Delegate *DposDelegate `json:"delegate"`
}

type DposDelegate struct {
	Username                string `json:"username"`
	ConsecutiveMissedBlocks uint64 `json:"consecutiveMissedBlocks"`
	LastForgedHeight        int32  `json:"lastForgedHeight"`
	IsBanned                bool   `json:"isBanned"`
	VoteWeight              string `json:"voteWeight"`
	TotalVotesReceived      string `json:"totalVotesReceived"`
	Rank                    uint64 `json:"Rank"`
	Status                  string `json:"status"`
}
