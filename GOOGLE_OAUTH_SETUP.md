# Hướng dẫn cấu hình Google OAuth với Gmail API

## Lỗi "no registered origin" và "invalid_client"

Lỗi này xảy ra khi Google Cloud Console chưa được cấu hình đúng. Làm theo các bước sau:

## Bước 1: Vào Google Cloud Console

1. Truy cập: https://console.cloud.google.com/
2. Chọn project của bạn (hoặc tạo project mới)

## Bước 2: Bật Gmail API

1. Vào **APIs & Services** > **Library**
2. Tìm "Gmail API"
3. Click **Enable**
4. Cũng nên enable "People API" và "Google+ API" (nếu có)

## Bước 3: Tạo OAuth 2.0 Credentials

1. Vào **APIs & Services** > **Credentials**
2. Click **Create Credentials** > **OAuth client ID**
3. Nếu chưa có OAuth consent screen, bạn sẽ được yêu cầu cấu hình:
   - Chọn **External** (hoặc Internal nếu dùng Google Workspace)
   - Điền App name, User support email
   - **QUAN TRỌNG**: Thêm Gmail scopes vào OAuth consent screen:
     - `https://www.googleapis.com/auth/gmail.readonly` (đọc emails)
     - `https://www.googleapis.com/auth/gmail.modify` (đánh dấu đã đọc, starred, etc.)
   - Thêm email của bạn vào Test users (nếu ở chế độ Testing)
   - Click **Save and Continue** qua các bước

## Bước 4: Cấu hình OAuth Client ID

1. **Application type**: Chọn **Web application**
2. **Name**: Đặt tên (ví dụ: "Email App Frontend")
3. **Authorized JavaScript origins**:
   - Thêm: `http://localhost:5173` (cho development)
   - Nếu deploy, thêm URL production (ví dụ: `https://your-app.netlify.app`)
4. **Authorized redirect URIs**:
   - Với @react-oauth/google, bạn có thể để trống hoặc thêm:
   - `http://localhost:5173` (cho development)
   - `https://your-app.netlify.app` (cho production)

## Bước 5: Copy Client ID và Client Secret

1. Sau khi tạo xong, copy **Client ID** và **Client Secret**
2. Dán vào file `.env`:

   **Backend (.env trong thư mục backend):**

   ```
   GOOGLE_CLIENT_ID=paste-your-client-id-here
   GOOGLE_CLIENT_SECRET=paste-your-client-secret-here
   GOOGLE_REDIRECT_URI=http://localhost:8080/api/auth/google/callback
   ```

   **Frontend (.env trong thư mục frontend):**

   ```
   VITE_GOOGLE_CLIENT_ID=paste-your-client-id-here
   ```

## Bước 6: Thêm Test Users (nếu app ở chế độ Testing)

1. Vào **APIs & Services** > **OAuth consent screen**
2. Scroll xuống phần **Test users**
3. Click **Add Users**
4. Thêm email Google mà bạn sẽ dùng để test

## Bước 7: Restart dev servers

**Backend:**

```bash
cd backend
go run main.go
```

**Frontend:**

```bash
cd frontend
npm run dev
```

## Lưu ý quan trọng:

- **Authorized JavaScript origins** PHẢI khớp chính xác với URL bạn đang chạy
- Nếu chạy trên `http://localhost:5173`, phải thêm chính xác `http://localhost:5173` (không có dấu `/` ở cuối)
- Nếu chạy trên port khác, phải thêm đúng port đó
- Sau khi thay đổi trong Google Cloud Console, có thể mất vài phút để có hiệu lực
- **QUAN TRỌNG**: App phải ở chế độ Testing và bạn phải thêm email của mình vào Test Users để truy cập Gmail

## Kiểm tra lại:

1. Đảm bảo file `.env` có `VITE_GOOGLE_CLIENT_ID` và `GOOGLE_CLIENT_SECRET` đúng
2. Đảm bảo đã restart cả backend và frontend server sau khi thay đổi `.env`
3. Đảm bảo Authorized JavaScript origins trong Google Console khớp với URL bạn đang dùng
4. Đảm bảo đã enable Gmail API trong Google Cloud Console
5. Đảm bảo đã thêm Gmail scopes vào OAuth consent screen
6. Clear cache trình duyệt nếu vẫn lỗi

## Cách sử dụng:

1. Đăng nhập bằng Google (với email đã thêm vào Test Users)
2. Chấp nhận quyền truy cập Gmail khi Google hỏi
3. Sau khi đăng nhập, app sẽ tự động lấy emails từ Gmail thật của bạn
4. Mailboxes sẽ hiển thị các label từ Gmail (INBOX, SENT, DRAFT, v.v.)
5. Emails sẽ được fetch trực tiếp từ Gmail API
