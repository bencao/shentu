package shield

import (
	"fmt"
	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/certikfoundation/shentu/common"
	"github.com/certikfoundation/shentu/x/shield/types"
)

// BeginBlock executes logics to begin a block.
func BeginBlock(ctx sdk.Context, req abci.RequestBeginBlock, k Keeper) {
	if ctx.BlockHeight() == common.Update1Height {
		k.SetShieldStakingRate(ctx, types.DefaultStakingShieldRate)
	}
}

// EndBlocker processes premium payment at every block.
func EndBlocker(ctx sdk.Context, k Keeper) {
	// Remove expired purchases and distribute service fees.
	k.RemoveExpiredPurchasesAndDistributeFees(ctx)

	// Process completed withdraws.
	k.DequeueCompletedWithdrawQueue(ctx)

	// Close pools who do not have any shield and shield limits are set to zero.
	k.ClosePools(ctx)

	fmt.Printf(">>> NOW: %s\n", ctx.BlockTime())
}
