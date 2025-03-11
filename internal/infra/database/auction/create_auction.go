package auction

import (
	"context"
	"fullcycle-auction_go/configuration/logger"
	"fullcycle-auction_go/internal/entity/auction_entity"
	"fullcycle-auction_go/internal/internal_error"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type AuctionEntityMongo struct {
	Id          string                          `bson:"_id"`
	ProductName string                          `bson:"product_name"`
	Category    string                          `bson:"category"`
	Description string                          `bson:"description"`
	Condition   auction_entity.ProductCondition `bson:"condition"`
	Status      auction_entity.AuctionStatus    `bson:"status"`
	Timestamp   int64                           `bson:"timestamp"`
}
type AuctionRepository struct {
	Collection *mongo.Collection
}

func NewAuctionRepository(database *mongo.Database) *AuctionRepository {
	return &AuctionRepository{
		Collection: database.Collection("auctions"),
	}
}

func (ar *AuctionRepository) CreateAuction(
	ctx context.Context,
	auctionEntity *auction_entity.Auction) *internal_error.InternalError {
	auctionEntityMongo := &AuctionEntityMongo{
		Id:          auctionEntity.Id,
		ProductName: auctionEntity.ProductName,
		Category:    auctionEntity.Category,
		Description: auctionEntity.Description,
		Condition:   auctionEntity.Condition,
		Status:      auctionEntity.Status,
		Timestamp:   auctionEntity.Timestamp.Unix(),
	}
	_, err := ar.Collection.InsertOne(ctx, auctionEntityMongo)
	if err != nil {
		logger.Error("Error trying to insert auction", err)
		return internal_error.NewInternalServerError("Error trying to insert auction")
	}

	go func(auctionId string) {
		auctionDuration := calculateAuctionDuration()
		logger.Info("leilão será fechado automaticamente após", zap.String("auctionId", auctionId), zap.Duration("auctionDuration", auctionDuration))
		time.Sleep(auctionDuration)

		if closeErr := ar.CloseAuction(ctx, auctionId); closeErr != nil {
			logger.Error("Erro ao fechar leilão automaticamente ", closeErr)
		} else {
			logger.Info("Leilão fechado automaticamente", zap.String("auctionId", auctionId))
		}
	}(auctionEntity.Id)

	return nil
}

func calculateAuctionDuration() time.Duration {
	durationStr := os.Getenv("AUCTION_DURATION")
	if durationStr == "" {
		return 60 * time.Second
	}
	durationInt, err := strconv.Atoi(durationStr)
	if err != nil {
		logger.Error("valor inválido para AUCTION_DURATION, utilizando 60 segundos como fallback", err)
		return 60 * time.Second
	}
	return time.Duration(durationInt) * time.Second
}

func (ar *AuctionRepository) CloseAuction(ctx context.Context, auctionId string) *internal_error.InternalError {
	filter := bson.M{"_id": auctionId, "status": auction_entity.Active}
	update := bson.M{"$set": bson.M{"status": auction_entity.Completed}}

	_, err := ar.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error("Error closing auction", err)
		return internal_error.NewInternalServerError("Error closing auction")
	}
	return nil
}

func (ar *AuctionRepository) MonitorExpiredAuctions(ctx context.Context) {
	auctionDuration := calculateAuctionDuration()
	ticker := time.NewTicker(auctionDuration / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			filter := bson.M{
				"status":    auction_entity.Active,
				"timestamp": bson.M{"$lte": time.Now().Add(-auctionDuration).Unix()},
			}
			cursor, err := ar.Collection.Find(ctx, filter)
			if err != nil {
				logger.Error("Erro ao consultar leilões expirados", err)
				continue
			}
			var expired []AuctionEntityMongo
			if err := cursor.All(ctx, &expired); err != nil {
				logger.Error("Erro ao ler leilões expirados", err)
				continue
			}
			for _, auction := range expired {
				if closeErr := ar.CloseAuction(ctx, auction.Id); closeErr != nil {
					logger.Error("Erro ao fechar leilão vencido", closeErr)
				} else {
					logger.Info("Leilão vencido fechado automaticamente", zap.String("auctionId", auction.Id))
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
