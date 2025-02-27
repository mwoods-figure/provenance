package app

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/gorilla/mux"
	"github.com/rakyll/statik/fs"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/streaming"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	quarantinekeeper "github.com/cosmos/cosmos-sdk/x/quarantine/keeper"
	quarantinemodule "github.com/cosmos/cosmos-sdk/x/quarantine/module"
	"github.com/cosmos/cosmos-sdk/x/sanction"
	sanctionkeeper "github.com/cosmos/cosmos-sdk/x/sanction/keeper"
	sanctionmodule "github.com/cosmos/cosmos-sdk/x/sanction/module"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ica "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v6/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v6/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v6/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v6/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
	ibctestingtypes "github.com/cosmos/ibc-go/v6/testing/types"

	appparams "github.com/provenance-io/provenance/app/params"
	"github.com/provenance-io/provenance/internal/antewrapper"
	piohandlers "github.com/provenance-io/provenance/internal/handlers"
	"github.com/provenance-io/provenance/internal/pioconfig"
	"github.com/provenance-io/provenance/internal/provwasm"
	"github.com/provenance-io/provenance/internal/statesync"
	"github.com/provenance-io/provenance/x/attribute"
	attributekeeper "github.com/provenance-io/provenance/x/attribute/keeper"
	attributetypes "github.com/provenance-io/provenance/x/attribute/types"
	attributewasm "github.com/provenance-io/provenance/x/attribute/wasm"
	ibchooks "github.com/provenance-io/provenance/x/ibchooks"
	ibchookskeeper "github.com/provenance-io/provenance/x/ibchooks/keeper"
	ibchookstypes "github.com/provenance-io/provenance/x/ibchooks/types"
	"github.com/provenance-io/provenance/x/marker"
	markerkeeper "github.com/provenance-io/provenance/x/marker/keeper"
	markertypes "github.com/provenance-io/provenance/x/marker/types"
	markerwasm "github.com/provenance-io/provenance/x/marker/wasm"
	"github.com/provenance-io/provenance/x/metadata"
	metadatakeeper "github.com/provenance-io/provenance/x/metadata/keeper"
	metadatatypes "github.com/provenance-io/provenance/x/metadata/types"
	metadatawasm "github.com/provenance-io/provenance/x/metadata/wasm"
	"github.com/provenance-io/provenance/x/msgfees"
	msgfeeskeeper "github.com/provenance-io/provenance/x/msgfees/keeper"
	msgfeesmodule "github.com/provenance-io/provenance/x/msgfees/module"
	msgfeestypes "github.com/provenance-io/provenance/x/msgfees/types"
	msgfeeswasm "github.com/provenance-io/provenance/x/msgfees/wasm"
	"github.com/provenance-io/provenance/x/name"
	nameclient "github.com/provenance-io/provenance/x/name/client"
	namekeeper "github.com/provenance-io/provenance/x/name/keeper"
	nametypes "github.com/provenance-io/provenance/x/name/types"
	namewasm "github.com/provenance-io/provenance/x/name/wasm"
	rewardkeeper "github.com/provenance-io/provenance/x/reward/keeper"
	rewardmodule "github.com/provenance-io/provenance/x/reward/module"
	rewardtypes "github.com/provenance-io/provenance/x/reward/types"
	triggerkeeper "github.com/provenance-io/provenance/x/trigger/keeper"
	triggermodule "github.com/provenance-io/provenance/x/trigger/module"
	triggertypes "github.com/provenance-io/provenance/x/trigger/types"

	_ "github.com/provenance-io/provenance/client/docs/statik" // registers swagger-ui files with statik
)

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// DefaultPowerReduction pio specific value for power reduction for TokensFromConsensusPower
	DefaultPowerReduction = sdk.NewIntFromUint64(1000000000)

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		staking.AppModuleBasic{},
		mint.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(append(
			wasmclient.ProposalHandlers,
			paramsclient.ProposalHandler,
			distrclient.ProposalHandler,
			upgradeclient.LegacyProposalHandler,
			upgradeclient.LegacyCancelProposalHandler,
			ibcclientclient.UpdateClientProposalHandler,
			ibcclientclient.UpgradeProposalHandler,
			nameclient.RootNameProposalHandler,
		),
		),
		params.AppModuleBasic{},
		crisis.AppModuleBasic{},
		slashing.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		vesting.AppModuleBasic{},
		quarantinemodule.AppModuleBasic{},
		sanctionmodule.AppModuleBasic{},

		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		ibchooks.AppModuleBasic{},

		marker.AppModuleBasic{},
		attribute.AppModuleBasic{},
		name.AppModuleBasic{},
		metadata.AppModuleBasic{},
		wasm.AppModuleBasic{},
		msgfeesmodule.AppModuleBasic{},
		rewardmodule.AppModuleBasic{},
		triggermodule.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},

		icatypes.ModuleName:         nil,
		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		ibchookstypes.ModuleName:    nil,

		attributetypes.ModuleName: nil,
		markertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		wasm.ModuleName:           {authtypes.Burner},
		rewardtypes.ModuleName:    nil,
		triggertypes.ModuleName:   nil,
	}
)

var (
	_ CosmosApp               = (*App)(nil)
	_ servertypes.Application = (*App)(nil)
)

// WasmWrapper allows us to use namespacing in the config file
// This is only used for parsing in the app, x/wasm expects WasmConfig
type WasmWrapper struct {
	Wasm wasm.Config `mapstructure:"wasm"`
}

// SdkCoinDenomRegex returns a new sdk base denom regex string
func SdkCoinDenomRegex() string {
	return pioconfig.DefaultReDnmString
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.Keeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     crisiskeeper.Keeper
	UpgradeKeeper    upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	AuthzKeeper      authzkeeper.Keeper
	GroupKeeper      groupkeeper.Keeper
	EvidenceKeeper   evidencekeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	MsgFeesKeeper    msgfeeskeeper.Keeper
	RewardKeeper     rewardkeeper.Keeper
	QuarantineKeeper quarantinekeeper.Keeper
	SanctionKeeper   sanctionkeeper.Keeper
	TriggerKeeper    triggerkeeper.Keeper

	IBCKeeper      *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	IBCHooksKeeper *ibchookskeeper.Keeper
	ICAHostKeeper  *icahostkeeper.Keeper
	TransferKeeper *ibctransferkeeper.Keeper

	MarkerKeeper    markerkeeper.Keeper
	MetadataKeeper  metadatakeeper.Keeper
	AttributeKeeper attributekeeper.Keeper
	NameKeeper      namekeeper.Keeper
	WasmKeeper      *wasm.Keeper
	ContractKeeper  *wasmkeeper.PermissionedKeeper

	// make scoped keepers public for test purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedICAHostKeeper  capabilitykeeper.ScopedKeeper

	TransferStack    *ibchooks.IBCMiddleware
	Ics20WasmHooks   *ibchooks.WasmHooks
	HooksICS4Wrapper ibchooks.ICS4Middleware

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager

	// module configurator
	configurator module.Configurator
}

func init() {
	DefaultNodeHome = os.ExpandEnv("$PIO_HOME")

	if strings.TrimSpace(DefaultNodeHome) == "" {
		configDir, err := os.UserConfigDir()
		if err != nil {
			panic(err)
		}
		DefaultNodeHome = filepath.Join(configDir, "Provenance")
	}

	// 614,400 = 600 * 1024 = our wasm params maxWasmCodeSize value before it was removed in wasmd v0.27.
	wasmtypes.MaxWasmSize = 614_400
}

// New returns a reference to an initialized Provenance Blockchain App.
func New(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions, baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp("provenanced", logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetMsgServiceRouter(piohandlers.NewPioMsgServiceRouter(encodingConfig.TxConfig.TxDecoder()))
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)
	sdk.SetCoinDenomRegex(SdkCoinDenomRegex)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey,
		authzkeeper.StoreKey, group.StoreKey,

		ibchost.StoreKey,
		ibctransfertypes.StoreKey,
		icahosttypes.StoreKey,
		ibchookstypes.StoreKey,

		metadatatypes.StoreKey,
		markertypes.StoreKey,
		attributetypes.StoreKey,
		nametypes.StoreKey,
		msgfeestypes.StoreKey,
		wasm.StoreKey,
		rewardtypes.StoreKey,
		quarantine.StoreKey,
		sanction.StoreKey,
		triggertypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &App{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// Register State listening services.
	app.RegisterStreamingServices(appOpts)

	// Register helpers for state-sync status.
	statesync.RegisterSyncStatus()

	app.ParamsKeeper = initParamsKeeper(appCodec, legacyAmino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	// set the BaseApp's parameter store
	bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()))

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(appCodec, keys[capabilitytypes.StoreKey], memKeys[capabilitytypes.MemStoreKey])

	// grant capabilities for the ibc and ibc-transfer modules
	scopedIBCKeeper := app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	scopedTransferKeeper := app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedWasmKeeper := app.CapabilityKeeper.ScopeToModule(wasm.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)

	// capability keeper must be sealed after scope to module registrations are completed.
	app.CapabilityKeeper.Seal()

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec, keys[authtypes.StoreKey], app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, AccountAddressPrefix,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec, keys[banktypes.StoreKey], app.AccountKeeper, app.GetSubspace(banktypes.ModuleName), app.ModuleAccountAddrs(),
	)
	stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName),
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), &stakingKeeper,
		app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, authtypes.FeeCollectorName,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], &stakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights, keys[upgradetypes.StoreKey], appCodec, homePath, app.BaseApp,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.MsgFeesKeeper = msgfeeskeeper.NewKeeper(
		appCodec, keys[msgfeestypes.StoreKey], app.GetSubspace(msgfeestypes.ModuleName), authtypes.FeeCollectorName, pioconfig.GetProvenanceConfig().FeeDenom, app.Simulate, encodingConfig.TxConfig.TxDecoder(), interfaceRegistry)

	pioMsgFeesRouter := app.MsgServiceRouter().(*piohandlers.PioMsgServiceRouter)
	pioMsgFeesRouter.SetMsgFeesKeeper(app.MsgFeesKeeper)

	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	restrictHooks := piohandlers.NewStakingRestrictionHooks(&app.StakingKeeper, *piohandlers.DefaultRestrictionOptions)
	app.StakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(restrictHooks, app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	app.RewardKeeper = rewardkeeper.NewKeeper(appCodec, keys[rewardtypes.StoreKey], app.StakingKeeper, &app.GovKeeper, app.BankKeeper, app.AccountKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey], appCodec, app.BaseApp.MsgServiceRouter(), app.AccountKeeper,
	)

	app.GroupKeeper = groupkeeper.NewKeeper(keys[group.StoreKey], appCodec, app.BaseApp.MsgServiceRouter(), app.AccountKeeper, group.DefaultConfig())

	// Create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec, keys[ibchost.StoreKey], app.GetSubspace(ibchost.ModuleName), app.StakingKeeper, app.UpgradeKeeper, scopedIBCKeeper,
	)

	// Configure the hooks keeper
	hooksKeeper := ibchookskeeper.NewKeeper(
		keys[ibchookstypes.StoreKey],
		app.GetSubspace(ibchookstypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		nil,
	)
	app.IBCHooksKeeper = &hooksKeeper

	// Setup the ICS4Wrapper used by the hooks middleware
	addrPrefix := sdk.GetConfig().GetBech32AccountAddrPrefix()        // We use this approach so running tests which use "cosmos" will work while we use "pb"
	wasmHooks := ibchooks.NewWasmHooks(&hooksKeeper, nil, addrPrefix) // The contract keeper needs to be set later
	app.Ics20WasmHooks = &wasmHooks
	app.HooksICS4Wrapper = ibchooks.NewICS4Middleware(
		app.IBCKeeper.ChannelKeeper,
		app.Ics20WasmHooks,
	)

	// Create Transfer Keepers
	transferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		app.HooksICS4Wrapper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		scopedTransferKeeper,
	)
	app.TransferKeeper = &transferKeeper
	transferModule := ibctransfer.NewIBCModule(*app.TransferKeeper)
	hooksTransferModule := ibchooks.NewIBCMiddleware(transferModule, &app.HooksICS4Wrapper)
	app.TransferStack = &hooksTransferModule

	app.NameKeeper = namekeeper.NewKeeper(
		appCodec, keys[nametypes.StoreKey], app.GetSubspace(nametypes.ModuleName),
	)

	app.AttributeKeeper = attributekeeper.NewKeeper(
		appCodec, keys[attributetypes.StoreKey], app.GetSubspace(attributetypes.ModuleName), app.AccountKeeper, &app.NameKeeper,
	)

	app.MetadataKeeper = metadatakeeper.NewKeeper(
		appCodec, keys[metadatatypes.StoreKey], app.GetSubspace(metadatatypes.ModuleName), app.AccountKeeper, app.AuthzKeeper, app.AttributeKeeper,
	)

	markerReqAttrBypassAddrs := []sdk.AccAddress{
		authtypes.NewModuleAddress(authtypes.FeeCollectorName),     // Allow collecting fees in restricted coins.
		authtypes.NewModuleAddress(rewardtypes.ModuleName),         // Allow rewards to hold onto restricted coins.
		authtypes.NewModuleAddress(quarantine.ModuleName),          // Allow quarantine to hold onto restricted coins.
		authtypes.NewModuleAddress(govtypes.ModuleName),            // Allow restricted coins in deposits.
		authtypes.NewModuleAddress(distrtypes.ModuleName),          // Allow fee denoms to be restricted coins.
		authtypes.NewModuleAddress(stakingtypes.BondedPoolName),    // Allow bond denom to be a restricted coin.
		authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName), // Allow bond denom to be a restricted coin.
	}
	app.MarkerKeeper = markerkeeper.NewKeeper(
		appCodec, keys[markertypes.StoreKey], app.GetSubspace(markertypes.ModuleName),
		app.AccountKeeper, app.BankKeeper, app.AuthzKeeper, app.FeeGrantKeeper,
		app.AttributeKeeper, app.NameKeeper, app.TransferKeeper, markerReqAttrBypassAddrs,
	)

	pioMessageRouter := MessageRouterFunc(func(msg sdk.Msg) baseapp.MsgServiceHandler {
		return pioMsgFeesRouter.Handler(msg)
	})
	app.TriggerKeeper = triggerkeeper.NewKeeper(appCodec, keys[triggertypes.StoreKey], app.MsgServiceRouter())
	icaHostKeeper := icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey], app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, scopedICAHostKeeper, pioMessageRouter,
	)
	app.ICAHostKeeper = &icaHostKeeper
	icaModule := ica.NewAppModule(nil, app.ICAHostKeeper)
	icaHostIBCModule := icahost.NewIBCModule(*app.ICAHostKeeper)

	// Init CosmWasm module
	wasmDir := filepath.Join(homePath, "data", "wasm")

	wasmWrap := WasmWrapper{Wasm: wasm.DefaultWasmConfig()}
	err := viper.Unmarshal(&wasmWrap)
	if err != nil {
		panic("error while reading wasm config: " + err.Error())
	}
	wasmConfig := wasmWrap.Wasm

	// Init CosmWasm encoder integrations
	encoderRegistry := provwasm.NewEncoderRegistry()
	encoderRegistry.RegisterEncoder(nametypes.RouterKey, namewasm.Encoder)
	encoderRegistry.RegisterEncoder(attributetypes.RouterKey, attributewasm.Encoder)
	encoderRegistry.RegisterEncoder(markertypes.RouterKey, markerwasm.Encoder)
	encoderRegistry.RegisterEncoder(metadatatypes.RouterKey, metadatawasm.Encoder)
	encoderRegistry.RegisterEncoder(msgfeestypes.RouterKey, msgfeeswasm.Encoder)

	// Init CosmWasm query integrations
	querierRegistry := provwasm.NewQuerierRegistry()
	querierRegistry.RegisterQuerier(nametypes.RouterKey, namewasm.Querier(app.NameKeeper))
	querierRegistry.RegisterQuerier(attributetypes.RouterKey, attributewasm.Querier(app.AttributeKeeper))
	querierRegistry.RegisterQuerier(markertypes.RouterKey, markerwasm.Querier(app.MarkerKeeper))
	querierRegistry.RegisterQuerier(metadatatypes.RouterKey, metadatawasm.Querier(app.MetadataKeeper))

	// Add the staking feature and indicate that provwasm contracts can be run on this chain.
	// Addition of cosmwasm_1_1 adds capability defined here: https://github.com/CosmWasm/cosmwasm/pull/1356
	supportedFeatures := "staking,provenance,stargate,iterator,cosmwasm_1_1"

	// The last arguments contain custom message handlers, and custom query handlers,
	// to allow smart contracts to use provenance modules.
	wasmKeeperInstance := wasm.NewKeeper(
		appCodec,
		keys[wasm.StoreKey],
		app.GetSubspace(wasm.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.DistrKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		scopedWasmKeeper,
		app.TransferKeeper,
		pioMessageRouter,
		app.GRPCQueryRouter(),
		wasmDir,
		wasmConfig,
		supportedFeatures,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		wasmkeeper.WithQueryPlugins(provwasm.QueryPlugins(querierRegistry, *app.GRPCQueryRouter(), appCodec)),
		wasmkeeper.WithMessageEncoders(provwasm.MessageEncoders(encoderRegistry, logger)),
	)
	app.WasmKeeper = &wasmKeeperInstance

	// Pass the wasm keeper to all the wrappers that need it
	app.ContractKeeper = wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper)
	app.Ics20WasmHooks.ContractKeeper = app.WasmKeeper // app.ContractKeeper -- this changes in the next version of wasm to a permissioned keeper
	app.IBCHooksKeeper.ContractKeeper = app.ContractKeeper

	unsanctionableAddrs := make([]sdk.AccAddress, 0, len(maccPerms)+1)
	for mName := range maccPerms {
		unsanctionableAddrs = append(unsanctionableAddrs, authtypes.NewModuleAddress(mName))
	}
	unsanctionableAddrs = append(unsanctionableAddrs, authtypes.NewModuleAddress(quarantine.ModuleName))
	app.SanctionKeeper = sanctionkeeper.NewKeeper(appCodec, keys[sanction.StoreKey],
		app.BankKeeper, &app.GovKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(), unsanctionableAddrs)

	// register the proposal types
	govRouter := govtypesv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govtypesv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.WasmKeeper, wasm.EnableAllProposals)).
		AddRoute(nametypes.ModuleName, name.NewProposalHandler(app.NameKeeper)).
		AddRoute(markertypes.ModuleName, marker.NewProposalHandler(app.MarkerKeeper)).
		AddRoute(msgfeestypes.ModuleName, msgfees.NewProposalHandler(app.MsgFeesKeeper, app.InterfaceRegistry()))
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, govRouter, app.BaseApp.MsgServiceRouter(), govtypes.Config{MaxMetadataLen: 10000},
	)
	app.GovKeeper.SetHooks(govtypes.NewMultiGovHooks(app.SanctionKeeper))

	// Create static IBC router, add transfer route, then set and seal it
	ibcRouter := porttypes.NewRouter()
	ibcRouter.
		AddRoute(ibctransfertypes.ModuleName, app.TransferStack).
		AddRoute(wasm.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCKeeper.ChannelKeeper)).
		AddRoute(icahosttypes.SubModuleName, icaHostIBCModule)
	app.IBCKeeper.SetRouter(ibcRouter)

	// Create evidence Keeper for to register the IBC light client misbehavior evidence route
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec, keys[evidencetypes.StoreKey], &app.StakingKeeper, app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	app.QuarantineKeeper = quarantinekeeper.NewKeeper(appCodec, keys[quarantine.StoreKey], app.BankKeeper, authtypes.NewModuleAddress(quarantine.ModuleName))

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	var skipGenesisInvariants = cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		quarantinemodule.NewAppModule(appCodec, app.QuarantineKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		sanctionmodule.NewAppModule(appCodec, app.SanctionKeeper, app.AccountKeeper, app.BankKeeper, app.GovKeeper, app.interfaceRegistry),

		// PROVENANCE
		metadata.NewAppModule(appCodec, app.MetadataKeeper, app.AccountKeeper),
		marker.NewAppModule(appCodec, app.MarkerKeeper, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.GovKeeper, app.AttributeKeeper, app.interfaceRegistry),
		name.NewAppModule(appCodec, app.NameKeeper, app.AccountKeeper, app.BankKeeper),
		attribute.NewAppModule(appCodec, app.AttributeKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),
		msgfeesmodule.NewAppModule(appCodec, app.MsgFeesKeeper, app.interfaceRegistry),
		wasm.NewAppModule(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		rewardmodule.NewAppModule(appCodec, app.RewardKeeper, app.AccountKeeper, app.BankKeeper),
		triggermodule.NewAppModule(appCodec, app.TriggerKeeper, app.AccountKeeper, app.BankKeeper),

		// IBC
		ibc.NewAppModule(app.IBCKeeper),
		ibchooks.NewAppModule(app.AccountKeeper, *app.IBCHooksKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		icaModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibchost.ModuleName,
		markertypes.ModuleName,
		icatypes.ModuleName,
		attributetypes.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,

		// no-ops
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		group.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		msgfeestypes.ModuleName,
		metadatatypes.ModuleName,
		wasm.ModuleName,
		ibchookstypes.ModuleName,
		ibctransfertypes.ModuleName,
		nametypes.ModuleName,
		vestingtypes.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		icatypes.ModuleName,
		group.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,

		// no-ops
		vestingtypes.ModuleName,
		distrtypes.ModuleName,
		authz.ModuleName,
		metadatatypes.ModuleName,
		nametypes.ModuleName,
		genutiltypes.ModuleName,
		ibchost.ModuleName,
		ibchookstypes.ModuleName,
		ibctransfertypes.ModuleName,
		msgfeestypes.ModuleName,
		wasm.ModuleName,
		slashingtypes.ModuleName,
		upgradetypes.ModuleName,
		attributetypes.ModuleName,
		capabilitytypes.ModuleName,
		evidencetypes.ModuleName,
		banktypes.ModuleName,
		minttypes.ModuleName,
		markertypes.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		markertypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		minttypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		group.ModuleName,
		feegrant.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,

		nametypes.ModuleName,
		attributetypes.ModuleName,
		metadatatypes.ModuleName,
		msgfeestypes.ModuleName,

		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
		icatypes.ModuleName,
		ibchookstypes.ModuleName,
		// wasm after ibc transfer
		wasm.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,

		// no-ops
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		upgradetypes.ModuleName,
	)

	app.mm.SetOrderMigrations(
		banktypes.ModuleName,
		authz.ModuleName,
		group.ModuleName,
		capabilitytypes.ModuleName,
		crisistypes.ModuleName,
		distrtypes.ModuleName,
		evidencetypes.ModuleName,
		feegrant.ModuleName,
		genutiltypes.ModuleName,
		govtypes.ModuleName,
		ibchost.ModuleName,
		minttypes.ModuleName,
		paramstypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		ibctransfertypes.ModuleName,
		upgradetypes.ModuleName,
		vestingtypes.ModuleName,
		quarantine.ModuleName,
		sanction.ModuleName,

		ibchookstypes.ModuleName,
		icatypes.ModuleName,
		wasm.ModuleName,

		attributetypes.ModuleName,
		markertypes.ModuleName,
		msgfeestypes.ModuleName,
		metadatatypes.ModuleName,
		nametypes.ModuleName,
		rewardtypes.ModuleName,
		triggertypes.ModuleName,

		// Last due to v0.44 issue: https://github.com/cosmos/cosmos-sdk/issues/10591
		authtypes.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.appCodec, app.BaseApp.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	app.sm = module.NewSimulationManager(
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		params.NewAppModule(app.ParamsKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		quarantinemodule.NewAppModule(appCodec, app.QuarantineKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		sanctionmodule.NewAppModule(appCodec, app.SanctionKeeper, app.AccountKeeper, app.BankKeeper, app.GovKeeper, app.interfaceRegistry),

		metadata.NewAppModule(appCodec, app.MetadataKeeper, app.AccountKeeper),
		marker.NewAppModule(appCodec, app.MarkerKeeper, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.GovKeeper, app.AttributeKeeper, app.interfaceRegistry),
		name.NewAppModule(appCodec, app.NameKeeper, app.AccountKeeper, app.BankKeeper),
		attribute.NewAppModule(appCodec, app.AttributeKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),
		msgfeesmodule.NewAppModule(appCodec, app.MsgFeesKeeper, app.interfaceRegistry),
		rewardmodule.NewAppModule(appCodec, app.RewardKeeper, app.AccountKeeper, app.BankKeeper),
		triggermodule.NewAppModule(appCodec, app.TriggerKeeper, app.AccountKeeper, app.BankKeeper),
		provwasm.NewWrapper(appCodec, app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.NameKeeper),

		// IBC
		ibc.NewAppModule(app.IBCKeeper),
		ibchooks.NewAppModule(app.AccountKeeper, *app.IBCHooksKeeper),
		ibctransfer.NewAppModule(*app.TransferKeeper),
		icaModule,
	)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	anteHandler, err := antewrapper.NewAnteHandler(
		antewrapper.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			MsgFeesKeeper:   app.MsgFeesKeeper,
			SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
		})
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
	msgfeehandler, err := piohandlers.NewAdditionalMsgFeeHandler(piohandlers.PioBaseAppKeeperOptions{
		AccountKeeper:  app.AccountKeeper,
		BankKeeper:     app.BankKeeper,
		FeegrantKeeper: app.FeeGrantKeeper,
		MsgFeesKeeper:  app.MsgFeesKeeper,
		Decoder:        encodingConfig.TxConfig.TxDecoder(),
	})

	if err != nil {
		panic(err)
	}
	app.SetFeeHandler(msgfeehandler)

	app.SetEndBlocker(app.EndBlocker)

	app.SetAggregateEventsFunc(piohandlers.AggregateEvents)

	// Add upgrade plans for each release. This must be done before the baseapp seals via LoadLatestVersion() down below.
	InstallCustomUpgradeHandlers(app)

	// Use the dump of $home/data/upgrade-info.json:{"name":"$plan","height":321654} to determine
	// if we load a store upgrade from the handlers. No file == no error from read func.
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(err)
	}

	// Currently in an upgrade hold for this block.
	if upgradeInfo.Name != "" && upgradeInfo.Height == app.LastBlockHeight()+1 {
		if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
			app.Logger().Info("Skipping upgrade based on height",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
		} else {
			app.Logger().Info("Managing upgrade",
				"plan", upgradeInfo.Name,
				"upgradeHeight", upgradeInfo.Height,
				"lastHeight", app.LastBlockHeight(),
			)
			// See if we have a custom store loader to use for upgrades.
			storeLoader := GetUpgradeStoreLoader(app, upgradeInfo)
			if storeLoader != nil {
				app.SetStoreLoader(storeLoader)
			}
		}
	}
	// --

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	app.ScopedIBCKeeper = scopedIBCKeeper
	app.ScopedTransferKeeper = scopedTransferKeeper
	app.ScopedICAHostKeeper = scopedICAHostKeeper

	return app
}

// GetBaseApp returns the base cosmos app
func (app *App) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetStakingKeeper returns the staking keeper (for ibc testing)
func (app *App) GetStakingKeeper() ibctestingtypes.StakingKeeper {
	return app.StakingKeeper
}

// GetIBCKeeper returns the ibc keeper (for ibc testing)
func (app *App) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper // This is a *ibckeeper.Keeper
}

// GetScopedIBCKeeper returns the scoped ibc keeper (for ibc testing)
func (app *App) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig implements the TestingApp interface (for ibc testing).
func (app *App) GetTxConfig() client.TxConfig {
	return MakeEncodingConfig().TxConfig
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Provenance's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Provenance's InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		RegisterSwaggerAPI(clientCtx, apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(clientCtx, app.BaseApp.GRPCQueryRouter(), app.interfaceRegistry, app.Query)
}

// RegisterNodeService registers the node query server.
func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// RegisterStreamingServices registers types.ABCIListener State Listening services with the App.
func (app *App) RegisterStreamingServices(appOpts servertypes.AppOptions) {
	// register streaming services
	streamingCfg := cast.ToStringMap(appOpts.Get(baseapp.StreamingTomlKey))
	for service := range streamingCfg {
		pluginKey := fmt.Sprintf("%s.%s.%s", baseapp.StreamingTomlKey, service, baseapp.StreamingABCIPluginTomlKey)
		pluginName := strings.TrimSpace(cast.ToString(appOpts.Get(pluginKey)))
		if len(pluginName) > 0 {
			logLevel := cast.ToString(appOpts.Get(flags.FlagLogLevel))
			plugin, err := streaming.NewStreamingPlugin(pluginName, logLevel)
			if err != nil {
				app.Logger().Error("failed to load streaming plugin", "error", err)
				os.Exit(1)
			}
			if err := baseapp.RegisterStreamingPlugin(app.BaseApp, appOpts, app.keys, plugin); err != nil {
				app.Logger().Error("failed to register streaming plugin", "error", err)
				os.Exit(1)
			}
			app.Logger().Info("streaming service registered", "service", service, "plugin", pluginName)
		}
	}
}

// RegisterSwaggerAPI registers swagger route with API Server
func RegisterSwaggerAPI(_ client.Context, rtr *mux.Router) {
	statikFS, err := fs.New()
	if err != nil {
		panic(err)
	}

	staticServer := http.FileServer(statikFS)
	rtr.PathPrefix("/swagger/").Handler(http.StripPrefix("/swagger/", staticServer))
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govtypesv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)

	paramsKeeper.Subspace(metadatatypes.ModuleName)
	paramsKeeper.Subspace(markertypes.ModuleName)
	paramsKeeper.Subspace(nametypes.ModuleName)
	paramsKeeper.Subspace(attributetypes.ModuleName)
	paramsKeeper.Subspace(msgfeestypes.ModuleName)
	paramsKeeper.Subspace(wasm.ModuleName)
	paramsKeeper.Subspace(rewardtypes.ModuleName)
	paramsKeeper.Subspace(triggertypes.ModuleName)

	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(ibchookstypes.ModuleName)

	return paramsKeeper
}

// injectUpgrade causes the named upgrade to be run as the chain starts.
//
// To use this, add a call to it in New after the call to InstallCustomUpgradeHandlers
// but before the line that looks for an upgrade file.
//
// This function is for testing an upgrade against an existing chain's data (e.g. mainnet).
// Here's how:
//  1. Run a node for the chain you want to test using its normal release.
//  2. In this provenance repo, check out the branch/version with the upgrade you want to test.
//  3. Add a call to this function as described above, e.g. injectUpgrade("ochre").
//  4. Compile it with `make build`.
//  5. I suggest renaming build/provenanced to something like provenanced-ochre-force-upgrade
//     and moving it somewhere handier for your node.
//  6. Stop your node.
//  7. Back up your data directory because we're about to mess it up.
//  8. Seriously, your data directory will need to be thrown away after this.
//  9. Restart your node with `--log_level debug` using your new force-upgrade binary.
//
// As the node starts, it should think that an upgrade is needed and attempt to execute it.
// If the upgrade finishes successfully, your node will then try and fail to sync with the rest of the nodes.
// Your chain now has a different state than the rest of the network and will be generating different hashes.
// There's no reason to let it continue to run.
//
// Deprecated:  This function should never be called in anything that gets merged into main or any sort of release branch.
// It's marked as deprecated so that things can complain about its use (e.g. the linter).
func (app *App) injectUpgrade(name string) { //nolint:unused // This is designed to only be used in unofficial code.
	plan := upgradetypes.Plan{
		Name:   name,
		Height: app.LastBlockHeight() + 1,
	}
	// Write the plan to $home/data/upgrade-info.json
	if err := app.UpgradeKeeper.DumpUpgradeInfoToDisk(plan.Height, plan); err != nil {
		panic(err)
	}
	// Define a new BeginBlocker that will inject the upgrade.
	injected := false
	app.SetBeginBlocker(func(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
		if !injected {
			app.Logger().Info("Injecting upgrade plan", "plan", plan)
			// Ideally, we'd just call ScheduleUpgrade(ctx, plan) here (and panic on error).
			// But the upgrade keeper has a migration in v0.46 that changes some store key stuff.
			// ScheduleUpgrade tries to read some of that changed store stuff and fails if the migration hasn't
			// been applied yet. So we're doing things the hard way here.
			app.UpgradeKeeper.ClearUpgradePlan(ctx)
			store := ctx.KVStore(app.GetKey(upgradetypes.StoreKey))
			bz := app.appCodec.MustMarshal(&plan)
			store.Set(upgradetypes.PlanKey(), bz)
			injected = true
		}
		return app.BeginBlocker(ctx, req)
	})
}
