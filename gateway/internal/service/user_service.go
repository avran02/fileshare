package service

type UserService interface {
	RegisterUser()
	LoginUser()
	UpdateToken()
}

type userService struct{}

func (s *userService) RegisterUser() {}

func (s *userService) LoginUser() {}

func (s *userService) UpdateToken() {}

func NewUserService() UserService {
	return &userService{}
}
