package rfidsecuritysvc

import (
	"fmt"
)

const (
	baseUrl      = "sounds"
	getUrlFormat = baseUrl + "/%v"
)

type SoundService interface {
	List() ([]Sound, error)
	Get(id int) (*Sound, error)
}

func (s *service) Sounds() SoundService {
	return &soundService{
		base: s,
	}
}

type soundService struct {
	base *service
}

func (s *soundService) List() ([]Sound, error) {
	var sounds []Sound
	if err := s.base.Get(baseUrl, 200, &sounds); err != nil {
		return nil, err
	}
	return sounds, nil
}

func (s *soundService) Get(id int) (*Sound, error) {
	var sound Sound
	if err := s.base.Get(fmt.Sprintf(getUrlFormat, id), 200, &sound); err != nil {
		return nil, err
	}
	return &sound, nil
}
