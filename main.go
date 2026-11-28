package main

import (
	"financialExchange/storage"
	"financialExchange/enum"
	"financialExchange/modifiers"

	"fmt"
	"log/slog"
	"os"
	"uuid"
	"time"
)

func main(){
	options := slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, &options)
	logger := slog.New(handler)

	// currencies := storage.NewStorage[string](5)
	// currencies.Add("BTC-RUB")
	// currencies.Add("ETH-RUB")
	// currencies.Add("USD-RUB")
	// currencies.Add("USD-RUB")

	// lengthOfCurrencies := currencies.Length()
	// logger.Info("Работа с валютами:",
	//  slog.Int("Количество уникальных комбинаций", lengthOfCurrencies))

	// traderStore := storage.NewStore[uuid.UUID, float64](10)
	// traderID := uuid.NewV7()

	// traderStore.Put(traderID, 50000.0004)
	// logger.Info("Новый трейдер добавлен в хранилище:",
	//  slog.String("UUID", traderID.String()),
	// slog.Float64("Начальный баланс", 50000.0004))


	// if balance, ok := traderStore.Get(traderID); ok{
	// 	logger.Info("поиск в хранилище:",
	//  	slog.String("найден трейдер с UUID", traderID.String()),
	// 	slog.Float64("его текущий баланс", balance))
	// }else{
	// 	logger.Error("элемент не найден в хранилище:",
	//  	slog.String("проверьте UUID", traderID.String()),)
	// }

	orderBook := storage.NewIndexedStorage[uuid.UUID, enum.Order](100)

	uuidOrder1 := uuid.NewV7()
	uuidOrder2 := uuid.NewV7()

	orderBuy := 	enum.Order{
						ID: uuidOrder1,
						TraderID: uuid.NewV4(),
						Side: enum.SideBuy,
						Type: enum.OrderTypeLimit,
						Symbo: "USD",
						Price: 1000.00,
						Quantity: 2.00,
						Filled: 0.00,
						Status: enum.OrderStatusNew,
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
							}
	// enum.LogOrderState(orderBuy, logger)




	orderSell := 	enum.Order{
						ID: uuidOrder2,
						TraderID: uuid.NewV4(),
						Side: enum.SideSell,
						Type: enum.OrderTypeLimit,
						Symbo: "USD",
						Price: 1000.00,
						Quantity: 2.00,
						Filled: 0.00,
						Status: enum.OrderStatusNew,
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
						}

	fmt.Scanln()
orderBook.Put(uuidOrder1, orderBuy)
orderBook.Put(uuidOrder2, orderSell)

fmt.Scanln()

orderSell = storage.ProcessSellOrder(orderSell, orderBook, logger)
orderBook.Put(uuidOrder2, orderSell)

	

	logger.Info("Зарегистрирована сделка",
			slog.String("Type", orderBuy.Type.String()),
			slog.Float64("Value:", orderBuy.Price),
			slog.Float64("Quantity", orderBuy.Quantity),
			slog.String("Symbo", orderBuy.Symbo),
			slog.String("TraderID", orderBuy.TraderID.String()))
								

	fmt.Scanln()

	fmt.Scanln()
	

}



func matchOrders(storage *storage.Store[uuid.UUID, enum.Order], buyID, sellID uuid.UUID, logger *slog.Logger) {
	// Получение ордера на покупку и продажу из хранилища
	buyOrder, buyOk := storage.Get(buyID)
	sellOrder, sellOk := storage.Get(sellID)
	if !buyOk || !sellOk {
		logger.Error("Возникли ошибки с поиском заказов в хранилище")
		return 
	}

	// Проврка совпадений торговых валют
	if buyOrder.Symbo != sellOrder.Symbo {
		logger.Warn("Сделка не может быть осуществлена в связи несовпадением валют")
		return
	}

	// Проверка что оба ордера еще не исполнены до конца 
	if buyOrder.Status == enum.OrderStatusFilled || sellOrder.Status == enum.OrderStatusFilled {
		logger.Warn("Заказы уже завершены")
		return
	}

	// Проверка что цена покупателя не ниже цены продавца
	if buyOrder.Price < sellOrder.Price {
		logger.Info("Сделка не может быть совершена из-за несовпадения цен")
		return
	}

	// Определение объема сделки 
	quantityTrade := min(buyOrder.Remaining(), sellOrder.Remaining())

	// Обновление остатков и статусов обоих ордеров в хранилище
	updatedBuyOrder, ok := storage.Update(buyID, modifiers.ModifierForOrder(quantityTrade))
	if !ok {
		logger.Error("Произошла ошибка при обновлении buyOrder")
	}

	updatedSellOrder, ok := storage.Update(sellID, modifiers.ModifierForOrder(quantityTrade))
	if !ok {
		logger.Error("Произошла ошибка при обновлении sellOrder")
	}

	// Логирование итогов обновления и детали совершенной сделки
	logger.Info("Заявки обнавлены", "buy_order", updatedBuyOrder, "sell_order", updatedSellOrder)
	logger.Info("Сделка успешно проведена",
		slog.Float64("quantity", quantityTrade),
		slog.Any("buy_order", updatedBuyOrder),
		slog.Any("sell_order", updatedSellOrder),
	)
}
