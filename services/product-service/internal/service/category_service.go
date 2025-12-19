package service

import (
	"GoProduct/internal/domain"
	"GoProduct/internal/repo"
	"errors"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

var errCategoryNotFound = errors.New("category_not_found")
var errSlugTaken = errors.New("slug_already_taken")

func IsCategoryNotFound(err error) bool { return errors.Is(err, errCategoryNotFound) }
func IsSlugTaken(err error) bool        { return errors.Is(err, errSlugTaken) }

type CreateCategoryInput struct {
	Name string
	Slug string
}

type UpdateCategoryInput struct {
	ID   uint
	Name string
	Slug string
}

type CategoryService struct {
	categories *repo.Categories
}

func NewCategoryService(categories *repo.Categories) *CategoryService {
	return &CategoryService{categories: categories}
}

func (s *CategoryService) CreateCategory(in CreateCategoryInput) (*domain.Category, error) {
	slug := in.Slug
	if slug == "" {
		slug = generateSlug(in.Name)
	}

	existing, _ := s.categories.GetBySlug(slug)
	if existing != nil {
		return nil, errSlugTaken
	}

	c := &domain.Category{
		Name: in.Name,
		Slug: slug,
	}

	if err := s.categories.Create(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *CategoryService) GetCategory(id uint) (*domain.Category, error) {
	c, err := s.categories.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errCategoryNotFound
		}
		return nil, err
	}
	return c, nil
}

func (s *CategoryService) ListCategories() ([]domain.Category, error) {
	return s.categories.List()
}

func (s *CategoryService) UpdateCategory(in UpdateCategoryInput) (*domain.Category, error) {
	c, err := s.categories.GetByID(in.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errCategoryNotFound
		}
		return nil, err
	}

	slug := in.Slug
	if slug == "" {
		slug = generateSlug(in.Name)
	}

	if slug != c.Slug {
		existing, _ := s.categories.GetBySlug(slug)
		if existing != nil {
			return nil, errSlugTaken
		}
	}

	c.Name = in.Name
	c.Slug = slug

	if err := s.categories.Update(c); err != nil {
		return nil, err
	}

	return c, nil
}

func (s *CategoryService) DeleteCategory(id uint) error {
	_, err := s.categories.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errCategoryNotFound
		}
		return err
	}

	return s.categories.Delete(id)
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile(`[^a-z0-9-]`)
	slug = reg.ReplaceAllString(slug, "")
	return slug
}
