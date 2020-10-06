package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/certikfoundation/shentu/x/shield/types"
)

func (k Keeper) SetPool(ctx sdk.Context, pool types.Pool) {
	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshalBinaryLengthPrefixed(pool)
	store.Set(types.GetPoolKey(pool.PoolID), bz)
}

func (k Keeper) GetPool(ctx sdk.Context, id uint64) (types.Pool, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetPoolKey(id))
	if bz == nil {
		return types.Pool{}, types.ErrNoPoolFound
	}
	var pool types.Pool
	k.cdc.MustUnmarshalBinaryLengthPrefixed(bz, &pool)
	return pool, nil
}

func (k Keeper) CreatePool(
	ctx sdk.Context, creator sdk.AccAddress, shield sdk.Coins, deposit types.MixedCoins, sponsor string,
	timeOfCoverage, blocksOfCoverage int64) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !creator.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}

	if !k.ValidatePoolDuration(ctx, timeOfCoverage, blocksOfCoverage) {
		return types.Pool{}, types.ErrPoolLifeTooShort
	}
	// check if shield is backed by admin's delegations
	provider, found := k.GetProvider(ctx, admin)
	if !found {
		k.addProvider(ctx, admin)
		provider, _ = k.GetProvider(ctx, admin)
	}
	provider.Collateral = provider.Collateral.Add(shield...)
	if shield.AmountOf(k.sk.BondDenom(ctx)).GT(provider.Available) {
		return types.Pool{}, sdkerrors.Wrapf(types.ErrInsufficientStaking,
			"available %s, shield %s", provider.Available, shield)
	}
	provider.Available = provider.Available.Sub(shield.AmountOf(k.sk.BondDenom(ctx)))

	// Store endTime. If not available, store endBlockHeight.
	var endTime, endBlockHeight int64
	startBlockHeight := ctx.BlockHeight()
	if timeOfCoverage != 0 {
		endTime = ctx.BlockTime().Unix() + timeOfCoverage
	} else if blocksOfCoverage != 0 {
		endBlockHeight = startBlockHeight + blocksOfCoverage
	}

	id := k.GetNextPoolID(ctx)
	depositDec := types.MixedDecCoinsFromMixedCoins(deposit)

	pool := types.NewPool(shield, depositDec, sponsor, endTime, startBlockHeight, endBlockHeight, id)

	// transfer deposit
	if err := k.DepositNativePremium(ctx, deposit.Native, creator); err != nil {
		return types.Pool{}, err
	}

	k.SetPool(ctx, pool)
	k.SetNextPoolID(ctx, id+1)
	k.SetProvider(ctx, admin, provider)
	k.SetCollateral(ctx, pool.PoolID, admin, types.NewCollateral(pool.PoolID, admin, shield))

	return pool, nil
}

func (k Keeper) UpdatePool(
	ctx sdk.Context, updater sdk.AccAddress, shield sdk.Coins, deposit types.MixedCoins, id uint64,
	additionalTime, additionalBlocks int64) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !updater.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}

	// check if shield is backed by admin's delegations
	provider, found := k.GetProvider(ctx, admin)
	if !found {
		return types.Pool{}, types.ErrNoDelegationAmount
	}
	provider.Collateral = provider.Collateral.Add(shield...)
	if shield.AmountOf(k.sk.BondDenom(ctx)).GT(provider.Available) {
		return types.Pool{}, sdkerrors.Wrapf(types.ErrInsufficientStaking,
			"available %s, shield %s", provider.Available, shield)
	}
	provider.Available = provider.Available.Sub(shield.AmountOf(k.sk.BondDenom(ctx)))

	pool, err := k.GetPool(ctx, id)
	if err != nil {
		return types.Pool{}, err
	}

	newCoverageTime := additionalTime + pool.EndTime - ctx.BlockTime().Unix()
	newCoverageBlocks := additionalBlocks + pool.EndBlockHeight - ctx.BlockHeight()
	if !k.ValidatePoolDuration(ctx, newCoverageTime, newCoverageBlocks) {
		return types.Pool{}, types.ErrPoolLifeTooShort
	}
	// Extend EndTime. If not available, extend EndBlockHeight.
	if additionalTime != 0 {
		if pool.EndTime == 0 {
			return types.Pool{}, types.ErrCannotExtend
		}
		pool.EndTime += additionalTime
	} else if additionalBlocks != 0 {
		if pool.EndBlockHeight == 0 {
			return types.Pool{}, types.ErrCannotExtend
		}
		pool.EndBlockHeight += additionalBlocks
	}

	pool.TotalCollateral = pool.TotalCollateral.Add(shield...)
	poolCertiKCollateral := k.GetPoolCertiKCollateral(ctx, pool)
	poolCertiKCollateral.Amount = poolCertiKCollateral.Amount.Add(shield...)
	k.SetCollateral(ctx, pool.PoolID, k.GetAdmin(ctx), poolCertiKCollateral)

	pool.Shield = pool.Shield.Add(shield...)
	pool.Premium = pool.Premium.Add(types.MixedDecCoinsFromMixedCoins(deposit))

	// transfer deposit and store
	if err := k.DepositNativePremium(ctx, deposit.Native, updater); err != nil {
		return types.Pool{}, err
	}
	k.SetPool(ctx, pool)
	k.SetProvider(ctx, admin, provider)

	return pool, nil
}

func (k Keeper) PausePool(ctx sdk.Context, updater sdk.AccAddress, id uint64) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !updater.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}
	pool, err := k.GetPool(ctx, id)
	if err != nil {
		return types.Pool{}, err
	}
	if !pool.Active {
		return types.Pool{}, types.ErrPoolAlreadyPaused
	}
	pool.Active = false
	k.SetPool(ctx, pool)
	return pool, nil
}

func (k Keeper) ResumePool(ctx sdk.Context, updater sdk.AccAddress, id uint64) (types.Pool, error) {
	admin := k.GetAdmin(ctx)
	if !updater.Equals(admin) {
		return types.Pool{}, types.ErrNotShieldAdmin
	}
	pool, err := k.GetPool(ctx, id)
	if err != nil {
		return types.Pool{}, err
	}
	if pool.Active {
		return types.Pool{}, types.ErrPoolAlreadyActive
	}
	pool.Active = true
	k.SetPool(ctx, pool)
	return pool, nil
}

// GetAllPools retrieves all pools in the store.
func (k Keeper) GetAllPools(ctx sdk.Context) (pools []types.Pool) {
	k.IterateAllPools(ctx, func(pool types.Pool) bool {
		pools = append(pools, pool)
		return false
	})
	return pools
}

// PoolEnded returns if pool has reached ending time and block height
func (k Keeper) PoolEnded(ctx sdk.Context, pool types.Pool) bool {
	if ctx.BlockTime().Unix() > pool.EndTime && ctx.BlockHeight() > pool.EndBlockHeight {
		return true
	}
	return false
}

// ClosePool closes the pool
func (k Keeper) ClosePool(ctx sdk.Context, poolID uint64) {
	// TODO: make sure nothing else needs to be done
	k.FreeCollaterals(ctx, poolID)
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.GetPoolKey(poolID))
}

// IterateAllPools iterates over the all the stored pools and performs a callback function.
func (k Keeper) IterateAllPools(ctx sdk.Context, callback func(pool types.Pool) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.PoolKey)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		var pool types.Pool
		k.cdc.MustUnmarshalBinaryLengthPrefixed(iterator.Value(), &pool)

		if callback(pool) {
			break
		}
	}
}

// ValidatePoolDuration validates new pool duration to be valid
func (k Keeper) ValidatePoolDuration(ctx sdk.Context, timeDuration, numBlocks int64) bool {
	poolParams := k.GetPoolParams(ctx)
	minPoolDuration := int64(poolParams.MinPoolLife.Seconds())
	return timeDuration > minPoolDuration || numBlocks*5 > minPoolDuration
}

// WithdrawFromPools withdraws coins from all pools to match total collateral to be less than or equal to total delegation.
func (k Keeper) WithdrawFromPools(ctx sdk.Context, addr sdk.AccAddress, amount sdk.Coins) {
	bondDenom := k.sk.BondDenom(ctx)
	provider, _ := k.GetProvider(ctx, addr)
	withdrawAmtDec := sdk.NewDecFromInt(amount.AmountOf(bondDenom))
	collateralDec := sdk.NewDecFromInt(provider.Collateral.AmountOf(bondDenom))
	proportion := withdrawAmtDec.Quo(collateralDec)
	if amount.IsAnyGT(provider.Collateral) {
		panic(types.ErrNotEnoughCollateral)
	}

	addrCollaterals := k.GetOnesCollaterals(ctx, addr)
	remainingWithdraw := amount
	for i, collateral := range addrCollaterals {
		var withdrawAmtDec sdk.Dec
		if i == len(addrCollaterals)-1 {
			withdrawAmtDec = sdk.NewDecFromInt(remainingWithdraw.AmountOf(bondDenom))
		} else {
			withdrawAmtDec = sdk.NewDecFromInt(collateral.Amount.AmountOf(bondDenom)).Mul(proportion)
		}
		withdrawAmt := withdrawAmtDec.TruncateInt()
		withdrawCoins := sdk.NewCoins(sdk.NewCoin(bondDenom, withdrawAmt))
		err := k.WithdrawCollateral(ctx, addr, collateral.PoolID, withdrawCoins)
		if err != nil {
			//TODO: address this error
			continue
		}
		remainingWithdraw = remainingWithdraw.Sub(withdrawCoins)
	}
}
