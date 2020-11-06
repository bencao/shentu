package types

import sdk "github.com/cosmos/cosmos-sdk/types"

const (
	QueryPoolByID            = "pool_id"
	QueryPoolBySponsor       = "pool_sponsor"
	QueryPools               = "pools"
	QueryPurchase            = "purchase"
	QueryPurchaseList        = "purchase_list"
	QueryPurchaserPurchases  = "purchaser_purchases"
	QueryPoolPurchases       = "pool_purchases"
	QueryPurchases           = "purchases"
	QueryProviderCollaterals = "provider_collaterals"
	QueryPoolCollaterals     = "pool_collaterals"
	QueryProvider            = "provider"
	QueryProviders           = "providers"
	QueryPoolParams          = "pool_params"
	QueryClaimParams         = "claim_params"
	QueryStatus              = "status"
	QueryStakedForShield     = "staked_for_shield"
	QueryShieldStakingRate   = "shield_staking_rate"
	QueryReimbursement       = "reimbursement"
	QueryReimbursements      = "reimbursements"
)

type QueryResStatus struct {
	TotalCollateral         sdk.Int       `json:"total_collateral" yaml:"total_collateral"`
	TotalShield             sdk.Int       `json:"total_shield" yaml:"total_shield"`
	TotalWithdrawing        sdk.Int       `json:"total_withdrawing" yaml:"total_withdrawing"`
	CurrentServiceFees      MixedDecCoins `json:"current_service_fees" yaml:"current_service_fees"`
	RemainingServiceFees    MixedDecCoins `json:"remaining_service_fees" yaml:"remaining_service_fees"`
	GlobalShieldStakingPool sdk.Int       `json:"global_shield_staking_pool" yaml:"global_shield_staking_pool"`
}

func NewQueryResStatus(totalCollateral, totalShield, totalWithdrawing sdk.Int, currentServiceFees, remainingServiceFees MixedDecCoins,
	globalStakingPool sdk.Int) QueryResStatus {
	return QueryResStatus{
		TotalCollateral:         totalCollateral,
		TotalShield:             totalShield,
		TotalWithdrawing:        totalWithdrawing,
		CurrentServiceFees:      currentServiceFees,
		RemainingServiceFees:    remainingServiceFees,
		GlobalShieldStakingPool: globalStakingPool,
	}
}

// QueryPaginationParams provides basic pagination parameters
// for queries in shield module.
type QueryPaginationParams struct {
	Page  int
	Limit int
}

// NewQueryPaginationParams creates new instance of the
// QueryPaginationParams.
func NewQueryPaginationParams(page, limit int) QueryPaginationParams {
	return QueryPaginationParams{
		Page:  page,
		Limit: limit,
	}
}
