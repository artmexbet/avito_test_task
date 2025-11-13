package service

type iUserRepository interface {
	// Define user repository methods here
}

type UserService struct {
	userRepo iUserRepository
}

func NewUserService(userRepo iUserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}
