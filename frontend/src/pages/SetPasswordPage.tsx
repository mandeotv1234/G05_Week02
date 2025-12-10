import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useMutation } from "@tanstack/react-query";
import { authService } from "@/services/auth.service";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import { toast } from "sonner";
import { Mail } from "lucide-react";

const setPasswordSchema = z
  .object({
    password: z.string().min(6, "Password must be at least 6 characters"),
    confirmPassword: z
      .string()
      .min(6, "Password must be at least 6 characters"),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: "Passwords do not match",
    path: ["confirmPassword"],
  });

type SetPasswordFormData = z.infer<typeof setPasswordSchema>;

export default function SetPasswordPage() {
  const navigate = useNavigate();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<SetPasswordFormData>({
    resolver: zodResolver(setPasswordSchema),
  });

  const setPasswordMutation = useMutation({
    mutationFn: (data: SetPasswordFormData) =>
      authService.setPassword(data.password),
    onSuccess: () => {
      toast.success("Mật khẩu đã được tạo thành công!");
      navigate("/inbox");
    },
    onError: (err: unknown) => {
      console.error(err);
      toast.error("Không thể tạo mật khẩu. Vui lòng thử lại.");
    },
  });

  const onSubmit = (data: SetPasswordFormData) => {
    setPasswordMutation.mutate(data);
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-gray-900 via-gray-800 to-gray-900 px-4 py-12">
      <div className="w-full max-w-md">
        {/* Logo */}
        <div className="flex items-center justify-center gap-2 mb-8">
          <div className="relative">
            <Mail className="h-10 w-10 text-blue-400" />
            <div className="absolute -bottom-1 -right-1 h-5 w-5 bg-green-500 rounded-full flex items-center justify-center shadow-lg">
              <div className="h-2.5 w-2.5 bg-white rounded-full"></div>
            </div>
          </div>
          <h1 className="text-3xl font-bold text-white">InboxFlow</h1>
        </div>

        {/* Title */}
        <div className="text-center mb-8">
          <h2 className="text-2xl font-bold text-white mb-2">
            Tạo mật khẩu hệ thống
          </h2>
          <p className="text-gray-400 text-sm">
            Bảo mật tài khoản của bạn với mật khẩu mới
          </p>
        </div>

        <Card className="bg-gray-800/80 backdrop-blur-sm border-gray-700 shadow-2xl">
          <CardContent className="pt-6">
            <div className="bg-blue-500/10 border border-blue-500/30 rounded-lg p-4 mb-6">
              <p className="text-blue-300 text-sm text-center leading-relaxed">
                Vui lòng tạo mật khẩu để đăng nhập vào hệ thống lần sau.
                <br />
                Mật khẩu này sẽ giúp bạn truy cập nhanh hơn mà không cần nhập
                cấu hình IMAP.
              </p>
            </div>

            <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
              <div className="space-y-2">
                <Label htmlFor="password" className="text-gray-300 font-medium">
                  Mật khẩu mới
                </Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Nhập mật khẩu (tối thiểu 6 ký tự)"
                  {...register("password")}
                  className="bg-gray-900/50 border-gray-600 text-white placeholder:text-gray-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all"
                />
                {errors.password && (
                  <p className="text-sm text-red-400 flex items-center gap-1">
                    <span className="text-xs">⚠️</span>
                    {errors.password.message}
                  </p>
                )}
              </div>
              <div className="space-y-2">
                <Label
                  htmlFor="confirmPassword"
                  className="text-gray-300 font-medium"
                >
                  Xác nhận mật khẩu
                </Label>
                <Input
                  id="confirmPassword"
                  type="password"
                  placeholder="Nhập lại mật khẩu"
                  {...register("confirmPassword")}
                  className="bg-gray-900/50 border-gray-600 text-white placeholder:text-gray-500 focus:border-blue-500 focus:ring-2 focus:ring-blue-500/20 transition-all"
                />
                {errors.confirmPassword && (
                  <p className="text-sm text-red-400 flex items-center gap-1">
                    <span className="text-xs">⚠️</span>
                    {errors.confirmPassword.message}
                  </p>
                )}
              </div>
              <Button
                type="submit"
                className="w-full bg-gradient-to-r from-blue-600 to-blue-500 hover:from-blue-700 hover:to-blue-600 text-white font-semibold py-2.5 shadow-lg shadow-blue-500/30 transition-all duration-200 hover:shadow-blue-500/50"
                disabled={isSubmitting || setPasswordMutation.isPending}
              >
                {isSubmitting || setPasswordMutation.isPending ? (
                  <span className="flex items-center gap-2">
                    <svg className="animate-spin h-5 w-5" viewBox="0 0 24 24">
                      <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"
                        fill="none"
                      />
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                      />
                    </svg>
                    Đang xử lý...
                  </span>
                ) : (
                  "Tạo mật khẩu"
                )}
              </Button>
            </form>
          </CardContent>
        </Card>

        {/* Security Note */}
        <p className="text-center text-gray-500 text-xs mt-6">
          🔒 Mật khẩu của bạn được mã hóa và bảo mật an toàn
        </p>
      </div>
    </div>
  );
}
