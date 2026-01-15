package sender

import (
	"producer-service/pkg/models"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

func createRandomOrder() models.Order {
	orderUID := gofakeit.UUID()
	trackNumber := gofakeit.Regex("[A-Z0-9]{10,15}")

	order := models.Order{
		OrderUID:    orderUID,
		TrackNumber: trackNumber,
		Entry:       gofakeit.Regex("ENT[A-Z0-9]{3}"),
		Delivery: models.Delivery{
			Name:    gofakeit.Name(),
			Phone:   gofakeit.Phone(),
			Zip:     gofakeit.Zip(),
			City:    gofakeit.City(),
			Address: gofakeit.StreetNumber(),
			Region:  gofakeit.State(),
			Email:   gofakeit.Email(),
		},
		Payment: models.Payment{
			Transaction:  orderUID,
			RequestID:    gofakeit.UUID(),
			Currency:     gofakeit.CurrencyShort(),
			Provider:     gofakeit.Word(),
			Amount:       int(gofakeit.Price(1, 5000)),
			PaymentDT:    gofakeit.Date().Unix(),
			Bank:         gofakeit.Company(),
			DeliveryCost: int(gofakeit.Price(1, 1000)),
			GoodsTotal:   int(gofakeit.Price(100, 3000)),
			CustomFee:    int(gofakeit.Price(0, 100)),
		},
		Items: []models.Item{
			{
				ChrtID:      int(gofakeit.Int64()),
				TrackNumber: trackNumber,
				Price:       int(gofakeit.Price(1, 1000)),
				RID:         gofakeit.UUID(),
				Name:        gofakeit.ProductName(),
				Sale:        gofakeit.Number(0, 100),
				Size:        gofakeit.RandomString([]string{"XXS", "XXL", "XS", "S", "M", "L", "XL"}),
				TotalPrice:  gofakeit.Number(0, 1000),
				NmID:        int(gofakeit.Int64()),
				Brand:       gofakeit.Company(),
				Status:      gofakeit.Number(100, 600),
			},
		},
		Locale:            gofakeit.LanguageAbbreviation(),
		InternalSignature: gofakeit.HexUint(128),
		CustomerID:        gofakeit.UUID(),
		DeliveryService:   gofakeit.Company(),
		ShardKey:          strconv.Itoa(gofakeit.Number(1, 9)),
		SmID:              gofakeit.Number(1, 100),
		DateCreated:       gofakeit.Date().Format(time.RFC3339),
		OofShard:          strconv.Itoa(gofakeit.Number(1, 9)),
	}

	// С вероятностью 20% добавляем ошибки валидации
	if gofakeit.Number(1, 100) <= 20 {
		introduceValidationError(&order)
	}

	return order
}

func introduceValidationError(order *models.Order) {
	switch gofakeit.Number(1, 10) {
	case 1:
		order.OrderUID = "short" // <10 символов
	case 2:
		order.TrackNumber = "invalid!" // содержит спецсимволы
	case 3:
		order.Delivery.Email = "not-an-email" // не email
	case 4:
		order.Locale = "toolonglocale" // больше 2 символов
	case 5:
		order.Payment.Currency = "INVALID" // не ISO4217
	case 6:
		order.Items = []models.Item{} // пустой массив
	case 7:
		order.Delivery.Name = "" // пустое обязательное поле
	case 8:
		order.Payment.Amount = -100 // отрицательная сумма
	case 9:
		order.SmID = 1000 // больше max=999
	case 10:
		order.DateCreated = "invalid-date" // невалидная дата
	}
}
