package user

type UserResponse struct {
	ID          int    `json:"id"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	Gender      string `json:"gender"`
	Username    string `json:"username"`
	DateOfBirth string `json:"date_of_birth"`
	Role        string `json:"role"`
}

type PaginationRequest struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"pageSize" binding:"omitempty,min=1,max=100"`
	Filter   string `form:"filter"`
}

type PaginatedResponse struct {
	StatusCode int            `json:"statusCode"`
	Message    string         `json:"message"`
	PageNumber int            `json:"pageNumber"`
	TotalPages int            `json:"totalPages"`
	FromItem   int            `json:"fromItem"`
	ToItem     int            `json:"toItem"`
	TotalItem  int            `json:"totalItem"`
	Data       []UserResponse `json:"data"`
}
