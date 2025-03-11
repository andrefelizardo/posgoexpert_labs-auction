package auction

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"fullcycle-auction_go/internal/entity/auction_entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func insertTestAuction(ar *AuctionRepository, auction *auction_entity.Auction) error {
	_, err := ar.Collection.InsertOne(context.TODO(), auction)
	return err
}

func TestMonitorExpiredAuctions(t *testing.T) {
	testDuration := 2
	os.Setenv("AUCTION_DURATION", strconv.Itoa(testDuration))
	defer os.Unsetenv("AUCTION_DURATION")

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("Deve fechar automaticamente leilão expirado", func(mt *mtest.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		auction := &auction_entity.Auction{
			Id:          "test-auction-1",
			ProductName: "Produto Teste",
			Category:    "Categoria",
			Description: "Descrição longa o suficiente",
			Condition:   auction_entity.New,
			Status:      auction_entity.Active,
			Timestamp:   time.Now().Add(-3 * time.Second),
		}

		ar := NewAuctionRepository(mt.DB)
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		if err := insertTestAuction(ar, auction); err != nil {
			t.Fatalf("falha ao inserir leilão de teste: %v", err)
		}

		ns := mt.DB.Name() + ".auctions"
		findResponse := mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bson.D{
			{Key: "_id", Value: auction.Id},
			{Key: "product_name", Value: auction.ProductName},
			{Key: "category", Value: auction.Category},
			{Key: "description", Value: auction.Description},
			{Key: "condition", Value: auction.Condition},
			{Key: "status", Value: auction_entity.Active},
			{Key: "timestamp", Value: auction.Timestamp.Unix()},
		})
		mt.AddMockResponses(findResponse)

		updateResponse := mtest.CreateSuccessResponse()
		mt.AddMockResponses(updateResponse)

		go ar.MonitorExpiredAuctions(ctx)
		time.Sleep(3 * time.Second)

		findOneResponse := mtest.CreateCursorResponse(1, ns, mtest.FirstBatch, bson.D{
			{Key: "_id", Value: auction.Id},
			{Key: "status", Value: auction_entity.Completed},
		})
		mt.AddMockResponses(findOneResponse)

		filter := bson.M{"_id": auction.Id}
		var result AuctionEntityMongo
		if err := ar.Collection.FindOne(ctx, filter).Decode(&result); err != nil {
			t.Fatalf("falha ao buscar leilão: %v", err)
		}
		if result.Status != auction_entity.Completed {
			t.Errorf("esperado que o status do leilão seja Completed, mas obteve %v", result.Status)
		}
	})
}