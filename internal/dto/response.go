package dto

type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type AuthTokenResponse struct {
	Token string `json:"token"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type TransferRequest struct {
	ReceiverEmail string `json:"receiver_email" validate:"required,email"`
	Amount        int64  `json:"amount" validate:"required,gt=0"`
}

type TopUpRequest struct {
	Amount int64 `json:"amount" validate:"required,gt=0"`
}

type WalletBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type TransferResponse struct {
	TransactionID uint   `json:"transaction_id"`
	Status        string `json:"status"`
}

type UserInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type TransactionHistoryItem struct {
	ID        uint     `json:"id"`
	Sender    UserInfo `json:"sender"`
	Receiver  UserInfo `json:"receiver"`
	Amount    int64    `json:"amount"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

type TransactionHistoryResponse struct {
	Transactions []TransactionHistoryItem `json:"transactions"`
}
