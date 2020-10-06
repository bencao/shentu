package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/certikfoundation/shentu/x/shield/types"
)

// GetPoolCollateral retrieves collateral for a pool-provider pair.
func (k Keeper) GetCollateral(ctx sdk.Context, poolID uint64, addr sdk.AccAddress) (types.Collateral, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetCollateralKey(poolID, addr))
	if bz == nil {
		return types.Collateral{}, false
	}
	var collateral types.Collateral
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &collateral)
	return collateral, true
}

// GetAllCollaterals gets all collaterals.
func (k Keeper) GetAllCollaterals(ctx sdk.Context) (collaterals []types.Collateral) {
	k.IterateCollaterals(ctx, func(collateral types.Collateral) bool {
		collaterals = append(collaterals, collateral)
		return false
	})
	return collaterals
}

// SetCollateral stores collateral based on pool-provider pair.
func (k Keeper) SetCollateral(ctx sdk.Context, poolID uint64, addr sdk.AccAddress, collateral types.Collateral) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(collateral)
	store.Set(types.GetCollateralKey(poolID, addr), bz)
}

// FreeCollateral frees collaterals deposited in a pool.
func (k Keeper) FreeCollaterals(ctx sdk.Context, poolID uint64) {
	store := ctx.KVStore(k.storeKey)
	k.IteratePoolCollaterals(ctx, poolID, func(collateral types.Collateral) bool {
		provider, _ := k.GetProvider(ctx, collateral.Provider)
		provider.Collateral = provider.Collateral.Sub(collateral.Amount)
		k.SetProvider(ctx, collateral.Provider, provider)
		store.Delete(types.GetCollateralKey(poolID, collateral.Provider))
		return false
	})
}

// IterateCollaterals iterates through all collaterals.
func (k Keeper) IterateCollaterals(ctx sdk.Context, callback func(collateral types.Collateral) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.CollateralKey)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var collateral types.Collateral
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &collateral)

		if callback(collateral) {
			break
		}
	}
}

// IteratePoolCollaterals iterates through collaterals in a pool
func (k Keeper) IteratePoolCollaterals(ctx sdk.Context, poolID uint64, callback func(collateral types.Collateral) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.GetPoolCollateralsKey(poolID))
	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var collateral types.Collateral
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &collateral)

		if callback(collateral) {
			break
		}
	}
}

// GetOnesCollaterals returns a community member's all collaterals.
func (k Keeper) GetOnesCollaterals(ctx sdk.Context, address sdk.AccAddress) (collaterals []types.Collateral) {
	k.IterateAllPools(ctx, func(pool types.Pool) bool {
		collateral, found := k.GetCollateral(ctx, pool.PoolID, address)
		if found {
			collaterals = append(collaterals, collateral)
		}
		return false
	})
	return collaterals
}

// GetPoolCertiKCollateral retrieves CertiK's provided collateral from a pool.
func (k Keeper) GetPoolCertiKCollateral(ctx sdk.Context, pool types.Pool) (collateral types.Collateral) {
	admin := k.GetAdmin(ctx)
	collateral, _ = k.GetCollateral(ctx, pool.PoolID, admin)
	return
}

// GetAllPoolCollaterals retrieves all collaterals in a pool.
func (k Keeper) GetAllPoolCollaterals(ctx sdk.Context, poolID uint64) (collaterals []types.Collateral) {
	k.IteratePoolCollaterals(ctx, poolID, func(collateral types.Collateral) bool {
		collaterals = append(collaterals, collateral)
		return false
	})
	return collaterals
}

// DepositCollateral deposits a community member's collateral for a pool.
func (k Keeper) DepositCollateral(ctx sdk.Context, from sdk.AccAddress, id uint64, amount sdk.Coins) error {
	pool, err := k.GetPool(ctx, id)
	if err != nil {
		return err
	}

	// check eligibility
	provider, found := k.GetProvider(ctx, from)
	if !found {
		k.addProvider(ctx, from)
		provider, _ = k.GetProvider(ctx, from)
	}
	provider.Collateral = provider.Collateral.Add(amount...)
	if amount.AmountOf(k.sk.BondDenom(ctx)).GT(provider.Available) {
		return types.ErrInsufficientStaking
	}
	provider.Available = provider.Available.Sub(amount.AmountOf(k.sk.BondDenom(ctx)))

	// update the pool, collateral and provider
	collateral, found := k.GetCollateral(ctx, pool.PoolID, from)
	if !found {
		collateral = types.NewCollateral(pool.PoolID, from, amount)
	} else {
		collateral.Amount = collateral.Amount.Add(amount...)
	}
	pool.TotalCollateral = pool.TotalCollateral.Add(amount...)
	k.SetPool(ctx, pool)
	k.SetCollateral(ctx, pool.PoolID, from, collateral)
	k.SetProvider(ctx, from, provider)

	return nil
}

// WithdrawCollateral withdraws a community member's collateral for a pool.
func (k Keeper) WithdrawCollateral(ctx sdk.Context, from sdk.AccAddress, id uint64, amount sdk.Coins) error {
	// retrieve the particular collateral to ensure that
	// amount is less than collateral minus collateral withdrawal
	collateral, found := k.GetCollateral(ctx, id, from)

	if !found {
		return types.ErrNoCollateralFound
	}

	withdrawable := collateral.Amount.Sub(collateral.Withdrawal)
	if amount.IsAnyGT(withdrawable) {
		return types.ErrOverWithdrawal
	}

	// insert into withdrawal queue
	poolParams := k.GetPoolParams(ctx)
	completionTime := ctx.BlockHeader().Time.Add(poolParams.WithdrawalPeriod)
	withdrawal := types.NewWithdrawal(id, from, amount)
	k.InsertWithdrawalQueue(ctx, withdrawal, completionTime)

	collateral.Withdrawal = collateral.Withdrawal.Add(amount...)
	k.SetCollateral(ctx, id, collateral.Provider, collateral)

	return nil
}
