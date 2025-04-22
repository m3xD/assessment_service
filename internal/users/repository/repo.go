package repository

import (
	models "assessment_service/internal/model"
	"assessment_service/internal/util"
	"gorm.io/gorm"
	"time"
)

type UserRepository interface {
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error
	List(params util.PaginationParams) ([]models.User, int64, error)
	UpdateLastLogin(id uint) error
	GetUserStats() (int64, int64, error)
	CountAll() (int64, error)
	GetNewUsersCount(i int) (int64, error)
	GetListUserByAttempt(params util.PaginationParams, attemptID uint) ([]models.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

func (r *userRepository) List(params util.PaginationParams) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	// Apply filters
	if params.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if params.Filters != nil {
		if role, ok := params.Filters["role"]; ok && role != "all" {
			query = query.Where("role = ?", role)
		}

		if status, ok := params.Filters["status"]; ok && status != "all" {
			query = query.Where("status = ?", status)
		}
	}

	// Count total before pagination
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortDir)
	} else {
		query = query.Order("created_at DESC")
	}

	query = query.Offset(params.Offset).Limit(params.Limit)

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) UpdateLastLogin(id uint) error {
	now := time.Now()
	return r.db.Model(&models.User{}).Where("id = ?", id).Update("last_login", now).Error
}

func (r *userRepository) GetUserStats() (int64, int64, error) {
	var activeCount, inactiveCount int64

	if err := r.db.Model(&models.User{}).Where("status = ?", "active").Count(&activeCount).Error; err != nil {
		return 0, 0, err
	}

	if err := r.db.Model(&models.User{}).Where("status = ?", "inactive").Count(&inactiveCount).Error; err != nil {
		return 0, 0, err
	}

	return activeCount, inactiveCount, nil
}

func (r *userRepository) CountAll() (int64, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepository) GetNewUsersCount(days int) (int64, error) {
	var count int64
	startDate := time.Now().AddDate(0, 0, -days)
	if err := r.db.Model(&models.User{}).Where("created_at >= ?", startDate).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userRepository) GetListUserByAttempt(params util.PaginationParams, attemptID uint) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{}).
		Joins("attempts at ON at.user_id = id").Where("at.id = ?", attemptID)

	// Apply filters
	if params.Search != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting and pagination
	if params.SortBy != "" {
		query = query.Order(params.SortBy + " " + params.SortDir)
	} else {
		query = query.Order("created_at DESC")
	}

	query = query.Offset(params.Offset).Limit(params.Limit)

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil

}
