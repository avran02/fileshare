package service

type ShareService interface {
	Share()
	Unshare()
}

type shareService struct{}

func (s *shareService) Share() {}

func (s *shareService) Unshare() {}

func NewShareService() ShareService {
	return &shareService{}
}
