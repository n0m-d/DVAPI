package domain

type AdminCreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	FullName string `json:"full_name" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=student instructor admin"`
}

type AdminUpdateUserRequest struct {
	Email    *string `json:"email" binding:"omitempty,email"`
	FullName *string `json:"full_name"`
	Role     *string `json:"role" binding:"omitempty,oneof=student instructor admin"`
}

type AdminUserListData struct {
	Users      []User      `json:"users"`
	Pagination *Pagination `json:"pagination"`
}

type AdminUserListResponse struct {
	Status string             `json:"status"`
	Data   *AdminUserListData `json:"data"`
}

type AdminStats struct {
	Users       int `json:"users"`
	Students    int `json:"students"`
	Instructors int `json:"instructors"`
	Courses     int `json:"courses"`
	Enrollments int `json:"enrollments"`
	Assignments int `json:"assignments"`
	Submissions int `json:"submissions"`
}
