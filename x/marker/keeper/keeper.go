package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"

	"github.com/provenance-io/provenance/x/marker/types"
)

// Handler is a handler function for use with IterateRecords.
type Handler func(record types.MarkerAccountI) error

// MarkerKeeperI provides a read/write iterate interface to marker acccounts in the auth account keeper store
type MarkerKeeperI interface {
	// Returns a new marker instance with the address and baseaccount assigned.  Does not save to auth store
	NewMarker(sdk.Context, types.MarkerAccountI) types.MarkerAccountI

	// GetMarker looks up a marker by a given address
	GetMarker(sdk.Context, sdk.AccAddress) (types.MarkerAccountI, error)
	// Set a marker in the auth account store
	SetMarker(sdk.Context, types.MarkerAccountI)
	// Remove a marker from the auth account store
	RemoveMarker(sdk.Context, types.MarkerAccountI)

	GetEscrow(sdk.Context, types.MarkerAccountI) sdk.Coins

	// IterateMarker processes all markers with the given handler function.
	IterateMarkers(sdk.Context, func(types.MarkerAccountI) bool)

	// GetAuthority returns the signing authority
	GetAuthority() string
}

// Keeper defines the name module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set account specific params
	paramSpace paramtypes.Subspace

	// To check whether accounts exist for addresses.
	authKeeper types.AccountKeeper

	// To check whether accounts exist for addresses.
	authzKeeper types.AuthzKeeper

	// To handle movement of coin between accounts and check total supply
	bankKeeper types.BankKeeper

	// To pass through grant creation for callers with admin access on a marker.
	feegrantKeeper types.FeeGrantKeeper

	// To access attributes for addresses
	attrKeeper types.AttrKeeper
	// To access names and normalize required attributes
	nameKeeper types.NameKeeper

	// Key to access the key-value store from sdk.Context.
	storeKey storetypes.StoreKey

	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// the signing authority for the gov proposals
	authority string

	markerModuleAddr sdk.AccAddress

	ibcTransferModuleAddr sdk.AccAddress

	// Used to transfer the ibc marker
	ibcTransferServer types.IbcTransferMsgServer

	// reqAttrBypassAddrs is a set of addresses that are allowed to bypass the required attribute check.
	// When sending to one of these, if there are required attributes, it behaves as if the addr has them;
	// if there aren't required attributes, the sender still needs transfer permission.
	// When sending from one of these, if there are required attributes, the destination must have them;
	// if there aren't required attributes, it behaves as if the sender has transfer permission.
	reqAttrBypassAddrs types.ImmutableAccAddresses
}

// NewKeeper returns a marker keeper. It handles:
// - managing MarkerAccounts
// - enforcing permissions for marker creation/deletion/management
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
	authKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
	authzKeeper types.AuthzKeeper,
	feegrantKeeper types.FeeGrantKeeper,
	attrKeeper types.AttrKeeper,
	nameKeeper types.NameKeeper,
	ibcTransferServer types.IbcTransferMsgServer,
	reqAttrBypassAddrs []sdk.AccAddress,
) Keeper {
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	rv := Keeper{
		paramSpace:            paramSpace,
		authKeeper:            authKeeper,
		authzKeeper:           authzKeeper,
		bankKeeper:            bankKeeper,
		feegrantKeeper:        feegrantKeeper,
		attrKeeper:            attrKeeper,
		nameKeeper:            nameKeeper,
		storeKey:              key,
		cdc:                   cdc,
		authority:             authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		markerModuleAddr:      authtypes.NewModuleAddress(types.CoinPoolName),
		ibcTransferModuleAddr: authtypes.NewModuleAddress(ibctypes.ModuleName),
		ibcTransferServer:     ibcTransferServer,
		reqAttrBypassAddrs:    types.NewImmutableAccAddresses(reqAttrBypassAddrs),
	}
	bankKeeper.AppendSendRestriction(rv.SendRestrictionFn)
	return rv
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

var _ MarkerKeeperI = &Keeper{}

// NewMarker returns a new marker instance with the address and baseaccount assigned.  Does not save to auth store
func (k Keeper) NewMarker(ctx sdk.Context, marker types.MarkerAccountI) types.MarkerAccountI {
	return k.authKeeper.NewAccount(ctx, marker).(types.MarkerAccountI)
}

// GetMarker looks up a marker by a given address
func (k Keeper) GetMarker(ctx sdk.Context, address sdk.AccAddress) (types.MarkerAccountI, error) {
	mac := k.authKeeper.GetAccount(ctx, address)
	if mac != nil {
		macc, ok := mac.(types.MarkerAccountI)
		if !ok {
			return nil, fmt.Errorf("account at %s is not a marker account", address.String())
		}
		return macc, nil
	}
	return nil, nil
}

// SetMarker sets a marker in the auth account store will panic if the marker account is not valid or
// if the auth module account keeper fails to marshall the account.
func (k Keeper) SetMarker(ctx sdk.Context, marker types.MarkerAccountI) {
	store := ctx.KVStore(k.storeKey)

	if err := marker.Validate(); err != nil {
		panic(err)
	}
	k.authKeeper.SetAccount(ctx, marker)
	store.Set(types.MarkerStoreKey(marker.GetAddress()), marker.GetAddress())
}

// RemoveMarker removes a marker from the auth account store. Note: if the account holds coins this will
// likely cause an invariant constraint violation for the coin supply
func (k Keeper) RemoveMarker(ctx sdk.Context, marker types.MarkerAccountI) {
	store := ctx.KVStore(k.storeKey)
	k.authKeeper.RemoveAccount(ctx, marker)

	store.Delete(types.MarkerStoreKey(marker.GetAddress()))
}

// IterateMarkers  iterates all markers with the given handler function.
func (k Keeper) IterateMarkers(ctx sdk.Context, cb func(marker types.MarkerAccountI) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.MarkerStoreKeyPrefix)

	defer iterator.Close()
	for ; iterator.Valid(); iterator.Next() {
		account := k.authKeeper.GetAccount(ctx, iterator.Value())
		ma, ok := account.(types.MarkerAccountI)
		if !ok {
			panic(fmt.Errorf("invalid account type in marker account registry"))
		}
		if cb(ma) {
			break
		}
	}
}

// GetEscrow returns the balances of all coins held in escrow in the marker
func (k Keeper) GetEscrow(ctx sdk.Context, marker types.MarkerAccountI) sdk.Coins {
	return k.bankKeeper.GetAllBalances(ctx, marker.GetAddress())
}

// GetAuthority is signer of the proposal
func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) IsSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) bool {
	store := ctx.KVStore(k.storeKey)
	return store.Has(types.DenySendKey(markerAddr, senderAddr))
}

func (k Keeper) AddSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.DenySendKey(markerAddr, senderAddr), []byte{})
}

func (k Keeper) RemoveSendDeny(ctx sdk.Context, markerAddr, senderAddr sdk.AccAddress) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.DenySendKey(markerAddr, senderAddr))
}

// GetReqAttrBypassAddrs returns a deep copy of the addresses that bypass the required attributes checking.
func (k Keeper) GetReqAttrBypassAddrs() []sdk.AccAddress {
	return k.reqAttrBypassAddrs.GetSlice()
}

// IsReqAttrBypassAddr returns true if the provided addr can bypass the required attributes checking.
func (k Keeper) IsReqAttrBypassAddr(addr sdk.AccAddress) bool {
	return k.reqAttrBypassAddrs.Has(addr)
}
