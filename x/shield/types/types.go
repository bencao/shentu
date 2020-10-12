package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Pool struct {
	PoolID          uint64         `json:"pool_id" yaml:"pool_id"`
	Active          bool           `json:"active" yaml:"active"`
	Description     string         `json:"description" yaml:"description"`
	Sponsor         string         `json:"sponsor" yaml:"sponsor"`
	SponsorAddr     sdk.AccAddress `json:"sponsor_address" yaml:"sponsor_address"`
	Premium         MixedDecCoins  `json:"premium" yaml:"premium"`
	TotalCollateral sdk.Int        `json:"total_collateral" yaml:"total_collateral"`
	Available       sdk.Int        `json:"available" yaml:"available"`
	Shield          sdk.Coins      `json:"shield" yaml:"shield"`
	EndTime         int64          `json:"end_time" yaml:"end_time"`
}

func NewPool(shield sdk.Coins, totalCollateral sdk.Int, deposit MixedDecCoins, sponsor string, sponsorAddr sdk.AccAddress, endTime int64, id uint64) Pool {
	return Pool{
		Shield:          shield,
		Premium:         deposit,
		Sponsor:         sponsor,
		SponsorAddr:     sponsorAddr,
		Active:          true,
		TotalCollateral: totalCollateral,
		EndTime:         endTime,
		PoolID:          id,
	}
}

type Collateral struct {
	PoolID            uint64             `json:"pool_id" yaml:"pool_id"`
	Provider          sdk.AccAddress     `json:"provider" yaml:"provider"`
	Amount            sdk.Int            `json:"amount" yaml:"amount"`
	Withdrawing       sdk.Int            `json:"withdrawing" yaml:"withdrawing"`
	LockedCollaterals []LockedCollateral `json:"locked_collaterals" yaml:"locked_collaterals"`
}

func NewCollateral(pool Pool, provider sdk.AccAddress, amount sdk.Int) Collateral {
	return Collateral{
		PoolID:   pool.PoolID,
		Provider: provider,
		Amount:   amount,
	}
}

// Provider tracks A or C's total delegation, total collateral,
// and rewards.
type Provider struct {
	// address of the provider
	Address sdk.AccAddress `json:"address" yaml:"address"`
	// bonded delegations
	DelegationBonded sdk.Int `json:"delegation_bonded" yaml:"delegation_bonded"`
	// collateral, including that in withdraw queue and excluding that being locked
	Collateral sdk.Int `json:"collateral" yaml:"collateral"`
	// coins locked because of claim proposals
	TotalLocked sdk.Int `json:"total_locked" yaml:"total_locked"`
	// amount of coins staked but not in any pool
	Available sdk.Int `json:"available" yaml:"available"`
	// amount of collateral that is in withdraw queue
	Withdrawing sdk.Int `json:"withrawal" yaml:"withdraw"`
	// rewards to be claimed
	Rewards MixedDecCoins `json:"rewards" yaml:"rewards"`
}

func NewProvider(addr sdk.AccAddress) Provider {
	return Provider{
		Address:          addr,
		DelegationBonded: sdk.ZeroInt(),
		Collateral:       sdk.ZeroInt(),
		TotalLocked:      sdk.ZeroInt(),
		Available:        sdk.ZeroInt(),
		Withdrawing:      sdk.ZeroInt(),
	}
}

type Purchase struct {
	TxHash             []byte         `json:"tx_hash" yaml:"tx_hash"`
	PoolID             uint64         `json:"pool_id" yaml:"pool_id"`
	Shield             sdk.Coins      `json:"shield" yaml:"shield"`
	StartBlockHeight   int64          `json:"start_block_height" yaml:"start_block_height"`
	ProtectionEndTime  time.Time      `json:"protection_end_time" yaml:"protection_end_time"`
	ClaimPeriodEndTime time.Time      `json:"claim_period_end_time" yaml:"claim_period_end_time"`
	Description        string         `json:"description" yaml:"description"`
	Purchaser          sdk.AccAddress `json:"purchaser" yaml:"purchaser"`
}

type PurchaseTxHash struct {
	TxHash []byte `json:"tx_hash" yaml:"tx_hash"`
}

func NewPurchase(txhash []byte, poolID uint64, shield sdk.Coins, startBlockHeight int64, protectionEndTime, claimPeriodEndTime time.Time, description string, purchaser sdk.AccAddress) Purchase {
	return Purchase{
		TxHash:             txhash,
		PoolID:             poolID,
		Shield:             shield,
		StartBlockHeight:   startBlockHeight,
		ProtectionEndTime:  protectionEndTime,
		ClaimPeriodEndTime: claimPeriodEndTime,
		Description:        description,
		Purchaser:          purchaser,
	}
}

// Withdraw stores an ongoing withdraw of pool collateral.
type Withdraw struct {
	PoolID         uint64         `json:"pool_id" yaml:"pool_id"`
	Address        sdk.AccAddress `json:"address" yaml:"address"`
	Amount         sdk.Int        `json:"amount" yaml:"amount"`
	CompletionTime time.Time      `json:"completion_time" yaml:"completion_time"`
}

func NewWithdraw(poolID uint64, addr sdk.AccAddress, amount sdk.Int, completionTime time.Time) Withdraw {
	return Withdraw{
		PoolID:         poolID,
		Address:        addr,
		Amount:         amount,
		CompletionTime: completionTime,
	}
}

type Withdraws []Withdraw
