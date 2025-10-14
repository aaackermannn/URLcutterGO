package service

import (
	"fmt"
	"log"
	"net/url"
	"time"
	"urlcutter/internal/models"
	"urlcutter/internal/repository"
	"urlcutter/pkg/shortener"
)

type Service interface {
	CreateShortURL(original string) (*models.CreateURLResponse, error)
	GetOriginalURL(short string) (string, error)
	Redirect(short string) (string, error)
}

type URLService struct {
	repo repository.Repository
}

func NewURLService(repo repository.Repository) *URLService {
	return &URLService{repo: repo}
}

func (s *URLService) CreateShortURL(original string) (*models.CreateURLResponse, error) {
	//Валидация URL
	if !isValidURL(original) {
		return nil, fmt.Errorf("invalid URL")
	}

	//Проверяем не сокращали ли уже этот url
	existing, err := s.repo.FindByOriginal(original)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return &models.CreateURLResponse{ShortURL: existing.Short}, nil
	}

	//Генерируем короткую ссылку
	short, err := shortener.GenerateShortURL()
	if err != nil {
		return nil, err
	}

	//Создаем запись в БД
	url := &models.URL{
		Id:        short,
		Original:  original,
		Short:     short,
		CreatedAt: time.Now(),
		Clicks:    0,
	}

	if err := s.repo.Create(url); err != nil {
		return nil, err
	}

	return &models.CreateURLResponse{ShortURL: short}, nil
}

func (s *URLService) GetOriginalURL(short string) (string, error) {
	url, err := s.repo.FindByShort(short)
	if err != nil {
		return "", err
	}
	if url == nil {
		return "", fmt.Errorf("URL not found")
	}
	return url.Original, nil
}

func (s *URLService) Redirect(short string) (string, error) {
	original, err := s.GetOriginalURL(short)
	if err != nil {
		return "", err
	}

	//Увеличиваем счетчик кликов
	if err := s.repo.IncrementClicks(short); err != nil {
		log.Printf("Failed to increment clicks: %v", err)
	}

	return original, nil
}

func isValidURL(u string) bool {
	parsed, err := url.Parse(u)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}
