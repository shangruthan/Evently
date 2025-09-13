# Create directories
New-Item -ItemType Directory -Force -Path cmd\api
New-Item -ItemType Directory -Force -Path internal\api\handler
New-Item -ItemType Directory -Force -Path internal\api\middleware
New-Item -ItemType Directory -Force -Path internal\config
New-Item -ItemType Directory -Force -Path internal\data
New-Item -ItemType Directory -Force -Path internal\service

# Create files
New-Item -ItemType File cmd\api\main.go
New-Item -ItemType File internal\api\handler\auth.go
New-Item -ItemType File internal\api\handler\booking.go
New-Item -ItemType File internal\api\handler\event.go
New-Item -ItemType File internal\api\handler\response.go
New-Item -ItemType File internal\api\middleware\auth.go
New-Item -ItemType File internal\api\router.go
New-Item -ItemType File internal\config\config.go
New-Item -ItemType File internal\data\models.go
New-Item -ItemType File internal\data\booking_repo.go
New-Item -ItemType File internal\data\event_repo.go
New-Item -ItemType File internal\data\user_repo.go
New-Item -ItemType File internal\service\auth_service.go
New-Item -ItemType File internal\service\booking_service.go
