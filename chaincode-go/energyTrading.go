package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// EnergyTradingContract defines the smart contract structure
type EnergyTradingContract struct {
	contractapi.Contract
}

// EnergyAsset defines the energy trading asset structure
type EnergyAsset struct {
	TokenID          string  `json:"tokenID"`
	BuyerAddress     string  `json:"buyerAddress"`
	SellerAddress    string  `json:"sellerAddress"`
	EnergyAmount     float64 `json:"energyAmount"`
	TransactionPrice float64 `json:"transactionPrice"`
	Timestamp        string  `json:"timestamp"`
	BuyerDeposit     float64 `json:"buyerDeposit"`
	SellerDeposit    float64 `json:"sellerDeposit"`
	TransactionState string  `json:"transactionState"`
	BuyerSignature   string  `json:"buyerSignature,omitempty"`
	SellerSignature  string  `json:"sellerSignature,omitempty"`
}

// TokenAccount defines a token account structure
type TokenAccount struct {
	AccountID string  `json:"accountID"`
	Balance   float64 `json:"balance"`
}

// Reputation defines the participant's reputation structure
type Reputation struct {
	ParticipantAddress string  `json:"participantAddress"`
	Score              float64 `json:"score"`
}

// ReputationPenaltyThreshold is the minimum acceptable reputation score
const ReputationPenaltyThreshold = 40.0

// InitLedger initializes ledger with energy assets and token accounts
func (e *EnergyTradingContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// 初始化账户余额
	accounts := []TokenAccount{
		{AccountID: "buyer1", Balance: 100.0},
		{AccountID: "seller1", Balance: 100.0},
	}

	for _, account := range accounts {
		accountJSON, err := json.Marshal(account)
		if err != nil {
			return err
		}
		if err := ctx.GetStub().PutState(account.AccountID, accountJSON); err != nil {
			return err
		}
	}

	// 初始化资产
	assets := []EnergyAsset{
		{
			TokenID:          "energy1",
			BuyerAddress:     "buyer1",
			SellerAddress:    "seller1",
			EnergyAmount:     100.0,
			TransactionPrice: 0.25,
			Timestamp:        "2025-05-03T10:00:00Z",
			BuyerDeposit:     10.0,
			SellerDeposit:    10.0,
			TransactionState: "CREATED",
			BuyerSignature:   "buyer_signature_example",
			SellerSignature:  "seller_signature_example",
		},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
		if err := ctx.GetStub().PutState(asset.TokenID, assetJSON); err != nil {
			return err
		}
	}

	// 初始化信誉分数
	reputations := []Reputation{
		{ParticipantAddress: "buyer1", Score: 80},
		{ParticipantAddress: "seller1", Score: 85},
	}

	for _, rep := range reputations {
		repJSON, err := json.Marshal(rep)
		if err != nil {
			return err
		}
		if err := ctx.GetStub().PutState(rep.ParticipantAddress, repJSON); err != nil {
			return err
		}
	}

	return nil
}

// Energy asset methods
func (e *EnergyTradingContract) ReadEnergyAsset(ctx contractapi.TransactionContextInterface, tokenID string) (*EnergyAsset, error) {
	assetJSON, err := ctx.GetStub().GetState(tokenID)
	if err != nil || assetJSON == nil {
		return nil, fmt.Errorf("asset %s does not exist", tokenID)
	}
	var asset EnergyAsset
	err = json.Unmarshal(assetJSON, &asset)
	return &asset, err
}

func (e *EnergyTradingContract) EnergyAssetExists(ctx contractapi.TransactionContextInterface, tokenID string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(tokenID)
	return assetJSON != nil, err
}

func (e *EnergyTradingContract) CreateEnergyAsset(ctx contractapi.TransactionContextInterface, tokenID, buyerAddress, sellerAddress string, energyAmount, transactionPrice float64, timestamp string, buyerDeposit, sellerDeposit float64) error {
	penalty, err := e.CheckReputationPenalty(ctx, buyerAddress)
	if penalty || err != nil {
		return fmt.Errorf("buyer %s reputation too low", buyerAddress)
	}
	penalty, err = e.CheckReputationPenalty(ctx, sellerAddress)
	if penalty || err != nil {
		return fmt.Errorf("seller %s reputation too low", sellerAddress)
	}
	exists, err := e.EnergyAssetExists(ctx, tokenID)
	if exists || err != nil {
		return fmt.Errorf("asset %s already exists", tokenID)
	}

	asset := EnergyAsset{
		TokenID:          tokenID,
		BuyerAddress:     buyerAddress,
		SellerAddress:    sellerAddress,
		EnergyAmount:     energyAmount,
		TransactionPrice: transactionPrice,
		Timestamp:        timestamp,
		BuyerDeposit:     buyerDeposit,
		SellerDeposit:    sellerDeposit,
		TransactionState: "CREATED",
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(tokenID, assetJSON)
}

// Reputation methods (已补充)
func (e *EnergyTradingContract) UpdateReputationScore(ctx contractapi.TransactionContextInterface, participantAddress string, delta float64) error {
	reputation, err := e.ReadReputationScore(ctx, participantAddress)
	if err != nil {
		return err
	}
	reputation.Score += delta
	if reputation.Score > 100 {
		reputation.Score = 100
	} else if reputation.Score < 0 {
		reputation.Score = 0
	}
	repJSON, err := json.Marshal(reputation)
	return ctx.GetStub().PutState(participantAddress, repJSON)
}

func (e *EnergyTradingContract) ReadReputationScore(ctx contractapi.TransactionContextInterface, participantAddress string) (*Reputation, error) {
	repJSON, err := ctx.GetStub().GetState(participantAddress)
	if repJSON == nil || err != nil {
		return &Reputation{ParticipantAddress: participantAddress, Score: 50}, nil
	}
	var rep Reputation
	err = json.Unmarshal(repJSON, &rep)
	return &rep, err
}

func (e *EnergyTradingContract) CheckReputationPenalty(ctx contractapi.TransactionContextInterface, participantAddress string) (bool, error) {
	reputation, err := e.ReadReputationScore(ctx, participantAddress)
	return reputation.Score < ReputationPenaltyThreshold, err
}

func main() {
	cc, err := contractapi.NewChaincode(new(EnergyTradingContract))
	if err != nil {
		log.Panic(err)
	}
	if err := cc.Start(); err != nil {
		log.Panic(err)
	}
}

