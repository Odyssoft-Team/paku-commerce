package runtime

import (
	cartmemory "paku-commerce/internal/commerce/cart/adapters/memory"
	cartdomain "paku-commerce/internal/commerce/cart/domain"
	checkoutmemory "paku-commerce/internal/commerce/checkout/adapters/memory"
	checkoutdomain "paku-commerce/internal/commerce/checkout/domain"
)

// Singletons de repositorios para compartir estado entre módulos.
var (
	CartRepoSingleton  cartdomain.CartRepository      = cartmemory.NewCartRepository()
	OrderRepoSingleton checkoutdomain.OrderRepository = checkoutmemory.NewOrderRepository()
)
