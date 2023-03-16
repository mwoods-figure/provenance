package types

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
	yaml "gopkg.in/yaml.v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Message type URLs are generated by running unit test in msg_test.go => TestPrintMessageTypeStrings
const (
	TypeURLMsgWriteScopeRequest                      = "/provenance.metadata.v1.MsgWriteScopeRequest"
	TypeURLMsgDeleteScopeRequest                     = "/provenance.metadata.v1.MsgDeleteScopeRequest"
	TypeURLMsgAddScopeDataAccessRequest              = "/provenance.metadata.v1.MsgAddScopeDataAccessRequest"
	TypeURLMsgDeleteScopeDataAccessRequest           = "/provenance.metadata.v1.MsgDeleteScopeDataAccessRequest"
	TypeURLMsgAddScopeOwnerRequest                   = "/provenance.metadata.v1.MsgAddScopeOwnerRequest"
	TypeURLMsgDeleteScopeOwnerRequest                = "/provenance.metadata.v1.MsgDeleteScopeOwnerRequest"
	TypeURLMsgWriteSessionRequest                    = "/provenance.metadata.v1.MsgWriteSessionRequest"
	TypeURLMsgWriteRecordRequest                     = "/provenance.metadata.v1.MsgWriteRecordRequest"
	TypeURLMsgDeleteRecordRequest                    = "/provenance.metadata.v1.MsgDeleteRecordRequest"
	TypeURLMsgWriteScopeSpecificationRequest         = "/provenance.metadata.v1.MsgWriteScopeSpecificationRequest"
	TypeURLMsgDeleteScopeSpecificationRequest        = "/provenance.metadata.v1.MsgDeleteScopeSpecificationRequest"
	TypeURLMsgWriteContractSpecificationRequest      = "/provenance.metadata.v1.MsgWriteContractSpecificationRequest"
	TypeURLMsgDeleteContractSpecificationRequest     = "/provenance.metadata.v1.MsgDeleteContractSpecificationRequest"
	TypeURLMsgAddContractSpecToScopeSpecRequest      = "/provenance.metadata.v1.MsgAddContractSpecToScopeSpecRequest"
	TypeURLMsgDeleteContractSpecFromScopeSpecRequest = "/provenance.metadata.v1.MsgDeleteContractSpecFromScopeSpecRequest"
	TypeURLMsgWriteRecordSpecificationRequest        = "/provenance.metadata.v1.MsgWriteRecordSpecificationRequest"
	TypeURLMsgDeleteRecordSpecificationRequest       = "/provenance.metadata.v1.MsgDeleteRecordSpecificationRequest"
	TypeURLMsgBindOSLocatorRequest                   = "/provenance.metadata.v1.MsgBindOSLocatorRequest"
	TypeURLMsgDeleteOSLocatorRequest                 = "/provenance.metadata.v1.MsgDeleteOSLocatorRequest"
	TypeURLMsgModifyOSLocatorRequest                 = "/provenance.metadata.v1.MsgModifyOSLocatorRequest"
)

// Compile time interface checks.
var (
	_ sdk.Msg = &MsgWriteScopeRequest{}
	_ sdk.Msg = &MsgDeleteScopeRequest{}
	_ sdk.Msg = &MsgAddScopeDataAccessRequest{}
	_ sdk.Msg = &MsgDeleteScopeDataAccessRequest{}
	_ sdk.Msg = &MsgAddScopeOwnerRequest{}
	_ sdk.Msg = &MsgDeleteScopeOwnerRequest{}
	_ sdk.Msg = &MsgWriteSessionRequest{}
	_ sdk.Msg = &MsgWriteRecordRequest{}
	_ sdk.Msg = &MsgDeleteRecordRequest{}
	_ sdk.Msg = &MsgWriteScopeSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteScopeSpecificationRequest{}
	_ sdk.Msg = &MsgWriteContractSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteContractSpecificationRequest{}
	_ sdk.Msg = &MsgAddContractSpecToScopeSpecRequest{}
	_ sdk.Msg = &MsgDeleteContractSpecFromScopeSpecRequest{}
	_ sdk.Msg = &MsgWriteRecordSpecificationRequest{}
	_ sdk.Msg = &MsgDeleteRecordSpecificationRequest{}
	_ sdk.Msg = &MsgBindOSLocatorRequest{}
	_ sdk.Msg = &MsgDeleteOSLocatorRequest{}
	_ sdk.Msg = &MsgModifyOSLocatorRequest{}
)

// private method to convert an array of strings into an array of Acc Addresses.
func stringsToAccAddresses(strings []string) []sdk.AccAddress {
	retval := make([]sdk.AccAddress, len(strings))

	for i, str := range strings {
		retval[i] = MustAccAddressFromBech32(str)
	}

	return retval
}

// MustAccAddressFromBech32 converts a Bech32 address to sdk.AccAddress
// Panics on error
func MustAccAddressFromBech32(s string) sdk.AccAddress {
	accAddress, err := sdk.AccAddressFromBech32(s)
	if err != nil {
		panic(err)
	}
	return accAddress
}

// ------------------  MsgWriteScopeRequest  ------------------

// NewMsgWriteScopeRequest creates a new msg instance
func NewMsgWriteScopeRequest(scope Scope, signers []string) *MsgWriteScopeRequest {
	return &MsgWriteScopeRequest{
		Scope:   scope,
		Signers: signers,
	}
}

func (msg MsgWriteScopeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgWriteScopeRequest) MsgTypeURL() string {
	return TypeURLMsgWriteScopeRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteScopeRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteScopeRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if err := msg.ConvertOptionalFields(); err != nil {
		return err
	}
	return msg.Scope.ValidateBasic()
}

// ConvertOptionalFields will look at the ScopeUuid and SpecUuid fields in the message.
// For each, if present, it will be converted to a MetadataAddress and set in the Scope appropriately.
// Once used, those uuid fields will be set to empty strings so that calling this again has no effect.
func (msg *MsgWriteScopeRequest) ConvertOptionalFields() error {
	if len(msg.ScopeUuid) > 0 {
		uid, err := uuid.Parse(msg.ScopeUuid)
		if err != nil {
			return fmt.Errorf("invalid scope uuid: %w", err)
		}
		scopeAddr := ScopeMetadataAddress(uid)
		if !msg.Scope.ScopeId.Empty() && !msg.Scope.ScopeId.Equals(scopeAddr) {
			return fmt.Errorf("msg.Scope.ScopeId [%s] is different from the one created from msg.ScopeUuid [%s]",
				msg.Scope.ScopeId, msg.ScopeUuid)
		}
		msg.Scope.ScopeId = scopeAddr
		msg.ScopeUuid = ""
	}
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		specAddr := ScopeSpecMetadataAddress(uid)
		if !msg.Scope.SpecificationId.Empty() && !msg.Scope.SpecificationId.Equals(specAddr) {
			return fmt.Errorf("msg.Scope.SpecificationId [%s] is different from the one created from msg.SpecUuid [%s]",
				msg.Scope.SpecificationId, msg.SpecUuid)
		}
		msg.Scope.SpecificationId = specAddr
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  NewMsgDeleteScopeRequest  ------------------

// NewMsgDeleteScopeRequest creates a new msg instance
func NewMsgDeleteScopeRequest(scopeID MetadataAddress, signers []string) *MsgDeleteScopeRequest {
	return &MsgDeleteScopeRequest{
		ScopeId: scopeID,
		Signers: signers,
	}
}

func (msg MsgDeleteScopeRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteScopeRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteScopeRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteScopeRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteScopeRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if !msg.ScopeId.IsScopeAddress() {
		return fmt.Errorf("invalid scope address")
	}
	return nil
}

// ------------------  MsgAddScopeDataAccessRequest  ------------------

// NewMsgAddScopeDataAccessRequest creates a new msg instance
func NewMsgAddScopeDataAccessRequest(scopeID MetadataAddress, dataAccessAddrs []string, signers []string) *MsgAddScopeDataAccessRequest {
	return &MsgAddScopeDataAccessRequest{
		ScopeId:    scopeID,
		DataAccess: dataAccessAddrs,
		Signers:    signers,
	}
}

func (msg MsgAddScopeDataAccessRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgAddScopeDataAccessRequest) MsgTypeURL() string {
	return TypeURLMsgAddScopeDataAccessRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddScopeDataAccessRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgAddScopeDataAccessRequest) ValidateBasic() error {
	if !msg.ScopeId.IsScopeAddress() {
		return fmt.Errorf("address is not a scope id: %v", msg.ScopeId.String())
	}
	if len(msg.DataAccess) < 1 {
		return fmt.Errorf("at least one data access address is required")
	}
	for _, da := range msg.DataAccess {
		_, err := sdk.AccAddressFromBech32(da)
		if err != nil {
			return fmt.Errorf("data access address is invalid: %s", da)
		}
	}
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgDeleteScopeDataAccessRequest  ------------------

// NewMsgDeleteScopeDataAccessRequest creates a new msg instance
func NewMsgDeleteScopeDataAccessRequest(scopeID MetadataAddress, dataAccessAddrs []string, signers []string) *MsgDeleteScopeDataAccessRequest {
	return &MsgDeleteScopeDataAccessRequest{
		ScopeId:    scopeID,
		DataAccess: dataAccessAddrs,
		Signers:    signers,
	}
}

func (msg MsgDeleteScopeDataAccessRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteScopeDataAccessRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteScopeDataAccessRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteScopeDataAccessRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteScopeDataAccessRequest) ValidateBasic() error {
	if !msg.ScopeId.IsScopeAddress() {
		return fmt.Errorf("address is not a scope id: %v", msg.ScopeId.String())
	}
	if len(msg.DataAccess) < 1 {
		return fmt.Errorf("at least one data access address is required")
	}
	for _, da := range msg.DataAccess {
		_, err := sdk.AccAddressFromBech32(da)
		if err != nil {
			return fmt.Errorf("data access address is invalid: %s", da)
		}
	}
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgAddScopeOwnerRequest  ------------------

// NewMsgAddScopeOwnerRequest creates a new msg instance
func NewMsgAddScopeOwnerRequest(scopeID MetadataAddress, owners []Party, signers []string) *MsgAddScopeOwnerRequest {
	return &MsgAddScopeOwnerRequest{
		ScopeId: scopeID,
		Owners:  owners,
		Signers: signers,
	}
}

func (msg MsgAddScopeOwnerRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgAddScopeOwnerRequest) MsgTypeURL() string {
	return TypeURLMsgAddScopeOwnerRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddScopeOwnerRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgAddScopeOwnerRequest) ValidateBasic() error {
	if !msg.ScopeId.IsScopeAddress() {
		return fmt.Errorf("address is not a scope id: %v", msg.ScopeId.String())
	}
	if err := ValidatePartiesBasic(msg.Owners); err != nil {
		return fmt.Errorf("invalid owners: %w", err)
	}
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgDeleteScopeOwnerRequest  ------------------

// NewMsgDeleteScopeOwnerRequest creates a new msg instance
func NewMsgDeleteScopeOwnerRequest(scopeID MetadataAddress, owners []string, signers []string) *MsgDeleteScopeOwnerRequest {
	return &MsgDeleteScopeOwnerRequest{
		ScopeId: scopeID,
		Owners:  owners,
		Signers: signers,
	}
}

func (msg MsgDeleteScopeOwnerRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteScopeOwnerRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteScopeOwnerRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteScopeOwnerRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteScopeOwnerRequest) ValidateBasic() error {
	if !msg.ScopeId.IsScopeAddress() {
		return fmt.Errorf("address is not a scope id: %v", msg.ScopeId.String())
	}
	if len(msg.Owners) < 1 {
		return fmt.Errorf("at least one owner address is required")
	}
	for _, owner := range msg.Owners {
		_, err := sdk.AccAddressFromBech32(owner)
		if err != nil {
			return fmt.Errorf("owner address is invalid: %s", owner)
		}
	}
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgWriteSessionRequest  ------------------

// NewMsgWriteSessionRequest creates a new msg instance
func NewMsgWriteSessionRequest(session Session, signers []string) *MsgWriteSessionRequest {
	return &MsgWriteSessionRequest{Session: session, Signers: signers}
}

func (msg MsgWriteSessionRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgWriteSessionRequest) MsgTypeURL() string {
	return TypeURLMsgWriteSessionRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteSessionRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteSessionRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if err := msg.ConvertOptionalFields(); err != nil {
		return err
	}
	return msg.Session.ValidateBasic()
}

// ConvertOptionalFields will look at the SessionIdComponents and SpecUuid fields in the message.
// For each, if present, it will be converted to a MetadataAddress and set in the Session appropriately.
// Once used, those fields will be emptied so that calling this again has no effect.
func (msg *MsgWriteSessionRequest) ConvertOptionalFields() error {
	if msg.SessionIdComponents != nil {
		sessionAddr, err := msg.SessionIdComponents.GetSessionAddr()
		if err != nil {
			return fmt.Errorf("invalid session id components: %w", err)
		}
		if sessionAddr != nil {
			if !msg.Session.SessionId.Empty() && !msg.Session.SessionId.Equals(*sessionAddr) {
				return fmt.Errorf("msg.Session.SessionId [%s] is different from the one created from msg.SessionIdComponents %v",
					msg.Session.SessionId, msg.SessionIdComponents)
			}
			msg.Session.SessionId = *sessionAddr
		}
		msg.SessionIdComponents = nil
	}
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		specAddr := ContractSpecMetadataAddress(uid)
		if !msg.Session.SpecificationId.Empty() && !msg.Session.SpecificationId.Equals(specAddr) {
			return fmt.Errorf("msg.Session.SpecificationId [%s] is different from the one created from msg.SpecUuid [%s]",
				msg.Session.SpecificationId, msg.SpecUuid)
		}
		msg.Session.SpecificationId = specAddr
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  MsgWriteRecordRequest  ------------------

// NewMsgWriteRecordRequest creates a new msg instance
func NewMsgWriteRecordRequest(record Record, sessionIDComponents *SessionIdComponents, contractSpecUUID string, signers []string, parties []Party) *MsgWriteRecordRequest {
	return &MsgWriteRecordRequest{Record: record, Parties: parties, Signers: signers, SessionIdComponents: sessionIDComponents, ContractSpecUuid: contractSpecUUID}
}

func (msg MsgWriteRecordRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgWriteRecordRequest) MsgTypeURL() string {
	return TypeURLMsgWriteRecordRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteRecordRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteRecordRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if err := msg.ConvertOptionalFields(); err != nil {
		return err
	}
	return msg.Record.ValidateBasic()
}

// ConvertOptionalFields will look at the SessionIdComponents and ContractSpecUuid fields in the message.
// For each, if present, it will be converted to a MetadataAddress and set in the Record appropriately.
// Once used, those fields will be emptied so that calling this again has no effect.
func (msg *MsgWriteRecordRequest) ConvertOptionalFields() error {
	if msg.SessionIdComponents != nil {
		sessionAddr, err := msg.SessionIdComponents.GetSessionAddr()
		if err != nil {
			return fmt.Errorf("invalid session id components: %w", err)
		}
		if sessionAddr != nil {
			if !msg.Record.SessionId.Empty() && !msg.Record.SessionId.Equals(*sessionAddr) {
				return fmt.Errorf("msg.Record.SessionId [%s] is different from the one created from msg.SessionIdComponents %v",
					msg.Record.SessionId, msg.SessionIdComponents)
			}
			msg.Record.SessionId = *sessionAddr
			msg.SessionIdComponents = nil
		}
	}
	if len(msg.ContractSpecUuid) > 0 {
		uid, err := uuid.Parse(msg.ContractSpecUuid)
		if err != nil {
			return fmt.Errorf("invalid contract spec uuid: %w", err)
		}
		if len(strings.TrimSpace(msg.Record.Name)) == 0 {
			return errors.New("empty record name")
		}
		specAddr := RecordSpecMetadataAddress(uid, msg.Record.Name)
		if !msg.Record.SpecificationId.Empty() && !msg.Record.SpecificationId.Equals(specAddr) {
			return fmt.Errorf("msg.Record.SpecificationId [%s] is different from the one created from msg.ContractSpecUuid [%s] and msg.Record.Name [%s]",
				msg.Record.SpecificationId, msg.ContractSpecUuid, msg.Record.Name)
		}
		msg.Record.SpecificationId = specAddr
		msg.ContractSpecUuid = ""
	}
	return nil
}

// ------------------  MsgDeleteRecordRequest  ------------------

// NewMsgDeleteScopeSpecificationRequest creates a new msg instance
func NewMsgDeleteRecordRequest(recordID MetadataAddress, signers []string) *MsgDeleteRecordRequest {
	return &MsgDeleteRecordRequest{RecordId: recordID, Signers: signers}
}

func (msg MsgDeleteRecordRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteRecordRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteRecordRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteRecordRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteRecordRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgWriteScopeSpecificationRequest  ------------------

// NewMsgAddScopeSpecificationRequest creates a new msg instance
func NewMsgWriteScopeSpecificationRequest(specification ScopeSpecification, signers []string) *MsgWriteScopeSpecificationRequest {
	return &MsgWriteScopeSpecificationRequest{Specification: specification, Signers: signers}
}

func (msg MsgWriteScopeSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgWriteScopeSpecificationRequest) MsgTypeURL() string {
	return TypeURLMsgWriteScopeSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteScopeSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteScopeSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if err := msg.ConvertOptionalFields(); err != nil {
		return err
	}
	return msg.Specification.ValidateBasic()
}

// ConvertOptionalFields will look at the SpecUuid field in the message.
// If present, it will be converted to a MetadataAddress and set in the Specification appropriately.
// Once used, it will be emptied so that calling this again has no effect.
func (msg *MsgWriteScopeSpecificationRequest) ConvertOptionalFields() error {
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		specAddr := ScopeSpecMetadataAddress(uid)
		if !msg.Specification.SpecificationId.Empty() && !msg.Specification.SpecificationId.Equals(specAddr) {
			return fmt.Errorf("msg.Specification.SpecificationId [%s] is different from the one created from msg.SpecUuid [%s]",
				msg.Specification.SpecificationId, msg.SpecUuid)
		}
		msg.Specification.SpecificationId = specAddr
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  MsgDeleteScopeSpecificationRequest  ------------------

// NewMsgDeleteScopeSpecificationRequest creates a new msg instance
func NewMsgDeleteScopeSpecificationRequest(specificationID MetadataAddress, signers []string) *MsgDeleteScopeSpecificationRequest {
	return &MsgDeleteScopeSpecificationRequest{SpecificationId: specificationID, Signers: signers}
}

func (msg MsgDeleteScopeSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteScopeSpecificationRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteScopeSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteScopeSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteScopeSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgWriteContractSpecificationRequest  ------------------

// NewMsgWriteContractSpecificationRequest creates a new msg instance
func NewMsgWriteContractSpecificationRequest(specification ContractSpecification, signers []string) *MsgWriteContractSpecificationRequest {
	return &MsgWriteContractSpecificationRequest{Specification: specification, Signers: signers}
}

func (msg MsgWriteContractSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgWriteContractSpecificationRequest) MsgTypeURL() string {
	return TypeURLMsgWriteContractSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteContractSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteContractSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if err := msg.ConvertOptionalFields(); err != nil {
		return err
	}
	return msg.Specification.ValidateBasic()
}

// ConvertOptionalFields will look at the SpecUuid field in the message.
// If present, it will be converted to a MetadataAddress and set in the Specification appropriately.
// Once used, it will be emptied so that calling this again has no effect.
func (msg *MsgWriteContractSpecificationRequest) ConvertOptionalFields() error {
	if len(msg.SpecUuid) > 0 {
		uid, err := uuid.Parse(msg.SpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		specAddr := ContractSpecMetadataAddress(uid)
		if !msg.Specification.SpecificationId.Empty() && !msg.Specification.SpecificationId.Equals(specAddr) {
			return fmt.Errorf("msg.Specification.SpecificationId [%s] is different from the one created from msg.SpecUuid [%s]",
				msg.Specification.SpecificationId, msg.SpecUuid)
		}
		msg.Specification.SpecificationId = specAddr
		msg.SpecUuid = ""
	}
	return nil
}

// ------------------  MsgDeleteContractSpecificationRequest  ------------------

// NewMsgDeleteContractSpecificationRequest creates a new msg instance
func NewMsgDeleteContractSpecificationRequest(specificationID MetadataAddress, signers []string) *MsgDeleteContractSpecificationRequest {
	return &MsgDeleteContractSpecificationRequest{SpecificationId: specificationID, Signers: signers}
}

func (msg MsgDeleteContractSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteContractSpecificationRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteContractSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteContractSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteContractSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgAddContractSpecToScopeSpecRequest  ------------------

// NewMsgAddContractSpecToScopeSpecRequest creates a new msg instance
func NewMsgAddContractSpecToScopeSpecRequest(contractSpecID MetadataAddress, scopeSpecID MetadataAddress, signers []string) *MsgAddContractSpecToScopeSpecRequest {
	return &MsgAddContractSpecToScopeSpecRequest{ContractSpecificationId: contractSpecID, ScopeSpecificationId: scopeSpecID, Signers: signers}
}

func (msg MsgAddContractSpecToScopeSpecRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgAddContractSpecToScopeSpecRequest) MsgTypeURL() string {
	return TypeURLMsgAddContractSpecToScopeSpecRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgAddContractSpecToScopeSpecRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgAddContractSpecToScopeSpecRequest) ValidateBasic() error {
	if !msg.ContractSpecificationId.IsContractSpecificationAddress() {
		return fmt.Errorf("address is not a contract specification id: %s", msg.ContractSpecificationId.String())
	}
	if !msg.ScopeSpecificationId.IsScopeSpecificationAddress() {
		return fmt.Errorf("address is not a scope specification id: %s", msg.ScopeSpecificationId.String())
	}
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgDeleteContractSpecFromScopeSpecRequest  ------------------

// NewMsgDeleteContractSpecFromScopeSpecRequest creates a new msg instance
func NewMsgDeleteContractSpecFromScopeSpecRequest(contractSpecID MetadataAddress, scopeSpecID MetadataAddress, signers []string) *MsgDeleteContractSpecFromScopeSpecRequest {
	return &MsgDeleteContractSpecFromScopeSpecRequest{ContractSpecificationId: contractSpecID, ScopeSpecificationId: scopeSpecID, Signers: signers}
}

func (msg MsgDeleteContractSpecFromScopeSpecRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteContractSpecFromScopeSpecRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteContractSpecFromScopeSpecRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteContractSpecFromScopeSpecRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteContractSpecFromScopeSpecRequest) ValidateBasic() error {
	if !msg.ContractSpecificationId.IsContractSpecificationAddress() {
		return fmt.Errorf("address is not a contract specification id: %s", msg.ContractSpecificationId.String())
	}
	if !msg.ScopeSpecificationId.IsScopeSpecificationAddress() {
		return fmt.Errorf("address is not a scope specification id: %s", msg.ScopeSpecificationId.String())
	}
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgWriteRecordSpecificationRequest  ------------------

// NewMsgAddRecordSpecificationRequest creates a new msg instance
func NewMsgWriteRecordSpecificationRequest(recordSpecification RecordSpecification, signers []string) *MsgWriteRecordSpecificationRequest {
	return &MsgWriteRecordSpecificationRequest{Specification: recordSpecification, Signers: signers}
}

func (msg MsgWriteRecordSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgWriteRecordSpecificationRequest) MsgTypeURL() string {
	return TypeURLMsgWriteRecordSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgWriteRecordSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgWriteRecordSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	if err := msg.ConvertOptionalFields(); err != nil {
		return err
	}
	return msg.Specification.ValidateBasic()
}

// ConvertOptionalFields will look at the ContractSpecUuid field in the message.
// If present, it will be converted to a MetadataAddress and set in the Specification appropriately.
// Once used, it will be emptied so that calling this again has no effect.
func (msg *MsgWriteRecordSpecificationRequest) ConvertOptionalFields() error {
	if len(msg.ContractSpecUuid) > 0 {
		uid, err := uuid.Parse(msg.ContractSpecUuid)
		if err != nil {
			return fmt.Errorf("invalid spec uuid: %w", err)
		}
		if len(strings.TrimSpace(msg.Specification.Name)) == 0 {
			return errors.New("empty specification name")
		}
		specAddr := RecordSpecMetadataAddress(uid, msg.Specification.Name)
		if !msg.Specification.SpecificationId.Empty() && !msg.Specification.SpecificationId.Equals(specAddr) {
			return fmt.Errorf("msg.Specification.SpecificationId [%s] is different from the one created from msg.ContractSpecUuid [%s] and msg.Specification.Name [%s]",
				msg.Specification.SpecificationId, msg.ContractSpecUuid, msg.Specification.Name)
		}
		msg.Specification.SpecificationId = specAddr
		msg.ContractSpecUuid = ""
	}
	return nil
}

// ------------------  MsgDeleteRecordSpecificationRequest  ------------------

// NewMsgDeleteRecordSpecificationRequest creates a new msg instance
func NewMsgDeleteRecordSpecificationRequest(specificationID MetadataAddress, signers []string) *MsgDeleteRecordSpecificationRequest {
	return &MsgDeleteRecordSpecificationRequest{SpecificationId: specificationID, Signers: signers}
}

func (msg MsgDeleteRecordSpecificationRequest) String() string {
	out, _ := yaml.Marshal(msg)
	return string(out)
}

func (msg MsgDeleteRecordSpecificationRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteRecordSpecificationRequest
}

// GetSigners returns the address(es) that must sign over msg.GetSignBytes()
func (msg MsgDeleteRecordSpecificationRequest) GetSigners() []sdk.AccAddress {
	return stringsToAccAddresses(msg.Signers)
}

// ValidateBasic performs a quick validity check
func (msg MsgDeleteRecordSpecificationRequest) ValidateBasic() error {
	if len(msg.Signers) < 1 {
		return fmt.Errorf("at least one signer is required")
	}
	return nil
}

// ------------------  MsgBindOSLocatorRequest  ------------------

// NewMsgBindOSLocatorRequest creates a new msg instance
func NewMsgBindOSLocatorRequest(obj ObjectStoreLocator) *MsgBindOSLocatorRequest {
	return &MsgBindOSLocatorRequest{
		Locator: obj,
	}
}

func (msg MsgBindOSLocatorRequest) MsgTypeURL() string {
	return TypeURLMsgBindOSLocatorRequest
}

func (msg MsgBindOSLocatorRequest) ValidateBasic() error {
	err := ValidateOSLocatorObj(msg.Locator.Owner, msg.Locator.EncryptionKey, msg.Locator.LocatorUri)
	if err != nil {
		return err
	}
	return nil
}

func (msg MsgBindOSLocatorRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{MustAccAddressFromBech32(msg.Locator.Owner)}
}

// ------------------  MsgDeleteOSLocatorRequest  ------------------

func NewMsgDeleteOSLocatorRequest(obj ObjectStoreLocator) *MsgDeleteOSLocatorRequest {
	return &MsgDeleteOSLocatorRequest{
		Locator: obj,
	}
}

func (msg MsgDeleteOSLocatorRequest) MsgTypeURL() string {
	return TypeURLMsgDeleteOSLocatorRequest
}

func (msg MsgDeleteOSLocatorRequest) ValidateBasic() error {
	err := ValidateOSLocatorObj(msg.Locator.Owner, msg.Locator.EncryptionKey, msg.Locator.LocatorUri)
	if err != nil {
		return err
	}

	return nil
}

// Signers returns the addrs of signers that must sign.
// CONTRACT: All signatures must be present to be valid.
// CONTRACT: Returns addrs in some deterministic order.
// here we assume msg for delete request has the right address
// should be verified later in the keeper?
func (msg MsgDeleteOSLocatorRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{MustAccAddressFromBech32(msg.Locator.Owner)}
}

// ValidateOSLocatorObj Validates OSLocatorObj data
func ValidateOSLocatorObj(ownerAddr, encryptionKey string, uri string) error {
	if strings.TrimSpace(ownerAddr) == "" {
		return fmt.Errorf("owner address cannot be empty")
	}

	if _, err := sdk.AccAddressFromBech32(ownerAddr); err != nil {
		return fmt.Errorf("failed to add locator for a given owner address,"+
			" invalid address: %s", ownerAddr)
	}

	if strings.TrimSpace(uri) == "" {
		return fmt.Errorf("uri cannot be empty")
	}

	if _, err := url.Parse(uri); err != nil {
		return fmt.Errorf("failed to add locator for a given"+
			" owner address, invalid uri: %s", uri)
	}

	if strings.TrimSpace(encryptionKey) != "" {
		if _, err := sdk.AccAddressFromBech32(encryptionKey); err != nil {
			return fmt.Errorf("failed to add locator for a given owner address: %s,"+
				" invalid encryption key address: %s", ownerAddr, encryptionKey)
		}
	}
	return nil
}

// ------------------  MsgModifyOSLocatorRequest  ------------------

func NewMsgModifyOSLocatorRequest(obj ObjectStoreLocator) *MsgModifyOSLocatorRequest {
	return &MsgModifyOSLocatorRequest{
		Locator: obj,
	}
}

func (msg MsgModifyOSLocatorRequest) MsgTypeURL() string {
	return TypeURLMsgModifyOSLocatorRequest
}

func (msg MsgModifyOSLocatorRequest) ValidateBasic() error {
	err := ValidateOSLocatorObj(msg.Locator.Owner, msg.Locator.EncryptionKey, msg.Locator.LocatorUri)
	if err != nil {
		return err
	}

	return nil
}

func (msg MsgModifyOSLocatorRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{MustAccAddressFromBech32(msg.Locator.Owner)}
}

// ------------------  SessionIdComponents  ------------------

func (msg *SessionIdComponents) GetSessionAddr() (*MetadataAddress, error) {
	var scopeUUID, sessionUUID *uuid.UUID
	if len(msg.SessionUuid) > 0 {
		uid, err := uuid.Parse(msg.SessionUuid)
		if err != nil {
			return nil, fmt.Errorf("invalid session uuid: %w", err)
		}
		sessionUUID = &uid
	}
	if msgScopeUUID := msg.GetScopeUuid(); len(msgScopeUUID) > 0 {
		uid, err := uuid.Parse(msgScopeUUID)
		if err != nil {
			return nil, fmt.Errorf("invalid scope uuid: %w", err)
		}
		scopeUUID = &uid
	} else if msgScopeAddr := msg.GetScopeAddr(); len(msgScopeAddr) > 0 {
		addr, addrErr := MetadataAddressFromBech32(msgScopeAddr)
		if addrErr != nil {
			return nil, fmt.Errorf("invalid scope addr: %w", addrErr)
		}
		uid, err := addr.ScopeUUID()
		if err != nil {
			return nil, fmt.Errorf("invalid scope addr: %w", err)
		}
		scopeUUID = &uid
	}

	if scopeUUID == nil && sessionUUID == nil {
		return nil, nil
	}
	if scopeUUID == nil {
		return nil, errors.New("session uuid provided but missing scope uuid or addr")
	}
	if sessionUUID == nil {
		return nil, errors.New("scope uuid or addr provided but missing session uuid")
	}
	ma := SessionMetadataAddress(*scopeUUID, *sessionUUID)
	return &ma, nil
}

// ------------------  Response Message Constructors  ------------------

func NewMsgWriteScopeResponse(scopeID MetadataAddress) *MsgWriteScopeResponse {
	return &MsgWriteScopeResponse{
		ScopeIdInfo: GetScopeIDInfo(scopeID),
	}
}

func NewMsgDeleteScopeResponse() *MsgDeleteScopeResponse {
	return &MsgDeleteScopeResponse{}
}

func NewMsgAddScopeDataAccessResponse() *MsgAddScopeDataAccessResponse {
	return &MsgAddScopeDataAccessResponse{}
}

func NewMsgDeleteScopeDataAccessResponse() *MsgDeleteScopeDataAccessResponse {
	return &MsgDeleteScopeDataAccessResponse{}
}

func NewMsgAddScopeOwnerResponse() *MsgAddScopeOwnerResponse {
	return &MsgAddScopeOwnerResponse{}
}

func NewMsgDeleteScopeOwnerResponse() *MsgDeleteScopeOwnerResponse {
	return &MsgDeleteScopeOwnerResponse{}
}

func NewMsgWriteSessionResponse(sessionID MetadataAddress) *MsgWriteSessionResponse {
	return &MsgWriteSessionResponse{
		SessionIdInfo: GetSessionIDInfo(sessionID),
	}
}

func NewMsgWriteRecordResponse(recordID MetadataAddress) *MsgWriteRecordResponse {
	return &MsgWriteRecordResponse{
		RecordIdInfo: GetRecordIDInfo(recordID),
	}
}

func NewMsgDeleteRecordResponse() *MsgDeleteRecordResponse {
	return &MsgDeleteRecordResponse{}
}

func NewMsgWriteScopeSpecificationResponse(scopeSpecID MetadataAddress) *MsgWriteScopeSpecificationResponse {
	return &MsgWriteScopeSpecificationResponse{
		ScopeSpecIdInfo: GetScopeSpecIDInfo(scopeSpecID),
	}
}

func NewMsgDeleteScopeSpecificationResponse() *MsgDeleteScopeSpecificationResponse {
	return &MsgDeleteScopeSpecificationResponse{}
}

func NewMsgWriteContractSpecificationResponse(contractSpecID MetadataAddress) *MsgWriteContractSpecificationResponse {
	return &MsgWriteContractSpecificationResponse{
		ContractSpecIdInfo: GetContractSpecIDInfo(contractSpecID),
	}
}

func NewMsgDeleteContractSpecificationResponse() *MsgDeleteContractSpecificationResponse {
	return &MsgDeleteContractSpecificationResponse{}
}

func NewMsgAddContractSpecToScopeSpecResponse() *MsgAddContractSpecToScopeSpecResponse {
	return &MsgAddContractSpecToScopeSpecResponse{}
}

func NewMsgDeleteContractSpecFromScopeSpecResponse() *MsgDeleteContractSpecFromScopeSpecResponse {
	return &MsgDeleteContractSpecFromScopeSpecResponse{}
}

func NewMsgWriteRecordSpecificationResponse(recordSpecID MetadataAddress) *MsgWriteRecordSpecificationResponse {
	return &MsgWriteRecordSpecificationResponse{
		RecordSpecIdInfo: GetRecordSpecIDInfo(recordSpecID),
	}
}

func NewMsgDeleteRecordSpecificationResponse() *MsgDeleteRecordSpecificationResponse {
	return &MsgDeleteRecordSpecificationResponse{}
}

func NewMsgBindOSLocatorResponse(objectStoreLocator ObjectStoreLocator) *MsgBindOSLocatorResponse {
	return &MsgBindOSLocatorResponse{
		Locator: objectStoreLocator,
	}
}

func NewMsgDeleteOSLocatorResponse(objectStoreLocator ObjectStoreLocator) *MsgDeleteOSLocatorResponse {
	return &MsgDeleteOSLocatorResponse{
		Locator: objectStoreLocator,
	}
}

func NewMsgModifyOSLocatorResponse(objectStoreLocator ObjectStoreLocator) *MsgModifyOSLocatorResponse {
	return &MsgModifyOSLocatorResponse{
		Locator: objectStoreLocator,
	}
}
