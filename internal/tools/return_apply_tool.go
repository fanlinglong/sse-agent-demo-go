package tools

type ReturnApplyTool struct{}

func NewReturnApplyTool() *ReturnApplyTool {
	return &ReturnApplyTool{}
}

func (t *ReturnApplyTool) ApplyReturn(orderID, reason string) string {
	return `{"success":true,"returnId":"RET-` + orderID + `","message":"退货申请已创建，请等待审核","nextSteps":"请在48小时内将商品寄回指定地址"}`
}
