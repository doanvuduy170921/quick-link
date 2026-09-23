package http

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"doanvuduyndh@gmail.com"`
	Password string `json:"password" binding:"required,min=6,max=32" example:"123456"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"doanvuduyndh@gmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type UserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}
