package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	regentypes "github.com/regen-network/regen-ledger/types"
	"github.com/regen-network/regen-ledger/x/ecocredit/base"
	types "github.com/regen-network/regen-ledger/x/ecocredit/base/types/v1"
)

// Batch queries for information on a credit batch.
func (k Keeper) Batch(ctx context.Context, request *types.QueryBatchRequest) (*types.QueryBatchResponse, error) {
	if err := base.ValidateBatchDenom(request.BatchDenom); err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("batch denom: %s", err)
	}

	batch, err := k.stateStore.BatchTable().GetByDenom(ctx, request.BatchDenom)
	if err != nil {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("could not get batch with denom %s: %s", request.BatchDenom, err.Error())
	}

	issuer := sdk.AccAddress(batch.Issuer)

	project, err := k.stateStore.ProjectTable().Get(ctx, batch.ProjectKey)
	if err != nil {
		return nil, err
	}

	info := types.BatchInfo{
		Issuer:       issuer.String(),
		ProjectId:    project.Id,
		Denom:        batch.Denom,
		Metadata:     batch.Metadata,
		StartDate:    regentypes.ProtobufToGogoTimestamp(batch.StartDate),
		EndDate:      regentypes.ProtobufToGogoTimestamp(batch.EndDate),
		IssuanceDate: regentypes.ProtobufToGogoTimestamp(batch.IssuanceDate),
		Open:         batch.Open,
	}

	return &types.QueryBatchResponse{Batch: &info}, nil
}
