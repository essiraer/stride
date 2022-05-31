package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/Stride-Labs/stride/x/stakeibc/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	icacontrollerkeeper "github.com/cosmos/ibc-go/v3/modules/apps/27-interchain-accounts/controller/keeper"
	ibckeeper "github.com/cosmos/ibc-go/v3/modules/core/keeper"
	ibctmtypes "github.com/cosmos/ibc-go/v3/modules/light-clients/07-tendermint/types"
)

type (
	Keeper struct {
		// *cosmosibckeeper.Keeper
		cdc                 codec.BinaryCodec
		storeKey            sdk.StoreKey
		memKey              sdk.StoreKey
		paramstore          paramtypes.Subspace
		icaControllerKeeper icacontrollerkeeper.Keeper
		ibcKeeper           ibckeeper.Keeper
		scopedKeeper        capabilitykeeper.ScopedKeeper
		bankKeeper 			bankKeeper.Keeper
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
	// channelKeeper cosmosibckeeper.ChannelKeeper,
	// portKeeper cosmosibckeeper.PortKeeper,
	// scopedKeeper cosmosibckeeper.ScopedKeeper,
	bankKeeper bankKeeper.Keeper,
	icacontrollerkeeper icacontrollerkeeper.Keeper,
	ibcKeeper ibckeeper.Keeper,
	scopedKeeper capabilitykeeper.ScopedKeeper,
) Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	// Scaffolding an ibc module using ignite creates a cosmosibckeeper.NewKeeper for the module,
	// but this is not compatible with ibc-v3
	// Keeper: cosmosibckeeper.NewKeeper(
	// 	types.PortKey,
	// 	storeKey,
	// 	channelKeeper,
	// 	portKeeper,
	// 	scopedKeeper,
	// ),
	return Keeper{
		cdc:                 cdc,
		storeKey:            storeKey,
		memKey:              memKey,
		paramstore:          ps,
		bankKeeper:          bankKeeper,
		icaControllerKeeper: icacontrollerkeeper,
		ibcKeeper:           ibcKeeper,
		scopedKeeper:        scopedKeeper,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// ClaimCapability claims the channel capability passed via the OnOpenChanInit callback
func (k *Keeper) ClaimCapability(ctx sdk.Context, cap *capabilitytypes.Capability, name string) error {
	return k.scopedKeeper.ClaimCapability(ctx, cap, name)
}

func (k Keeper) GetChainID(ctx sdk.Context, connectionID string) (string, error) {
	conn, found := k.ibcKeeper.ConnectionKeeper.GetConnection(ctx, connectionID)
	if !found {
		return "", fmt.Errorf("invalid connection id, \"%s\" not found", connectionID)
	}
	clientState, found := k.ibcKeeper.ClientKeeper.GetClientState(ctx, conn.ClientId)
	if !found {
		return "", fmt.Errorf("client id \"%s\" not found for connection \"%s\"", conn.ClientId, connectionID)
	}
	client, ok := clientState.(*ibctmtypes.ClientState)
	if !ok {
		return "", fmt.Errorf("invalid client state for client \"%s\" on connection \"%s\"", conn.ClientId, connectionID)
	}

	return client.ChainId, nil
}

func (k Keeper) GetConnectionId(ctx sdk.Context, portId string) (string, error) {
	icas := k.icaControllerKeeper.GetAllInterchainAccounts(ctx)
	for _, ica := range icas {
		if ica.PortId == portId {
			return ica.ConnectionId, nil
		}
	}
	return "", fmt.Errorf("portId %s has no associated connectionId", portId)
}

