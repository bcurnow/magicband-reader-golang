package rfidsecuritysvc

import (
	"fmt"
	"time"
)

const (
	baseUrl      = "sounds"
	getUrlFormat = baseUrl + "/%v"
)

type Sound struct {
	ID                  int       `json:"id"`
	Name                string    `json:"name"`
	LastUpdateTimestamp time.Time `json:"last_update_timestamp"`
	Content             string    `json:"content"`
}

type SoundService interface {
	List() ([]Sound, error)
	Get(name string) (*Sound, error)
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

func (s *soundService) Get(name string) (*Sound, error) {
	var sound Sound
	if err := s.base.Get(fmt.Sprintf(getUrlFormat, name), 200, &sound); err != nil {
		return nil, err
	}
	return &sound, nil
}
