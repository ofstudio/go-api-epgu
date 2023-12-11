package dto

type ErrorResponse struct {
	Code    string `JSON:"code"`
	Message string `JSON:"message"`
}

type OrderIdResponse struct {
	OrderId int `JSON:"orderId"`
}

type OrderInfoResponse struct {
	Code    string `JSON:"code"`
	Message string `JSON:"message"`
	Order   string `JSON:"order"`
}
