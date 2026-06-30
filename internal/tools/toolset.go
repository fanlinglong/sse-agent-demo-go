package tools

type Toolset struct {
	OrderQueryTool     *OrderQueryTool
	LogisticsQueryTool *LogisticsQueryTool
	NotificationTool   *NotificationTool
	ReturnApplyTool    *ReturnApplyTool
}

func NewToolset(orderQuery *OrderQueryTool, logisticsQuery *LogisticsQueryTool, notification *NotificationTool, returnApply *ReturnApplyTool) *Toolset {
	return &Toolset{
		OrderQueryTool:     orderQuery,
		LogisticsQueryTool: logisticsQuery,
		NotificationTool:   notification,
		ReturnApplyTool:    returnApply,
	}
}
