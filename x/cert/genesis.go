package cert

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/certikfoundation/shentu/x/cert/internal/keeper"
	"github.com/certikfoundation/shentu/x/cert/internal/types"
)

func InitDefaultGenesis(ctx sdk.Context, k keeper.Keeper) {
	InitGenesis(ctx, k, types.DefaultGenesisState())
}

// InitGenesis initialize default parameters and the keeper's address to pubkey map.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, data types.GenesisState) {
	certifiers := data.Certifiers
	validators := data.Validators
	platforms := data.Platforms
	certificates := data.Certificates
	libraries := data.Libraries

	for _, certifier := range certifiers {
		k.SetCertifier(ctx, certifier)
	}
	if len(certifiers) > 0 {
		cert := certifiers[0].Address
		for _, platform := range platforms {
			_ = k.CertifyPlatform(ctx, cert, platform.Validator, platform.Description)
		}
	}
	for _, validator := range validators {
		k.SetValidator(ctx, validator.PubKey, validator.Certifier)
	}
	
	sort.Slice(certificates, func(i, j int) bool { 
		return certificates[i].ID() < certificates[j].ID()
	})
	for _, certificate := range certificates {
		k.AddCertIDToCertifier(ctx, certificate.Certifier(), certificate.ID())
		k.SetContentCertID(ctx, certificate.Type(), certificate.RequestContent(), certificate.ID())
		k.SetCertificate(ctx, certificate)
	}
	for _, library := range libraries {
		k.SetLibrary(ctx, library.Address, library.Publisher)
	}
	k.SetNextCertificateID(ctx, data.NextCertificateID)
}

// ExportGenesis writes the current store values to a genesis file, which can be imported again with InitGenesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) types.GenesisState {
	certifiers := k.GetAllCertifiers(ctx)
	validators := k.GetAllValidators(ctx)
	platforms := k.GetAllPlatforms(ctx)
	certificates := k.GetAllCertificates(ctx)
	libraries := k.GetAllLibraries(ctx)
	nextCertID := k.GetNextCertificateID(ctx)

	return GenesisState{
		Certifiers:        certifiers,
		Validators:        validators,
		Platforms:         platforms,
		Certificates:      certificates,
		Libraries:         libraries,
		NextCertificateID: nextCertID,
	}
}
