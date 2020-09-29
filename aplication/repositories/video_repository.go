package repositories

import (
	"encoder/domain"
	"fmt"

	uuid "github.com/satori/go.uuid"

	"github.com/jinzhu/gorm"
)

//VideoRepository - Video Repository interface
type VideoRepository interface {
	Insert(video *domain.Video) (*domain.Video, error)
	Find(id string) (*domain.Video, error)
}

//VideoRepositoryDb - Video Repository
type VideoRepositoryDb struct {
	Db *gorm.DB
}

//NewVideoRepository - create a VideoRepositoryDb instance
func NewVideoRepository(db *gorm.DB) *VideoRepositoryDb {
	return &VideoRepositoryDb{Db: db}
}

//Insert a new video
func (repo VideoRepositoryDb) Insert(video *domain.Video) (*domain.Video, error) {
	if video.ID == "" {
		video.ID = uuid.NewV4().String()
	}

	err := repo.Db.Create(video).Error

	if err != nil {
		return nil, err
	}

	return video, nil
}

//Find a video
func (repo VideoRepositoryDb) Find(id string) (*domain.Video, error) {
	var video domain.Video
	repo.Db.Preload("Jobs").First(&video, "id = ?", id)

	if video.ID == "" {
		return nil, fmt.Errorf("Video doest not exists")
	}

	return &video, nil
}
