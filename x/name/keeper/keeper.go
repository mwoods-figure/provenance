package keeper

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"

	"github.com/google/uuid"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/provenance-io/provenance/x/name/types"
)

// Keeper defines the name module Keeper
type Keeper struct {
	// The reference to the Paramstore to get and set account specific params
	paramSpace paramtypes.Subspace

	// Key to access the key-value store from sdk.Context.
	storeKey storetypes.StoreKey

	// The codec for binary encoding/decoding.
	cdc codec.BinaryCodec

	// the signing authority for the gov proposals
	authority string

	attrKeeper types.AttributeKeeper
}

// NewKeeper returns a name keeper. It handles:
// - managing a hierarchy of names
// - enforcing permissions for name creation/deletion
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec,
	key storetypes.StoreKey,
	paramSpace paramtypes.Subspace,
) Keeper {
	// set KeyTable if it has not already been set
	if !paramSpace.HasKeyTable() {
		paramSpace = paramSpace.WithKeyTable(types.ParamKeyTable())
	}

	return Keeper{
		storeKey:   key,
		paramSpace: paramSpace,
		cdc:        cdc,
		authority:  authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	}
}

// Logger returns a module-specific logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// SetAttributeKeeper sets the attribute keeper
func (k *Keeper) SetAttributeKeeper(ak types.AttributeKeeper) {
	if k.attrKeeper != nil && ak != nil && k.attrKeeper != ak {
		panic("the attribute keeper has already been set")
	}
	k.attrKeeper = ak
}

// ResolvesTo to determines whether a name resolves to a given address.
func (k Keeper) ResolvesTo(ctx sdk.Context, name string, addr sdk.AccAddress) bool {
	stored, err := k.GetRecordByName(ctx, name)
	if err != nil {
		return false
	}
	return addr.String() == stored.Address
}

// SetNameRecord binds a name to an address.
func (k Keeper) SetNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	var err error
	if name, err = k.Normalize(ctx, name); err != nil {
		return err
	}
	if err = types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	if err = k.addRecord(ctx, name, addr, restrict, false); err != nil {
		return err
	}

	nameBoundEvent := types.NewEventNameBound(addr.String(), name, restrict)

	return ctx.EventManager().EmitTypedEvent(nameBoundEvent)
}

// UpdateNameRecord updates the owner address and restricted flag on a name.
func (k Keeper) UpdateNameRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict bool) error {
	var err error
	if name, err = k.Normalize(ctx, name); err != nil {
		return err
	}
	if err = types.ValidateAddress(addr); err != nil {
		return types.ErrInvalidAddress.Wrap(err.Error())
	}

	// If there's an existing record, and the address is changing, we need to
	// delete the existing address -> name index entry. If there's an error getting
	// it, we don't really care; either it doesn't exist or the same error will
	// come up again later (when we add the new record).
	existing, _ := k.GetRecordByName(ctx, name)
	if existing != nil && existing.Address != addr.String() {
		var oldAddr sdk.AccAddress
		var oldNameKeyPre, oldAddrKey []byte
		oldAddr, err = sdk.AccAddressFromBech32(existing.Address)
		if err != nil {
			return types.ErrInvalidAddress.Wrapf("invalid existing %s record address: %v", name, err)
		}
		oldNameKeyPre, err = types.GetNameKeyPrefix(name)
		if err != nil {
			return err
		}
		oldAddrKey, err = types.GetAddressKeyPrefix(oldAddr)
		if err != nil {
			return types.ErrInvalidAddress.Wrapf("invalid existing %s record address format: %v", name, err)
		}
		oldAddrKey = append(oldAddrKey, oldNameKeyPre...)
		store := ctx.KVStore(k.storeKey)
		store.Delete(oldAddrKey)
	}

	if err = k.addRecord(ctx, name, addr, restrict, true); err != nil {
		return err
	}

	nameUpdateEvent := types.NewEventNameUpdate(addr.String(), name, restrict)

	return ctx.EventManager().EmitTypedEvent(nameUpdateEvent)
}

// GetRecordByName resolves a record by name.
func (k Keeper) GetRecordByName(ctx sdk.Context, name string) (record *types.NameRecord, err error) {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return nil, err
	}
	return getNameRecord(ctx, k, key)
}

func getNameRecord(ctx sdk.Context, keeper Keeper, key []byte) (record *types.NameRecord, err error) {
	store := ctx.KVStore(keeper.storeKey)
	if !store.Has(key) {
		return nil, types.ErrNameNotBound
	}
	bz := store.Get(key)
	record = &types.NameRecord{}
	err = keeper.cdc.Unmarshal(bz, record)
	return record, err
}

// NameExists returns true if store contains a record for the given name.
func (k Keeper) NameExists(ctx sdk.Context, name string) bool {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return false
	}
	store := ctx.KVStore(k.storeKey)
	return store.Has(key)
}

// GetRecordsByAddress looks up all names bound to an address.
func (k Keeper) GetRecordsByAddress(ctx sdk.Context, address sdk.AccAddress) (types.NameRecords, error) {
	// Return value data structure.
	records := types.NameRecords{}
	// Handler that adds records if account address matches.
	appendToRecords := func(record types.NameRecord) error {
		if record.Address == address.String() {
			records = append(records, record)
		}
		return nil
	}
	// Calculate address prefix
	addrPrefix, err := types.GetAddressKeyPrefix(address)
	if err != nil {
		return nil, err
	}
	// Collect and return all names that match.
	if err := k.IterateRecords(ctx, addrPrefix, appendToRecords); err != nil {
		return records, err
	}
	return records, nil
}

// DeleteRecord removes a name record from the kvstore.
func (k Keeper) DeleteRecord(ctx sdk.Context, name string) error {
	// Need the record to clear the address index
	record, err := k.GetRecordByName(ctx, name)
	if err != nil {
		return err
	}
	address, err := sdk.AccAddressFromBech32(record.Address)
	if err != nil {
		return err
	}
	// Delete the main name record
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return err
	}
	store := ctx.KVStore(k.storeKey)
	store.Delete(key)
	// Delete the address index record
	addrPrefix, err := types.GetAddressKeyPrefix(address)
	if err != nil {
		return err
	}
	addrPrefix = append(addrPrefix, key...) // [0x02] :: [addr-bytes] :: [name-key-bytes]
	if store.Has(addrPrefix) {
		store.Delete(addrPrefix)
	}

	nameUnboundEvent := types.NewEventNameUnbound(record.Address, name, record.Restricted)

	return ctx.EventManager().EmitTypedEvent(nameUnboundEvent)
}

// IterateRecords iterates over all the stored name records and passes them to a callback function.
func (k Keeper) IterateRecords(ctx sdk.Context, prefix []byte, handle func(record types.NameRecord) error) error {
	// Init a name record iterator
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, prefix)
	defer iterator.Close()
	// Iterate over records, processing callbacks.
	for ; iterator.Valid(); iterator.Next() {
		record := types.NameRecord{}
		if err := k.cdc.Unmarshal(iterator.Value(), &record); err != nil {
			return err
		}
		if err := handle(record); err != nil {
			return err
		}
	}
	return nil
}

// Normalize returns a name is storage format.
func (k Keeper) Normalize(ctx sdk.Context, name string) (string, error) {
	comps := make([]string, 0)
	for _, comp := range strings.Split(name, ".") {
		comp = strings.ToLower(strings.TrimSpace(comp))
		lenComp := uint32(len(comp))
		isUUID := isValidUUID(comp)
		if lenComp < k.GetMinSegmentLength(ctx) {
			return "", types.ErrNameSegmentTooShort
		}
		if lenComp > k.GetMaxSegmentLength(ctx) && !isUUID {
			return "", types.ErrNameSegmentTooLong
		}
		if !isValid(comp) {
			return "", types.ErrNameInvalid
		}
		comps = append(comps, comp)
	}
	if uint32(len(comps)) > k.GetMaxNameLevels(ctx) {
		return "", types.ErrNameHasTooManySegments
	}
	return strings.Join(comps, "."), nil
}

// Check whether a name component is valid
func isValid(s string) bool {
	// Allow valid UUID
	if isValidUUID(s) {
		return true
	}
	// Only allow a single dash if not a UUID
	if strings.Count(s, "-") > 1 {
		return false
	}
	for _, c := range s {
		if c == '-' {
			continue
		}
		if !unicode.IsGraphic(c) {
			return false
		}
		if !unicode.IsLower(c) && !unicode.IsDigit(c) {
			return false
		}
	}
	return true
}

// Ensure a string can be parsed into a UUID.
func isValidUUID(s string) bool {
	if _, err := uuid.Parse(s); err != nil {
		return false
	}
	return true
}

func (k Keeper) addRecord(ctx sdk.Context, name string, addr sdk.AccAddress, restrict, isModifiable bool) error {
	key, err := types.GetNameKeyPrefix(name)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	if store.Has(key) && !isModifiable {
		return types.ErrNameAlreadyBound
	}

	record := types.NewNameRecord(name, addr, restrict)
	if err = record.Validate(); err != nil {
		return err
	}
	bz, err := k.cdc.Marshal(&record)
	if err != nil {
		return err
	}
	store.Set(key, bz)
	// Now index by address
	addrPrefix, err := types.GetAddressKeyPrefix(addr)
	if err != nil {
		return err
	}
	addrPrefix = append(addrPrefix, key...) // [0x04] :: [addr-bytes] :: [name-key-bytes]
	store.Set(addrPrefix, bz)

	return nil
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

// DeleteInvalidAddressIndexEntries is only for the rust upgrade. It goes over all the address -> name entries and
// deletes any that are no longer accurate.
func (k Keeper) DeleteInvalidAddressIndexEntries(ctx sdk.Context) {
	logger := k.Logger(ctx)
	logger.Info("Checking address -> name index entries.")

	keepCount := 0
	var toDelete [][]byte

	extractNameKey := func(key []byte) []byte {
		// byte 1 is the type byte (0x05), it's ignored here.
		// The 2nd byte is the length of the address that immediately follows it.
		// The name key starts directly after the address, and is the rest of the key.
		addrLen := int(key[1])
		nameKeyStart := addrLen + 2
		return key[nameKeyStart:]
	}

	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.AddressKeyPrefix)
	defer func() {
		if iter != nil {
			iter.Close()
		}
	}()

	for ; iter.Valid(); iter.Next() {
		// If the key points to a non-existent name, delete it.
		key := iter.Key()
		nameKey := extractNameKey(key)
		if !store.Has(nameKey) {
			toDelete = append(toDelete, key)
			continue
		}

		// If the index value and main value are different, delete the index.
		indValBz := iter.Value()
		mainValBz := store.Get(nameKey)
		if !bytes.Equal(indValBz, mainValBz) {
			toDelete = append(toDelete, key)
			continue
		}

		keepCount++
	}

	iter.Close()
	iter = nil

	if len(toDelete) == 0 {
		logger.Info(fmt.Sprintf("Done checking address -> name index entries. All %d entries are valid", keepCount))
		return
	}

	logger.Info(fmt.Sprintf("Found %d invalid address -> name index entries. Deleting them now.", len(toDelete)))

	for _, key := range toDelete {
		store.Delete(key)
	}

	logger.Info(fmt.Sprintf("Done checking address -> name index entries. Deleted %d invalid entries and kept %d valid entries.", len(toDelete), keepCount))
}
