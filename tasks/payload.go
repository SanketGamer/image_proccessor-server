// A "task" in Asynq = a job description.
// It has a type name + a payload (the data the worker needs).
// Think of it as an order ticket at a restaurant.

package tasks

const (
    TypeImageTransform = "image:transform" // the task type name
)

// TransformPayload holds everything the worker needs to do the transformation.
// This is serialized to JSON and stored in Redis.
type TransformPayload struct {
    ImageID    string `json:"image_id"`
    UserID     string `json:"user_id"`
    S3Key      string `json:"s3_key"`
    OutputKey  string `json:"output_key"`

    // Transformation options — all optional
    Resize     *ResizeOpts  `json:"resize,omitempty"`
    Crop       *CropOpts    `json:"crop,omitempty"`
    Rotate     float64      `json:"rotate,omitempty"`
    Grayscale  bool         `json:"grayscale,omitempty"`
    Sepia      bool         `json:"sepia,omitempty"`
    Flip       bool         `json:"flip,omitempty"`
    Mirror     bool         `json:"mirror,omitempty"`
    Format     string       `json:"format,omitempty"` // jpeg, png, webp
    Quality    int          `json:"quality,omitempty"`
}

type ResizeOpts struct {
    Width  int `json:"width"`
    Height int `json:"height"`
}

type CropOpts struct {
    Width  int `json:"width"`
    Height int `json:"height"`
    X      int `json:"x"`
    Y      int `json:"y"`
}