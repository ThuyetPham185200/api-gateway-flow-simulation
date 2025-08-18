# API Gateway Test Commands

## 1. Auth - Login
curl -X POST http://localhost:8080/login \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"login":"alice","password":"123"}'

## 2. Auth - Register
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"123"}'

## 3. Auth - Change Password
curl -X PUT http://localhost:8080/me/password \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"old_password":"123","new_password":"456"}'

## 4. Auth - Delete Account
curl -X DELETE http://localhost:8080/me \
  -H "Authorization: Bearer fake-jwt-token"

## 5. User - Get User Profile
curl -X GET http://localhost:8080/users/123 \
  -H "Authorization: Bearer fake-jwt-token"

## 6. User - Update Own Profile
curl -X PATCH http://localhost:8080/me \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice_new","bio":"Hello world"}'

## 7. User - Search Users
curl -X GET "http://localhost:8080/users?search=alice&offset=0&limit=10&sort=username" \
  -H "Authorization: Bearer fake-jwt-token"

## 8. Posts - Get Post
curl -X GET http://localhost:8080/posts/1 \
  -H "Authorization: Bearer fake-jwt-token"

## 9. Posts - Get User Posts
curl -X GET "http://localhost:8080/users/123/posts?offset=0&limit=10" \
  -H "Authorization: Bearer fake-jwt-token"

## 10. Posts - Get Own Posts
curl -X GET "http://localhost:8080/me/posts?offset=0&limit=10" \
  -H "Authorization: Bearer fake-jwt-token"

## 11. Posts - Create Post
curl -X POST http://localhost:8080/posts \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello world","media_ids":[1,2]}'

## 12. Posts - Update Post
curl -X PATCH http://localhost:8080/posts/1 \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"content":"Updated content"}'

## 13. Posts - Delete Post
curl -X DELETE http://localhost:8080/posts/1 \
  -H "Authorization: Bearer fake-jwt-token"

## 14. Reactions - Get Reactions
curl -X GET http://localhost:8080/posts/1/reactions \
  -H "Authorization: Bearer fake-jwt-token"

## 15. Reactions - React to Post
curl -X POST http://localhost:8080/posts/1/reactions \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"reaction_type":"like"}'

## 16. Reactions - Remove Reaction
curl -X DELETE http://localhost:8080/posts/1/reactions \
  -H "Authorization: Bearer fake-jwt-token" \
  -H "Content-Type: application/json" \
  -d '{"reaction_type":"like"}'
