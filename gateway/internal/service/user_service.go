package service

type UserService interface {
	RegisterUser()
	LoginUser()
	RefreshToken()
	Logout()
}

type userService struct {
	userServiceClient pb.UserServiceClient
}

func (s *userService) RegisterUser() {

}

func (s *userService) LoginUser() {}

func (s *userService) RefreshToken() {}

func (s *userService) Logout() {}

func NewUserService() UserService {
	return &userService{}
}
