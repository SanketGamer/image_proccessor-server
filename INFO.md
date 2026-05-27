
docker exec -it imageservice_postgres psql -U postgres -d authdb  (to see inside docker)

1. register -- http://localhost:8080/api/v1/register   (work)
{
  "username": "sanketuser",
  "email": "guddu@gmail.com",
  "password": "12345678"
}

2. login -- http://localhost:8080/api/v1/login   (work)
{
  "email": "guddu@gmail.com",
  "password": "12345678"
}

3. post image upload -- http://localhost:8080/api/v1/images  (work)
add headers - Authorization : Bearer <token>
body - form-data : image(file type):image

4. Get image by ID -- http://localhost:8080/api/v1/images/abc123  (concurrently do 5 task at a time)
add headers - Authorization : Bearer <token>

5. transform image(resize,crop,color chnage) -- http://localhost:8080/api/v1/images/image_id/transform
add headers - Authorization : Bearer <token>
{
  "resize": {
    "width": 400,
    "height": 400
  },
  "grayscale": true,
  "rotate": 90
}

6.  List all images (paginated) -- http://localhost:8080/api/v1/images?page=1&limit=10
add headers - Authorization : Bearer <token>


# test error - 
1. curl -X POST http://localhost:8080/api/v1/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "email": "test@gmail.com", "password": }' 


----
Tech Stack -- 
Gin     = makes routing and HTTP handling simple and fast
Asynq   = moves slow work to background so user never waits
Redis   = keeps hot data in RAM so DB is never hammered
## why we use aws s3 and postgres db
S3 stores the FILE, PostgreSQL stores the INFORMATION about the file
User asks → "show me my images"

1. Go hits PostgreSQL
   → SELECT * FROM images WHERE user_id = 'abc'
   → returns list of URLs instantly

2. User clicks an image
   → browser opens the S3 URL directly
   → Go server not even involved
   → S3 serves the file directly to browser