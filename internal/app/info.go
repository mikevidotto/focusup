package app

type Info struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InfoService struct{}

func NewInfoService() *InfoService {
	return &InfoService{}
}

func (s *InfoService) GetInfo() Info {
	return Info{
		Name:    "FocusUp",
		Version: "0.1.0",
	}
}
