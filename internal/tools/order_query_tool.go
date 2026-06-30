package tools

type OrderQueryTool struct{}

func NewOrderQueryTool() *OrderQueryTool {
	return &OrderQueryTool{}
}

func (t *OrderQueryTool) QueryOrder(orderID string) string {
	switch orderID {
	case "123456":
		return `{"status":"已签收","logistics":"已放在门口","signTime":"2026-06-23 14:30"}`
	case "67890":
		return `{"status":"配送中","estimatedArrival":"30分钟后","courierPhone":"13800138000"}`
	default:
		return `{"error":"订单不存在","orderId":"` + orderID + `"}`
	}
}
