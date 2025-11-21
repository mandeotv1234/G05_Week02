# Tích hợp Gmail API - Tóm tắt thay đổi

## Tổng quan

Đã tích hợp thành công Gmail API vào ứng dụng email để lấy dữ liệu thật từ Gmail của người dùng.

## Các thay đổi chính

### Backend

#### 1. Dependencies mới

- `google.golang.org/api/gmail/v1` - Gmail API client
- `golang.org/x/oauth2` - OAuth2 authentication

#### 2. Files mới

- `backend/pkg/gmail/service.go` - Service để tương tác với Gmail API
  - `GetMailboxes()` - Lấy danh sách labels/mailboxes
  - `GetEmails()` - Lấy danh sách emails từ label
  - `GetEmailByID()` - Lấy chi tiết một email
  - `MarkAsRead()` - Đánh dấu email đã đọc
  - `ToggleStar()` - Toggle star/unstar
  - `SendEmail()` - Gửi email (future feature)

#### 3. Files đã sửa

**backend/internal/auth/domain/user.go**

- Thêm fields: `AccessToken`, `RefreshToken`, `TokenExpiry`
- Lưu Google OAuth tokens vào user record

**backend/internal/auth/dto/auth.go**

- Thêm `AccessToken` và `RefreshToken` vào `GoogleSignInRequest`

**backend/internal/auth/usecase/auth_usecase.go**

- Cập nhật `GoogleSignIn()` để nhận và lưu access/refresh tokens
- Lưu tokens vào user khi đăng nhập

**backend/internal/email/usecase/email_usecase.go**

- Hoàn toàn viết lại để sử dụng Gmail API
- Lấy access token từ user record
- Fallback về mock data nếu không có token
- Tất cả methods giờ cần `userID` parameter

**backend/internal/email/delivery/handler.go**

- Cập nhật tất cả handlers để lấy user từ context
- Pass userID đến usecase methods

**backend/main.go**

- Initialize Gmail service
- Inject Gmail service vào email usecase

### Frontend

#### 1. Dependencies

Sử dụng `@react-oauth/google` với `useGoogleLogin` hook thay vì `GoogleLogin` component

#### 2. Files đã sửa

**frontend/src/pages/LoginPage.tsx**

- Thay `GoogleLogin` component bằng `useGoogleLogin` hook
- Thêm Gmail scopes: `gmail.readonly`, `gmail.modify`
- Gửi access token đến backend
- Custom Google sign-in button

**frontend/src/pages/SignUpPage.tsx**

- Tương tự LoginPage
- Thêm Gmail scopes và custom button

**frontend/src/types/auth.ts**

- Thêm `accessToken` và `refreshToken` vào `GoogleSignInRequest`

### Documentation

#### 1. Files mới

- `GMAIL_API_INTEGRATION.md` - Hướng dẫn chi tiết về tích hợp Gmail API

#### 2. Files đã cập nhật

- `GOOGLE_OAUTH_SETUP.md` - Thêm hướng dẫn enable Gmail API và scopes

## Cách sử dụng

### 1. Cấu hình Google Cloud Console

```
1. Enable Gmail API
2. Thêm scopes vào OAuth consent screen:
   - https://www.googleapis.com/auth/gmail.readonly
   - https://www.googleapis.com/auth/gmail.modify
3. Thêm test users (nếu app ở chế độ Testing)
```

### 2. Environment Variables

**Backend (.env):**

```env
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-client-secret
```

**Frontend (.env):**

```env
VITE_GOOGLE_CLIENT_ID=your-client-id
```

### 3. Chạy ứng dụng

```bash
# Backend
cd backend
go run main.go

# Frontend
cd frontend
npm run dev
```

### 4. Đăng nhập

1. Click "Sign in with Google"
2. Chọn tài khoản Google (phải là test user)
3. Chấp nhận quyền truy cập Gmail
4. App sẽ tự động lấy emails từ Gmail

## Tính năng hoạt động

✅ Đăng nhập bằng Google OAuth với Gmail scopes
✅ Lấy danh sách mailboxes/labels từ Gmail
✅ Hiển thị emails từ INBOX, SENT, DRAFT, etc.
✅ Xem chi tiết email với HTML content
✅ Đánh dấu email đã đọc (remove UNREAD label)
✅ Toggle star/unstar email
✅ Fallback về mock data nếu không có access token

## Limitations

- App phải ở chế độ Testing trong Google Cloud Console
- Chỉ test users mới có thể sử dụng
- Access token hết hạn sau 1 giờ (chưa implement auto-refresh)
- Chưa implement send email feature
- Quota limits của Gmail API áp dụng

## Testing

### Build Backend

```bash
cd backend
go build -o app.exe .
```

### Build Frontend

```bash
cd frontend
npm run build
```

Cả hai đều build thành công ✅

## Next Steps (Optional)

1. Implement refresh token flow để auto-renew access tokens
2. Implement send email feature
3. Add email attachments support
4. Add search/filter emails
5. Implement email compose with rich text editor
6. Add email pagination
7. Publish app để không cần test users

## Troubleshooting

Xem chi tiết trong `GMAIL_API_INTEGRATION.md`
