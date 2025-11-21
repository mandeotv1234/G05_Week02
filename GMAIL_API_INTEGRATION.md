# Gmail API Integration

Ứng dụng này đã được tích hợp Gmail API để lấy email thật từ tài khoản Google của bạn.

## Tính năng

- ✅ Đăng nhập bằng Google OAuth 2.0
- ✅ Lấy danh sách mailbox/labels từ Gmail
- ✅ Đọc emails từ Gmail inbox
- ✅ Đánh dấu email đã đọc
- ✅ Toggle star/unstar email
- ✅ Xem chi tiết email với HTML content

## Cấu hình

### 1. Google Cloud Console Setup

Xem chi tiết tại [GOOGLE_OAUTH_SETUP.md](./GOOGLE_OAUTH_SETUP.md)

**Các bước quan trọng:**

1. Tạo OAuth 2.0 Client ID
2. Enable Gmail API
3. Thêm scopes: `gmail.readonly` và `gmail.modify`
4. Thêm test users (nếu app ở chế độ Testing)

### 2. Environment Variables

**Backend (.env):**

```env
PORT=8080
JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=168h

# Google OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URI=http://localhost:8080/api/auth/google/callback
```

**Frontend (.env):**

```env
VITE_API_URL=http://localhost:8080/api
VITE_GOOGLE_CLIENT_ID=your-google-client-id
```

### 3. Chạy ứng dụng

**Backend:**

```bash
cd backend
go run main.go
```

**Frontend:**

```bash
cd frontend
npm install
npm run dev
```

## Kiến trúc

### Backend

```
backend/
├── pkg/gmail/           # Gmail API service
│   └── service.go       # Gmail API integration
├── internal/
│   ├── auth/
│   │   ├── domain/      # User với AccessToken & RefreshToken
│   │   ├── usecase/     # GoogleSignIn lưu tokens
│   │   └── delivery/    # HTTP handlers
│   └── email/
│       ├── usecase/     # Sử dụng Gmail service
│       └── delivery/    # HTTP handlers với userID
```

### Flow

1. **Đăng nhập:**

   - User click "Sign in with Google"
   - Frontend request Google OAuth với scopes Gmail
   - Google trả về access token + refresh token
   - Frontend gửi tokens đến backend
   - Backend lưu tokens vào user record

2. **Lấy emails:**

   - Frontend gọi API `/mailboxes` hoặc `/emails`
   - Backend lấy access token từ user record
   - Backend gọi Gmail API với access token
   - Gmail API trả về emails thật
   - Backend chuyển đổi format và trả về frontend

3. **Thao tác emails:**
   - Mark as read: Gọi Gmail API để remove UNREAD label
   - Toggle star: Gọi Gmail API để add/remove STARRED label
   - View email: Fetch full email content từ Gmail API

## Scopes được sử dụng

- `email` - Email address của user
- `profile` - Tên và avatar
- `https://www.googleapis.com/auth/gmail.readonly` - Đọc emails
- `https://www.googleapis.com/auth/gmail.modify` - Sửa labels (read/star status)

## Limitations

- App phải ở chế độ Testing trong Google Cloud Console
- Chỉ test users được thêm mới có thể sử dụng
- Access token hết hạn sau 1 giờ (có thể implement refresh token flow)
- Quota limits của Gmail API (mặc định: 1 billion quota units/day)

## Troubleshooting

### "Access blocked: This app's request is invalid"

- Kiểm tra lại Authorized JavaScript origins trong Google Cloud Console
- Phải khớp chính xác với URL đang chạy (http://localhost:5173)

### "Access token expired"

- Implement refresh token flow để tự động renew access token
- Hoặc đăng xuất và đăng nhập lại

### "Insufficient Permission"

- Kiểm tra Gmail API đã được enable trong Google Cloud Console
- Kiểm tra scopes đã được thêm vào OAuth consent screen
- Revoke access và authorize lại: https://myaccount.google.com/permissions

### Không lấy được emails

- Kiểm tra access token đã được lưu đúng chưa
- Kiểm tra user đã authorize với Gmail scopes
- Xem logs backend để debug Gmail API errors

## References

- [Gmail API Documentation](https://developers.google.com/gmail/api)
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2)
- [@react-oauth/google](https://www.npmjs.com/package/@react-oauth/google)
