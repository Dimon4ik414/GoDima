package main

import (
	"financialExchange/enum"
	"financialExchange/modifiers"
	"financialExchange/storage"

	// "fmt"
	"log/slog"
	"os"
	"time"
	"uuid"
)

func main() {
	options := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, &options)
	logger := slog.New(handler)

	orderStorage := storage.NewStore[uuid.UUID, enum.Order](100)
	tradeStorage := storage.NewStore[uuid.UUID, enum.Trade](100)

	uuidOrder1 := uuid.NewV7()
	uuidOrder2 := uuid.NewV7()

	orderBuy := enum.Order{
		ID:        uuidOrder1,
		TraderID:  uuid.NewV4(),
		Side:      enum.SideBuy,
		Type:      enum.OrderTypeLimit,
		Symbo:     "USD",
		Price:     1000.00,
		Quantity:  2.00,
		Filled:    0.00,
		Status:    enum.OrderStatusNew,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	orderSell := enum.Order{
		ID:        uuidOrder2,
		TraderID:  uuid.NewV4(),
		Side:      enum.SideSell,
		Type:      enum.OrderTypeLimit,
		Symbo:     "USD",
		Price:     1000.00,
		Quantity:  2.00,
		Filled:    0.00,
		Status:    enum.OrderStatusNew,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}


	orderStorage.Put(uuidOrder1, orderBuy)
	orderStorage.Put(uuidOrder2, orderSell)


	trade, ok := matchOrders(orderStorage, tradeStorage, uuidOrder1, uuidOrder2, logger)
	if ok {
		logger.Info("Зарегистрирована сделка", slog.Any("trade", trade))
	}
}

	func matchOrders(storage *storage.Store[uuid.UUID, enum.Order], tradeStorage *storage.Store[uuid.UUID, enum.Trade], buyID, sellID uuid.UUID, logger *slog.Logger) (enum.Trade, bool) {
	// Получение ордера на покупку и продажу из хранилища
	buyOrder, buyOk := storage.Get(buyID)
	sellOrder, sellOk := storage.Get(sellID)
	if !buyOk || !sellOk {
		logger.Error("Возникли ошибки с поиском заказов в хранилище")
		return enum.Trade{}, false
	}

	// Проверка совпадений торговых валют
	if buyOrder.Symbo != sellOrder.Symbo {
		logger.Warn("Сделка не может быть осуществлена в связи несовпадением валют")
		return enum.Trade{}, false
	}

	// Проверка что оба ордера еще не исполнены до конца
	if buyOrder.Status == enum.OrderStatusFilled || sellOrder.Status == enum.OrderStatusFilled {
		logger.Warn("Заказы уже завершены")
		return enum.Trade{}, false
	}

	// Проверка что цена покупателя не ниже цены продавца
	if buyOrder.Price < sellOrder.Price {
		logger.Info("Сделка не может быть совершена из-за несовпадения цен")
		return enum.Trade{}, false
	}

	// Определение объема сделки
	quantityTrade := min(buyOrder.Remaining(), sellOrder.Remaining())

	// Обновление остатков и статусов обоих ордеров в хранилище
	updatedBuyOrder, ok := storage.Update(buyID, modifiers.ModifierForOrder(quantityTrade))
	if !ok {
		logger.Error("Произошла ошибка при обновлении buyOrder")
		return enum.Trade{}, false
	}

	updatedSellOrder, ok := storage.Update(sellID, modifiers.ModifierForOrder(quantityTrade))
	if !ok {
		logger.Error("Произошла ошибка при обновлении sellOrder")
		return enum.Trade{}, false
	}

	//значение структуры
	trade := enum.Trade{
		ID:          uuid.New(),
		Symbol:      buyOrder.Symbo,
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		BuyerID:     buyOrder.TraderID,
		SellerID:    sellOrder.TraderID,
		Price:       sellOrder.Price,
		Quantity:    quantityTrade,
		ExecutedAt:  time.Now().UTC(),
	}

	tradeStorage.Put(trade.ID, trade) //возврат и хранение

	// Логирование итогов обновления и детали совершенной сделки
	logger.Info("Заявки обнавлены", "buy_order", updatedBuyOrder, "sell_order", updatedSellOrder)
	logger.Info("Сделка успешно проведена",
		slog.Any("trade", trade),
		slog.Float64("quantity", quantityTrade),
		slog.Any("buy_order", updatedBuyOrder),
		slog.Any("sell_order", updatedSellOrder),
	)
	return trade, true
}
