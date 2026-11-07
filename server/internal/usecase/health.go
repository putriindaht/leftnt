package usecase

type HealthStatus struct {
	Status string
	Game   string
}

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) Check() HealthStatus {
	return HealthStatus{
		Status: "ok",
		Game:   "LEFTN'T",
	}
}
