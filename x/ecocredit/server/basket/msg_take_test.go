package basket_test

import (
	"strconv"
	"testing"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/regen-network/gocuke"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	api "github.com/regen-network/regen-ledger/api/regen/ecocredit/basket/v1"
	coreapi "github.com/regen-network/regen-ledger/api/regen/ecocredit/v1"
	"github.com/regen-network/regen-ledger/x/ecocredit/basket"
	"github.com/regen-network/regen-ledger/x/ecocredit/core"
)

type takeSuite struct {
	*baseSuite
	alice             sdk.AccAddress
	bob               sdk.AccAddress
	aliceTokenBalance sdk.Coin
	bobTokenBalance   sdk.Coin
	classId           string
	creditTypeAbbrev  string
	batchDenom        string
	basketDenom       string
	tokenAmount       string
	jurisdiction      string
	res               *basket.MsgTakeResponse
	err               error
}

func TestTake(t *testing.T) {
	gocuke.NewRunner(t, &takeSuite{}).Path("./features/msg_take.feature").Run()
}

func (s *takeSuite) Before(t gocuke.TestingT) {
	s.baseSuite = setupBase(t)
	s.alice = s.addrs[0]
	s.bob = s.addrs[1]
	s.classId = "C01"
	s.creditTypeAbbrev = "C"
	s.batchDenom = "C01-001-20200101-20210101-001"
	s.basketDenom = "NCT"
	s.tokenAmount = "100"
	s.jurisdiction = "US-WA"
}

func (s *takeSuite) ABasket() {
	basketId, err := s.stateStore.BasketTable().InsertReturningID(s.ctx, &api.Basket{
		BasketDenom: s.basketDenom,
	})
	require.NoError(s.t, err)

	// add balance with credit amount = token amount
	s.addBasketClassAndBalance(basketId, s.tokenAmount)
}

func (s *takeSuite) ABasketWithDenom(a string) {
	basketId, err := s.stateStore.BasketTable().InsertReturningID(s.ctx, &api.Basket{
		BasketDenom: a,
	})
	require.NoError(s.t, err)

	// add balance with credit amount = token amount
	s.addBasketClassAndBalance(basketId, s.tokenAmount)
}

func (s *takeSuite) ABasketWithDisableAutoRetire(a string) {
	disableAutoRetire, err := strconv.ParseBool(a)
	require.NoError(s.t, err)

	basketId, err := s.stateStore.BasketTable().InsertReturningID(s.ctx, &api.Basket{
		BasketDenom:       s.basketDenom,
		DisableAutoRetire: disableAutoRetire,
	})
	require.NoError(s.t, err)

	// add balance with credit amount = token amount
	s.addBasketClassAndBalance(basketId, s.tokenAmount)
}

func (s *takeSuite) ABasketWithCreditBalance(a string) {
	basketId, err := s.stateStore.BasketTable().InsertReturningID(s.ctx, &api.Basket{
		BasketDenom: s.basketDenom,
	})
	require.NoError(s.t, err)

	s.addBasketClassAndBalance(basketId, a)
}

func (s *takeSuite) ABasketWithExponentAndDisableAutoRetire(a string, b string) {
	exponent, err := strconv.ParseUint(a, 10, 32)
	require.NoError(s.t, err)

	disableAutoRetire, err := strconv.ParseBool(b)
	require.NoError(s.t, err)

	basketId, err := s.stateStore.BasketTable().InsertReturningID(s.ctx, &api.Basket{
		BasketDenom:       s.basketDenom,
		Exponent:          uint32(exponent),
		DisableAutoRetire: disableAutoRetire,
	})
	require.NoError(s.t, err)

	// add balance with credit amount = token amount
	s.addBasketClassAndBalance(basketId, s.tokenAmount)
}

func (s *takeSuite) ABasketWithExponentAndCreditBalance(a string, b string) {
	exponent, err := strconv.ParseUint(a, 10, 32)
	require.NoError(s.t, err)

	basketId, err := s.stateStore.BasketTable().InsertReturningID(s.ctx, &api.Basket{
		BasketDenom: s.basketDenom,
		Exponent:    uint32(exponent),
	})
	require.NoError(s.t, err)

	s.addBasketClassAndBalance(basketId, b)
}

func (s *takeSuite) AliceOwnsBasketTokens() {
	amount, ok := sdk.NewIntFromString(s.tokenAmount)
	require.True(s.t, ok)

	s.aliceTokenBalance = sdk.NewCoin(s.basketDenom, amount)
}

func (s *takeSuite) AliceOwnsBasketTokenAmount(a string) {
	amount, ok := sdk.NewIntFromString(a)
	require.True(s.t, ok)

	s.aliceTokenBalance = sdk.NewCoin(s.basketDenom, amount)
}

func (s *takeSuite) AliceOwnsTokensWithDenom(a string) {
	amount, ok := sdk.NewIntFromString(s.tokenAmount)
	require.True(s.t, ok)

	s.aliceTokenBalance = sdk.NewCoin(a, amount)
}

func (s *takeSuite) AliceAttemptsToTakeCreditsWithBasketDenom(a string) {
	var coins sdk.Coins

	// send balance when amount not specified
	if !s.aliceTokenBalance.Equal(sdk.Coin{}) {
		coins = sdk.NewCoins(s.aliceTokenBalance)
	}

	s.bankKeeper.EXPECT().
		GetBalance(s.sdkCtx, s.alice, a).
		Return(s.aliceTokenBalance).
		AnyTimes() // not expected on failed attempt

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(s.sdkCtx, s.alice, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.bankKeeper.EXPECT().
		BurnCoins(s.sdkCtx, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.res, s.err = s.k.Take(s.ctx, &basket.MsgTake{
		Owner:                  s.alice.String(),
		BasketDenom:            a,
		Amount:                 s.aliceTokenBalance.Amount.String(),
		RetirementJurisdiction: s.jurisdiction,
		RetireOnTake:           true, // satisfy default auto-retire
	})
}

func (s *takeSuite) AliceAttemptsToTakeCreditsWithBasketTokenAmount(a string) {
	amount, ok := sdk.NewIntFromString(a)
	require.True(s.t, ok)

	coins := sdk.NewCoins(sdk.NewCoin(s.basketDenom, amount))

	s.bankKeeper.EXPECT().
		GetBalance(s.sdkCtx, s.alice, s.basketDenom).
		Return(s.aliceTokenBalance).
		Times(1)

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(s.sdkCtx, s.alice, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.bankKeeper.EXPECT().
		BurnCoins(s.sdkCtx, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.res, s.err = s.k.Take(s.ctx, &basket.MsgTake{
		Owner:                  s.alice.String(),
		BasketDenom:            s.basketDenom,
		Amount:                 a,
		RetirementJurisdiction: s.jurisdiction,
		RetireOnTake:           true, // satisfy default auto-retire
	})
}

func (s *takeSuite) AliceAttemptsToTakeCreditsWithBasketTokenAmountAndRetireOnTake(a string, b string) {
	amount, ok := sdk.NewIntFromString(a)
	require.True(s.t, ok)

	retireOnTake, err := strconv.ParseBool(b)
	require.NoError(s.t, err)

	coins := sdk.NewCoins(sdk.NewCoin(s.basketDenom, amount))

	s.bankKeeper.EXPECT().
		GetBalance(s.sdkCtx, s.alice, s.basketDenom).
		Return(s.aliceTokenBalance).
		Times(1)

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(s.sdkCtx, s.alice, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.bankKeeper.EXPECT().
		BurnCoins(s.sdkCtx, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.res, s.err = s.k.Take(s.ctx, &basket.MsgTake{
		Owner:                  s.alice.String(),
		BasketDenom:            s.basketDenom,
		Amount:                 a,
		RetirementJurisdiction: s.jurisdiction,
		RetireOnTake:           retireOnTake,
	})
}

func (s *takeSuite) AliceAttemptsToTakeCreditsWithRetireOnTake(a string) {
	retireOnTake, err := strconv.ParseBool(a)
	require.NoError(s.t, err)

	var coins sdk.Coins

	// send balance when amount not specified
	if !s.aliceTokenBalance.Equal(sdk.Coin{}) {
		coins = sdk.NewCoins(s.aliceTokenBalance)
	}

	s.bankKeeper.EXPECT().
		GetBalance(s.sdkCtx, s.alice, s.basketDenom).
		Return(s.aliceTokenBalance).
		AnyTimes() // not expected on failed attempt

	s.bankKeeper.EXPECT().
		SendCoinsFromAccountToModule(s.sdkCtx, s.alice, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.bankKeeper.EXPECT().
		BurnCoins(s.sdkCtx, basket.BasketSubModuleName, coins).
		Return(nil).
		AnyTimes() // not expected on failed attempt

	s.res, s.err = s.k.Take(s.ctx, &basket.MsgTake{
		Owner:                  s.alice.String(),
		BasketDenom:            s.basketDenom,
		Amount:                 s.aliceTokenBalance.Amount.String(),
		RetirementJurisdiction: s.jurisdiction,
		RetireOnTake:           retireOnTake,
	})
}

func (s *takeSuite) AliceHasATradableCreditBalanceWithAmount(a string) {
	batch, err := s.coreStore.BatchTable().GetByDenom(s.ctx, s.batchDenom)
	require.NoError(s.t, err)

	balance, err := s.coreStore.BatchBalanceTable().Get(s.ctx, s.alice, batch.Key)
	require.NoError(s.t, err)

	require.Equal(s.t, a, balance.Tradable)
}

func (s *takeSuite) AliceHasARetiredCreditBalanceWithAmount(a string) {
	batch, err := s.coreStore.BatchTable().GetByDenom(s.ctx, s.batchDenom)
	require.NoError(s.t, err)

	balance, err := s.coreStore.BatchBalanceTable().Get(s.ctx, s.alice, batch.Key)
	require.NoError(s.t, err)

	require.Equal(s.t, a, balance.Retired)
}

func (s *takeSuite) AliceHasABasketTokenBalanceWithAmount(a string) {
	amount, err := strconv.ParseInt(a, 10, 32)
	require.NoError(s.t, err)

	coin := sdk.NewInt64Coin(s.basketDenom, amount)

	s.bankKeeper.EXPECT().
		GetBalance(s.sdkCtx, s.alice, s.basketDenom).
		Return(coin).
		Times(1)

	balance := s.bankKeeper.GetBalance(s.sdkCtx, s.alice, s.basketDenom)

	require.Equal(s.t, coin, balance)
}

func (s *takeSuite) TheBasketHasACreditBalanceWithAmount(a string) {
	basket, err := s.stateStore.BasketTable().GetByBasketDenom(s.ctx, s.basketDenom)
	require.NoError(s.t, err)

	balance, err := s.stateStore.BasketBalanceTable().Get(s.ctx, basket.Id, s.batchDenom)
	require.NoError(s.t, err)

	require.Equal(s.t, a, balance.Balance)
}

func (s *takeSuite) TheBasketTokenHasATotalSupplyWithAmount(a string) {
	basket, err := s.stateStore.BasketTable().GetByBasketDenom(s.ctx, s.basketDenom)
	require.NoError(s.t, err)

	amount, err := strconv.ParseInt(a, 10, 32)
	require.NoError(s.t, err)

	coin := sdk.NewInt64Coin(s.basketDenom, amount)

	s.bankKeeper.EXPECT().
		GetSupply(s.sdkCtx, basket.BasketDenom).
		Return(coin).
		Times(1)

	supply := s.bankKeeper.GetSupply(s.sdkCtx, s.basketDenom)
	require.Equal(s.t, coin, supply)
}

func (s *takeSuite) ExpectNoError() {
	require.NoError(s.t, s.err)
}

func (s *takeSuite) ExpectTheError(a string) {
	require.EqualError(s.t, s.err, a)
}

func (s *takeSuite) ExpectTheResponse(a gocuke.DocString) {
	res := &basket.MsgTakeResponse{}
	err := jsonpb.UnmarshalString(a.Content, res)
	require.NoError(s.t, err)

	require.Equal(s.t, res, s.res)
}

func (s *takeSuite) addBasketClassAndBalance(basketId uint64, creditAmount string) {
	err := s.stateStore.BasketClassTable().Insert(s.ctx, &api.BasketClass{
		BasketId: basketId,
		ClassId:  s.classId,
	})
	require.NoError(s.t, err)

	classId := core.GetClassIdFromBatchDenom(s.batchDenom)
	creditTypeAbbrev := core.GetCreditTypeAbbrevFromClassId(classId)

	classKey, err := s.coreStore.ClassTable().InsertReturningID(s.ctx, &coreapi.Class{
		Id:               classId,
		CreditTypeAbbrev: creditTypeAbbrev,
	})
	require.NoError(s.t, err)

	projectKey, err := s.coreStore.ProjectTable().InsertReturningID(s.ctx, &coreapi.Project{
		ClassKey: classKey,
	})
	require.NoError(s.t, err)

	batchKey, err := s.coreStore.BatchTable().InsertReturningID(s.ctx, &coreapi.Batch{
		ProjectKey: projectKey,
		Denom:      s.batchDenom,
	})
	require.NoError(s.t, err)

	err = s.coreStore.BatchSupplyTable().Insert(s.ctx, &coreapi.BatchSupply{
		BatchKey:       batchKey,
		TradableAmount: creditAmount,
	})
	require.NoError(s.t, err)

	err = s.stateStore.BasketBalanceTable().Insert(s.ctx, &api.BasketBalance{
		BasketId:   basketId,
		BatchDenom: s.batchDenom,
		Balance:    creditAmount,
	})
	require.NoError(s.t, err)
}