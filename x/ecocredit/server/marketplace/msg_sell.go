package marketplace

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/cosmos/cosmos-sdk/orm/types/ormerrors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	marketApi "github.com/regen-network/regen-ledger/api/regen/ecocredit/marketplace/v1"
	"github.com/regen-network/regen-ledger/types/math"
	"github.com/regen-network/regen-ledger/x/ecocredit"
	"github.com/regen-network/regen-ledger/x/ecocredit/marketplace"
	"github.com/regen-network/regen-ledger/x/ecocredit/server/utils"
)

// Sell creates new sell orders for credits
func (k Keeper) Sell(ctx context.Context, req *marketplace.MsgSell) (*marketplace.MsgSellResponse, error) {
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	ownerAcc, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	sellOrderIds := make([]uint64, len(req.Orders))

	for i, order := range req.Orders {
		orderIndex := fmt.Sprintf("order[%d]", i)

		batch, err := k.coreStore.BatchTable().GetByDenom(ctx, order.BatchDenom)
		if err != nil {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf(
				"%s: batch denom %s: %s", orderIndex, order.BatchDenom, err.Error(),
			)
		}

		creditType, err := utils.GetCreditTypeFromBatchDenom(ctx, k.coreStore, batch.Denom)
		if err != nil {
			return nil, err
		}

		marketId, err := k.getOrCreateMarketId(ctx, creditType.Abbreviation, order.AskPrice.Denom)
		if err != nil {
			return nil, err
		}

		// verify expiration is in the future
		if order.Expiration != nil && !order.Expiration.After(sdkCtx.BlockTime()) {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf(
				"%s: expiration must be in the future: %s", orderIndex, order.Expiration,
			)
		}

		sellQty, err := math.NewPositiveFixedDecFromString(order.Quantity, creditType.Precision)
		if err != nil {
			return nil, err
		}

		if err = k.escrowCredits(ctx, orderIndex, ownerAcc, batch.Key, sellQty); err != nil {
			return nil, err
		}

		allowed, err := isDenomAllowed(ctx, order.AskPrice.Denom, k.stateStore.AllowedDenomTable())
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, sdkerrors.ErrInvalidRequest.Wrapf(
				"%s: %s is not allowed to be used in sell orders", orderIndex, order.AskPrice.Denom,
			)
		}

		var expiration *timestamppb.Timestamp
		if order.Expiration != nil {
			expiration = timestamppb.New(*order.Expiration)
		}

		id, err := k.stateStore.SellOrderTable().InsertReturningID(ctx, &marketApi.SellOrder{
			Seller:            ownerAcc,
			BatchKey:          batch.Key,
			Quantity:          order.Quantity,
			MarketId:          marketId,
			AskAmount:         order.AskPrice.Amount.String(),
			DisableAutoRetire: order.DisableAutoRetire,
			Expiration:        expiration,
			Maker:             true, // maker is always true for sell orders
		})
		if err != nil {
			return nil, err
		}

		sellOrderIds[i] = id

		if err = sdkCtx.EventManager().EmitTypedEvent(&marketplace.EventSell{
			OrderId: id,
		}); err != nil {
			return nil, err
		}

		sdkCtx.GasMeter().ConsumeGas(ecocredit.GasCostPerIteration, "ecocredit/core/MsgSell order iteration")
	}

	return &marketplace.MsgSellResponse{SellOrderIds: sellOrderIds}, nil
}

// getOrCreateMarketId attempts to get a market, creating one otherwise, and return the Id.
func (k Keeper) getOrCreateMarketId(ctx context.Context, creditTypeAbbrev, bankDenom string) (uint64, error) {
	market, err := k.stateStore.MarketTable().GetByCreditTypeAbbrevBankDenom(ctx, creditTypeAbbrev, bankDenom)
	switch err {
	case nil:
		return market.Id, nil
	case ormerrors.NotFound:
		return k.stateStore.MarketTable().InsertReturningID(ctx, &marketApi.Market{
			CreditTypeAbbrev:  creditTypeAbbrev,
			BankDenom:         bankDenom,
			PrecisionModifier: 0,
		})
	default:
		return 0, err
	}
}

func (k Keeper) escrowCredits(ctx context.Context, orderIndex string, account sdk.AccAddress, batchId uint64, quantity math.Dec) error {
	bal, err := k.coreStore.BatchBalanceTable().Get(ctx, account, batchId)
	if err != nil {
		return ecocredit.ErrInsufficientCredits.Wrapf(
			"%s: credit quantity: %v, tradable balance: 0", orderIndex, quantity,
		)
	}

	tradable, err := math.NewDecFromString(bal.TradableAmount)
	if err != nil {
		return err
	}

	newTradable, err := math.SafeSubBalance(tradable, quantity)
	if err != nil {
		return ecocredit.ErrInsufficientCredits.Wrapf(
			"%s: credit quantity: %v, tradable balance: %v", orderIndex, quantity, tradable,
		)
	}

	escrowed, err := math.NewDecFromString(bal.EscrowedAmount)
	if err != nil {
		return err
	}

	newEscrowed, err := math.SafeAddBalance(escrowed, quantity)
	if err != nil {
		return err
	}

	bal.TradableAmount = newTradable.String()
	bal.EscrowedAmount = newEscrowed.String()

	return k.coreStore.BatchBalanceTable().Update(ctx, bal)
}