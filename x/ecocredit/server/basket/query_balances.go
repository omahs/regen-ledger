package basket

import (
	"context"
	"github.com/cosmos/cosmos-sdk/orm/model/ormlist"
	basketv1 "github.com/regen-network/regen-ledger/api/regen/ecocredit/basket/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	baskettypes "github.com/regen-network/regen-ledger/x/ecocredit/basket"
)

func (k Keeper) BasketBalances(ctx context.Context, request *baskettypes.QueryBasketBalancesRequest) (*baskettypes.QueryBasketBalancesResponse, error) {
	if request == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty request")
	}

	basket, err := k.stateStore.BasketStore().GetByBasketDenom(ctx, request.BasketDenom)
	if err != nil {
		return nil, err
	}

	pulsarPageReq, err := GogoPageReqToPulsarPageReq(request.Pagination)
	if err != nil {
		return nil, err
	}

	it, err := k.stateStore.BasketBalanceStore().List(ctx, basketv1.BasketBalancePrimaryKey{}.WithBasketId(basket.Id),
		ormlist.Paginate(pulsarPageReq),
	)
	if err != nil {
		return nil, err
	}

	res := &baskettypes.QueryBasketBalancesResponse{}
	for it.Next() {
		bal, err := it.Value()
		if err != nil {
			return nil, err
		}
		balanceGogo := &baskettypes.BasketBalance{}
		if err = PulsarToGogoSlow(bal, balanceGogo); err != nil {
			return nil, err
		}
		res.Balances = append(res.Balances, balanceGogo)
	}
	it.Close()

	res.Pagination, err = PulsarPageResToGogoPageRes(it.PageResponse())
	return res, err
}