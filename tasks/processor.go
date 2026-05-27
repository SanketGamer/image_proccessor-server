package tasks

import (
    "bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"github.com/aws/aws-sdk-go-v2/aws"
	goaws "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/hibiken/asynq"
	"gorm.io/gorm"
	"image_service/models"
	appSentry "image_service/sentry"
)

type ImageProcessor struct {
    DB       *gorm.DB
    S3Client *goaws.Client
    S3Bucket string
}

// ProcessImageTransform is called by Asynq whenever a transform job is ready.
// asynq passes the task → we decode payload → apply transformations → save to S3 → update DB
// decode -> then download img from s3 -> encode into bytes
func (p *ImageProcessor) ProcessImageTransform(ctx context.Context, t *asynq.Task) error{
	defer appSentry.RecoverWithSentry()
	//decode the JSON payload back into our struct
	var payload TransformPayload
	if err:=json.Unmarshal(t.Payload(), &payload); err!=nil{
       appSentry.CaptureWorkerError(err, TypeImageTransform, t.ResultWriter().TaskID())
		return fmt.Errorf("failed to decode payload: %w",err)
	}
	//download the org image from s3
	result,err:=p.S3Client.GetObject(ctx, &goaws.GetObjectInput{
		Bucket: aws.String(p.S3Bucket),
		Key: aws.String(payload.S3Key),
	})
	if err!=nil{
		return fmt.Errorf("s3 download failed: %w",err)
	}
	defer result.Body.Close()
	//Step 3: decode the image bytes into an image.Image (Go's image interface)
    //imaging.Decode handles JPEG, PNG, GIF, TIFF, BMP automatically
	//now raw byte actual image object
	img,err:=imaging.Decode(result.Body)
	if err!=nil{
		return fmt.Errorf("image decode failed: %w", err)
	}
	//apply each transformation in order, applyTranformations do imaeg crop,resize,color all the stuf
	img=applyTransformations(img,payload)
	//encode the res back to bytes
	format:=imaging.JPEG
	contentType:= "image/jpeg"
	if payload.Format == "png" {
		format=imaging.PNG
		contentType="image/png"
	}else if payload.Format == "webp"{
		format=imaging.JPEG
	}
	var buf bytes.Buffer
	if err:=imaging.Encode(&buf,img,format); err!=nil{
	  return fmt.Errorf("image encode failed: %w", err)
	}

	//upload the transformed image to S3 under a new key. means add a file in s3bucket
	_,err=p.S3Client.PutObject(ctx, &goaws.PutObjectInput{
		Bucket: aws.String(p.S3Bucket),
		Key: aws.String(payload.OutputKey),
		Body: bytes.NewReader(buf.Bytes()),
		ContentType: aws.String(contentType),
	})
	if err!=nil{
       appSentry.CaptureWorkerError(err, TypeImageTransform, t.ResultWriter().TaskID())
	   return fmt.Errorf("s3 upload failed: %w", err)
	}
	//Build the transformed image URL
    transformedURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", p.S3Bucket, payload.OutputKey)
	//update the image record in db
	 if err := p.DB.Model(&models.Image{}).
        Where("id = ?", payload.ImageID).
        Updates(map[string]interface{}{
            "transformed_url": transformedURL,
            "status":          "done",
        }).Error; err != nil {
        return fmt.Errorf("db update failed: %w", err)
    }
	return nil  // returning nil = success, Asynq removes the job from queue
}

// applyTransformations applies each operation in sequence.
// imaging library works like a pipeline: each step returns a new image.
//img is org image, p means payload(data) or user choices, edited images
func applyTransformations(img image.Image, p TransformPayload) image.Image {
	//resize -> changing the image size
	if p.Resize!=nil{
		img=imaging.Resize(img,p.Resize.Width, p.Resize.Height, imaging.Lanczos)
	}
	//crop -> cutting only one part of the img, 
	if p.Crop!=nil{
		img=imaging.Crop(img,image.Rect(
			p.Crop.X, p.Crop.Y,
			p.Crop.X+p.Crop.Width, p.Crop.Y+p.Crop.Height,
		))
	}
	//rotate -> turning img (90,180 deg)
	if p.Rotate!=0{
		img=imaging.Rotate(img,p.Rotate,color.White)
	}
	//flip -> vertical flip(top becomes bottom,bottom becomes top)
	if p.Flip{
		img=imaging.FlipV(img)
	}
	//mirror- horizontal flip(left-right)
	if p.Mirror{
		img=imaging.FlipH(img)
	}
	//grayscale-remove color turns images into ->black,white,gray
	if p.Grayscale{
		img=imaging.Grayscale(img)
	}
	 // SEPIA — warm brown vintage tone. means old vintage brown photo effect.
    // We do it manually: grayscale first, then tint
	if p.Sepia{
		img=imaging.Grayscale(img)
		img=imaging.AdjustFunc(img,func(c color.NRGBA) color.NRGBA{
			r:=float64(c.R)
			g:=float64(c.G)
			b:=float64(c.B)
			return color.NRGBA{
				R: clamp(r*0.393 + g*0.769 + b*0.189),
				G: clamp(r*0.349 + g*0.686 + b*0.168),
                B: clamp(r*0.272 + g*0.534 + b*0.131),
				A: c.A,
			}
		})
	}
	return img
}

//color value must stay between 0 to 255
func clamp(v float64) uint8{
	if v<0 { return 0}
	if v>255 {return 255}
	return uint8(v)
}