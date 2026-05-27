package services

import (
	"bytes"
	"context"
	"fmt"
	"image_service/models"
	"mime/multipart"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)


type ImageService struct{
	DB *gorm.DB
	Redis *redis.Client
	S3Client *s3.Client
	S3Bucket string
}

//Receive image => Read binary data => Get file extension => Create unique filename => Upload to S3 => Generate public URL => Save metadata in PostgreSQL
func (s *ImageService) UploadImage(file multipart.File, header *multipart.FileHeader, userID string) (*models.Image,error){
	//it reads the file data into memory as bytes (binary data).
	buf:=new(bytes.Buffer)
	if _,err:=buf.ReadFrom(file); err!=nil{
		return nil,err
	}
    // Build a unique S3 key (filename) using UUID to avoid collisions, ex:"images/550e8400-e29b-41d4-a716-446655440000.jpg"
	ext:=getExtension(header.Filename)
	s3Key:=fmt.Sprintf("image/%s%s",uuid.NewString(),ext)
	//upload to s3 ,PutObject = "put this object (file) at this key (path) in this bucket"
	 _,err:=s.S3Client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(s.S3Bucket),
        Key:         aws.String(s3Key),
        Body:        bytes.NewReader(buf.Bytes()),
        ContentType: aws.String(header.Header.Get("Content-Type")),
	 })
	 if err!=nil{
		return nil,fmt.Errorf("s3 upload failed: %w",err)
	 }

	 //Build public url
	url := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.S3Bucket, s3Key)
	//Parse userID
	uid,err:=uuid.Parse(userID)
	if err!=nil{
		return nil,err
	}
	 // Save metadata to PostgreSQL
    img := &models.Image{
        UserID:      uid,
        OriginalURL: url,
        Filename:    header.Filename,
        Format:      ext[1:],          // remove the dot from ".jpg" → "jpg"
        Size:        header.Size,
        Status:      "uploaded",
    }
    if err := s.DB.Create(img).Error; err != nil {
        return nil, err
    }
    return img, nil
}

//retrive the image by image id, checks redis 1st, Cache-aside pattern: check cache → if miss → check DB → store in cache
func (s *ImageService) GetImage(imageID,userID string) (*models.Image,error){
	cacheKey:=fmt.Sprintf("image:%s",imageID)
	//try redis first - if found skip db entirely
	cached,err:= s.Redis.Get(context.Background(),cacheKey).Result()
	if err!=nil{
		_=cached
	}
	var img models.Image
	result:= s.DB.Where("id = ? AND user_id = ?",imageID,userID).First(&img)
	if result.Error != nil{
		return nil,result.Error
	}

	//Store in redis for next time -- expires after 1 hr
	s.Redis.Set(context.Background(),cacheKey,img.OriginalURL, time.Hour)
	return &img,nil
}

//returns a paginated list of images for a user
// page=1, limit=10 → OFFSET 0 LIMIT 10
// page=2, limit=10 → OFFSET 10 LIMIT 10
func (s *ImageService) ListImage(userID string,page,limit int) ([]models.Image,int64,error){
	var images []models.Image
	var total int64
	offset:=(page-1)*limit
	s.DB.Model(&models.Image{}).Where("user_id = ?",userID).Count(&total)
	result:=s.DB.Where("user_id = ?",userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&images)
	return images,total,result.Error
}

// UpdateStatus updates the processing status of an image.
// Called after enqueuing a transform job → sets status to "processing"
// Called by worker after done → sets status to "done"
func (s *ImageService) UpdateStatus(imageID string, status string) error {
    result := s.DB.Model(&models.Image{}).
        Where("id = ?", imageID).
        Update("status", status)

    return result.Error
}

//return photo.png -> .png
func getExtension(filename string) string {
	for i:=len(filename)-1;i>=0;i--{
		if filename[i]=='.'{
			return filename[i:]
		}
	}
	return ".jpg"
}


