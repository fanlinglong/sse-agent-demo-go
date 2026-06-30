package tools

type LogisticsQueryTool struct{}

func NewLogisticsQueryTool() *LogisticsQueryTool {
	return &LogisticsQueryTool{}
}

func (t *LogisticsQueryTool) QueryLogistics(orderID string) string {
	switch orderID {
	case "123456":
		return `{"tracking":[{"time":"2026-06-22 08:00","event":"包裹已揽收"},{"time":"2026-06-22 18:30","event":"到达分拣中心"},{"time":"2026-06-23 09:00","event":"派送中"},{"time":"2026-06-23 11:25","event":"已签收，放在门口"}],"carrier":"顺丰速运","contactPhone":"95338"}`
	default:
		return `{"error":"未找到物流信息","orderId":"` + orderID + `"}`
	}
}
