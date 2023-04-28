package middleware

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/overload-ak/cosmos-firewall/config"
	"github.com/overload-ak/cosmos-firewall/internal/chain"
	"github.com/overload-ak/cosmos-firewall/internal/utils"
)

type Validator struct {
	chain chain.IChain
	cfg   *config.Config
}

func NewValidator(cfg *config.Config) Validator {
	validator := Validator{cfg: cfg}
	validator.chain = chain.CreateChainFactory(validator.cfg.ChainID, validator.cfg.JSONRPC, validator.cfg.GRPC, validator.cfg.Rest, validator.cfg.Whitelist)
	return validator
}

func (v Validator) IsJSONPRCPathAllowed(path string) bool {
	for _, p := range v.chain.GetJSONRPCPaths() {
		if strings.EqualFold(p, path) {
			return true
		}
	}
	return false
}

func (v Validator) IsGRPCPathAllowed(path string) bool {
	for _, p := range v.chain.GetGRPCPaths() {
		if strings.EqualFold(p, path) {
			return true
		}
	}
	return false
}

func (v Validator) IsRESTPathAllowed(path string) bool {
	pathPatterns := make([]utils.PathPattern, 0, len(v.chain.GetRESTPaths()))
	for _, p := range v.chain.GetRESTPaths() {
		pathPatterns = append(pathPatterns, utils.NewPathPattern(p))
	}
	for _, pattern := range pathPatterns {
		if pattern.Match(path) {
			return true
		}
	}
	return false
}

func (v Validator) CheckTxBytes(txBytes []byte) error {
	txRaw := tx.TxRaw{}
	if err := proto.Unmarshal(txBytes, &txRaw); err != nil {
		return errors.Wrapf(err, "proto unmarshal txBytes")
	}
	if len(txRaw.Signatures) <= 0 {
		return errors.New("signatures is empty")
	}
	for _, signature := range txRaw.Signatures {
		if len(signature) != 64 && len(signature) != 65 {
			return errors.New("signature format error")
		}
	}
	authInfo := tx.AuthInfo{}
	if err := proto.Unmarshal(txRaw.AuthInfoBytes, &authInfo); err != nil {
		return errors.Wrapf(err, "proto unmarshal authInfo")
	}
	if authInfo.Fee.GasLimit < 30000 {
		return errors.New("GasLimit is too small")
	}
	if err := v.CheckTxAuthInfo(authInfo); err != nil {
		return errors.Wrapf(err, "check txAuthInfo")
	}

	txBody := tx.TxBody{}
	if err := proto.Unmarshal(txRaw.BodyBytes, &txBody); err != nil {
		return errors.Wrapf(err, "proto unmarshal txBody")
	}
	// with whitelist
	// feeFreeWhitelist := FeeFreeWhitelist(txBody)
	if err := v.CheckTxBody(txBody); err != nil {
		return errors.Wrapf(err, "check txBody")
	}
	return nil
}

func (v Validator) CheckTxBody(txBody tx.TxBody) error {
	if len(txBody.Memo) > 256 {
		return errors.New("memo field length exceeds limit")
	}
	//if txBody.TimeoutHeight > 0 {
	//	return One, "Set transaction timeout block height."
	//}
	if len(txBody.ExtensionOptions) > 0 {
		return errors.New("fill in illegal field ExtensionOptions")
	}
	if len(txBody.NonCriticalExtensionOptions) > 0 {
		return errors.New("fill in illegal field NonCriticalExtensionOptions")
	}
	if len(txBody.Messages) <= 0 {
		return errors.New("transaction message is empty")
	}
	for _, message := range txBody.Messages {
		if message.TypeUrl == "" {
			return errors.New("message type url is empty")
		}
		if !v.IsGRPCPathAllowed(message.TypeUrl) {
			return errors.New("unsupported transaction message type")
		}
	}
	return nil
}

func (v Validator) CheckTxAuthInfo(authInfo tx.AuthInfo) error {
	if authInfo.Fee.Granter != "" {
		return errors.New("fill in illegal field Granter")
	}
	if authInfo.Fee.Payer != "" {
		return errors.New("set Payer, non-normal client request")
	}
	if len(authInfo.SignerInfos) != 1 {
		return errors.New("multiple SignerInfos, non-normal client request")
	}
	for _, info := range authInfo.SignerInfos {
		if info.PublicKey == nil {
			return errors.New("public key is empty")
		}
		if info.PublicKey.TypeUrl != "/cosmos.crypto.secp256k1.PubKey" &&
			info.PublicKey.TypeUrl != "/ethermint.crypto.v1.ethsecp256k1.PubKey" {
			return errors.New("illegal public key type")
		}
		if len(info.PublicKey.Value) != 35 {
			return errors.New("public key format error")
		}
		if single, ok := info.ModeInfo.Sum.(*tx.ModeInfo_Single_); !ok {
			return errors.New("invalid signature sum")
		} else {
			if single.Single.Mode == signing.SignMode_SIGN_MODE_UNSPECIFIED || single.Single.Mode == signing.SignMode_SIGN_MODE_TEXTUAL {
				return errors.New("signature mode error")
			}
		}
	}
	return nil
}
