# api-gateway-flow-simulation

# 1. Run gateway.go and internal_service.go in seperate terminal

# 2. curl from user to simulated request to test
## 1. Login
curl -X POST http://localhost:8080/auth/login   -H "Authorization: Bearer fake-jwt-token"   -d '{"login":"alice","password":"123"}'

## 2. Đăng ký user mới (/auth/register)
curl -X POST http://localhost:8080/auth/register \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"login":"bob","password":"123"}'

## 3. Lấy profile (/profile/get)
curl -X GET http://localhost:8080/profile/get \
  -H "Authorization: Bearer fake-jwt-token"

## 4. Cập nhật profile (/profile/update)
curl -X POST http://localhost:8080/profile/update \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","email":"bob@example.com"}'

