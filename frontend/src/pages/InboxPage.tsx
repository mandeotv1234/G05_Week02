import { useState, useEffect } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useAppDispatch, useAppSelector } from "@/store/hooks";
import { logout } from "@/store/authSlice";
import { authService } from "@/services/auth.service";
import { emailService } from "@/services/email.service";
import { getAccessToken } from "@/lib/api-client";
import type { Email } from "@/types/email";
import MailboxList from "@/components/inbox/MailboxList";
import EmailList from "@/components/inbox/EmailList";
import EmailDetail from "@/components/inbox/EmailDetail";
import ComposeEmail from "@/components/inbox/ComposeEmail";
import { useQueryClient, useMutation } from "@tanstack/react-query";

export default function InboxPage() {
  const navigate = useNavigate();
  const dispatch = useAppDispatch();
  const queryClient = useQueryClient();
  const user = useAppSelector((state) => state.auth.user);
  const { mailbox, emailId } = useParams<{
    mailbox?: string;
    emailId?: string;
  }>();

  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  const [mobileView, setMobileView] = useState<"list" | "detail">("list");
  const [isComposeOpen, setIsComposeOpen] = useState(false);
  const [composeInitialData, setComposeInitialData] = useState({
    to: [] as string[],
    cc: [] as string[],
    subject: "",
    body: "",
  });
  const [theme, setTheme] = useState<"light" | "dark">(() => {
    if (typeof window !== "undefined") {
      return (localStorage.getItem("theme") as "light" | "dark") || "light";
    }
    return "light";
  });

  useEffect(() => {
    const savedTheme = localStorage.getItem("theme");
    const systemTheme = window.matchMedia("(prefers-color-scheme: dark)")
      .matches
      ? "dark"
      : "light";
    const initialTheme = (savedTheme as "light" | "dark") || systemTheme;

    setTheme(initialTheme);
    if (initialTheme === "dark") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  }, []);

  const toggleTheme = () => {
    const newTheme = theme === "light" ? "dark" : "light";
    setTheme(newTheme);
    localStorage.setItem("theme", newTheme);
    if (newTheme === "dark") {
      document.documentElement.classList.add("dark");
    } else {
      document.documentElement.classList.remove("dark");
    }
  };

  // Use URL params or default to 'inbox'
  const selectedMailboxId = mailbox || "inbox";
  const selectedEmailId = emailId || null;

  useEffect(() => {
    if (user) {
      // Start watching for email updates
      emailService.watchMailbox().catch(console.error);

      // Connect to SSE
      const token = getAccessToken();
      const eventSource = new EventSource(
        `${
          import.meta.env.VITE_API_URL || "http://localhost:8080"
        }/api/events?token=${token}`,
        {
          withCredentials: true,
        }
      );

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === "email_update") {
            console.log("Received email update:", data.payload);
            // Invalidate queries to refresh lists
            queryClient.invalidateQueries({ queryKey: ["emails"] });
            queryClient.invalidateQueries({ queryKey: ["mailboxes"] });
          }
        } catch (error) {
          console.error("Error parsing SSE message:", error);
        }
      };

      eventSource.onerror = (error) => {
        console.error("SSE error:", error);
        eventSource.close();
      };

      return () => {
        eventSource.close();
      };
    }
  }, [user, queryClient]);

  const logoutMutation = useMutation({
    mutationFn: authService.logout,
    onSuccess: () => {
      dispatch(logout());
      queryClient.clear();
      navigate("/login");
    },
  });

  const handleLogout = () => {
    logoutMutation.mutate();
  };

  const handleSelectMailbox = (id: string) => {
    navigate(`/${id}`);
    setMobileView("list");
    setIsMobileMenuOpen(false);
  };

  const handleSelectEmail = (email: Email) => {
    navigate(`/${selectedMailboxId}/${email.id}`);
    setMobileView("detail");
  };

  const handleToggleStar = () => {
    queryClient.invalidateQueries({ queryKey: ["emails"] });
  };

  const handleForward = (email: Email) => {
    setComposeInitialData({
      to: [],
      cc: [],
      subject: `Fwd: ${email.subject}`,
      body: `<br><br>---------- Forwarded message ---------<br>From: ${
        email.from
      }<br>Date: ${new Date(email.received_at).toLocaleString()}<br>Subject: ${
        email.subject
      }<br>To: ${email.to.join(", ")}<br><br>${
        email.body || email.preview || ""
      }`,
    });
    setIsComposeOpen(true);
  };

  const handleReply = (email: Email) => {
    const date = new Date(email.received_at);
    const weekday = date.toLocaleDateString("vi-VN", { weekday: "short" });
    const day = date.getDate();
    const month = date.getMonth() + 1;
    const year = date.getFullYear();
    const time = date.toLocaleTimeString("vi-VN", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    });

    const dateStr = `Vào ${weekday}, ${day} thg ${month}, ${year} vào lúc ${time}`;

    let senderName = email.from;
    let senderEmail = email.from;
    const match = email.from.match(/^(.*?)\s*<(.*)>$/);
    if (match) {
      senderName = match[1].replace(/"/g, "").trim();
      senderEmail = match[2].trim();
    } else {
      senderName = email.from.replace(/"/g, "").trim();
      if (senderName.includes("@")) {
        senderEmail = senderName;
      }
    }

    const senderHtml = `${senderName} &lt;<a href="mailto:${senderEmail}">${senderEmail}</a>&gt;`;

    setComposeInitialData({
      to: [senderEmail],
      cc: [],
      subject: `Re: ${email.subject}`,
      body: `<div dir="ltr"><br></div><br><div class="gmail_quote gmail_quote_container"><div dir="ltr" class="gmail_attr">${dateStr} ${senderHtml} đã viết:<br></div><blockquote class="gmail_quote" style="margin:0px 0px 0px 0.8ex;border-left:1px solid rgb(204,204,204);padding-left:1ex">${
        email.body || email.preview || ""
      }</blockquote></div>`,
    });
    setIsComposeOpen(true);
  };

  const handleReplyAll = (email: Email) => {
    const date = new Date(email.received_at);
    const weekday = date.toLocaleDateString("vi-VN", { weekday: "short" });
    const day = date.getDate();
    const month = date.getMonth() + 1;
    const year = date.getFullYear();
    const time = date.toLocaleTimeString("vi-VN", {
      hour: "2-digit",
      minute: "2-digit",
      hour12: false,
    });

    const dateStr = `Vào ${weekday}, ${day} thg ${month}, ${year} vào lúc ${time}`;

    let senderName = email.from;
    let senderEmail = email.from;
    const match = email.from.match(/^(.*?)\s*<(.*)>$/);
    if (match) {
      senderName = match[1].replace(/"/g, "").trim();
      senderEmail = match[2].trim();
    } else {
      senderName = email.from.replace(/"/g, "").trim();
      if (senderName.includes("@")) {
        senderEmail = senderName;
      }
    }

    const senderHtml = `${senderName} &lt;<a href="mailto:${senderEmail}">${senderEmail}</a>&gt;`;

    // Calculate CC list
    // CC = (Original To + Original CC) - (Me + Sender)
    const myEmail = user?.email || "";
    const allRecipients = [...(email.to || []), ...(email.cc || [])];

    const ccList = allRecipients
      .map((r) => {
        const match = r.match(/^(.*?)\s*<(.*)>$/);
        return match ? match[2].trim() : r.trim();
      })
      .filter(
        (email) =>
          email.toLowerCase() !== myEmail.toLowerCase() &&
          email.toLowerCase() !== senderEmail.toLowerCase()
      );

    // Remove duplicates
    const uniqueCcList = [...new Set(ccList)];

    setComposeInitialData({
      to: [senderEmail],
      cc: uniqueCcList,
      subject: `Re: ${email.subject}`,
      body: `<div dir="ltr"><br></div><br><div class="gmail_quote gmail_quote_container"><div dir="ltr" class="gmail_attr">${dateStr} ${senderHtml} đã viết:<br></div><blockquote class="gmail_quote" style="margin:0px 0px 0px 0.8ex;border-left:1px solid rgb(204,204,204);padding-left:1ex">${
        email.body || email.preview || ""
      }</blockquote></div>`,
    });
    setIsComposeOpen(true);
  };

  return (
    <div className="flex h-screen bg-gray-50 dark:bg-[#111418] text-gray-900 dark:text-white overflow-hidden font-sans transition-colors duration-200">
      {/* Desktop: 3-column layout */}
      <div className="hidden md:flex w-full h-full">
        {/* Column 1: Sidebar */}
        <div className="w-[220px] shrink-0 border-r border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-[#111418]">
          <MailboxList
            selectedMailboxId={selectedMailboxId}
            onSelectMailbox={handleSelectMailbox}
            onComposeClick={() => setIsComposeOpen(true)}
            onLogout={handleLogout}
            theme={theme}
            onToggleTheme={toggleTheme}
          />
        </div>

        {/* Column 2: Email List */}
        <div className="w-[360px] shrink-0 border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-[#111418] flex flex-col">
          <EmailList
            mailboxId={selectedMailboxId}
            selectedEmailId={selectedEmailId}
            onSelectEmail={handleSelectEmail}
            onToggleStar={handleToggleStar}
          />
        </div>

        {/* Column 3: Email Detail */}
        <div className="flex-1 bg-white dark:bg-[#111418] min-w-0">
          <EmailDetail
            emailId={selectedEmailId}
            onToggleStar={handleToggleStar}
            onReply={handleReply}
            onReplyAll={handleReplyAll}
            onForward={handleForward}
            theme={theme}
          />
        </div>
      </div>

      {/* Mobile Layout */}
      <div className="md:hidden flex-1 flex flex-col w-full h-full relative">
        {/* Mobile Header */}
        <div className="h-14 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between px-4 bg-white dark:bg-[#111418] shrink-0">
          <button
            onClick={() => setIsMobileMenuOpen(true)}
            className="p-2 -ml-2 rounded-full hover:bg-gray-100 dark:hover:bg-white/10"
          >
            <span className="material-symbols-outlined">menu</span>
          </button>
          <span className="font-bold text-lg">MailApp</span>
          <div className="w-10"></div> {/* Spacer */}
        </div>

        {/* Mobile Menu Overlay */}
        {isMobileMenuOpen && (
          <div className="absolute inset-0 z-50 bg-white dark:bg-[#111418] flex flex-col">
            <div className="h-14 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between px-4 shrink-0">
              <span className="font-bold text-lg">Menu</span>
              <button
                onClick={() => setIsMobileMenuOpen(false)}
                className="p-2 -mr-2 rounded-full hover:bg-gray-100 dark:hover:bg-white/10"
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>
            <div className="flex-1 overflow-y-auto">
              <MailboxList
                selectedMailboxId={selectedMailboxId}
                onSelectMailbox={handleSelectMailbox}
                onComposeClick={() => {
                  setIsComposeOpen(true);
                  setIsMobileMenuOpen(false);
                }}
                onLogout={handleLogout}
                theme={theme}
                onToggleTheme={toggleTheme}
              />
            </div>
          </div>
        )}

        {/* Mobile Content */}
        <div className="flex-1 overflow-hidden relative">
          {mobileView === "list" ? (
            <EmailList
              mailboxId={selectedMailboxId}
              selectedEmailId={selectedEmailId}
              onSelectEmail={handleSelectEmail}
              onToggleStar={handleToggleStar}
            />
          ) : (
            <div className="absolute inset-0 flex flex-col bg-white dark:bg-[#111418]">
              <div className="h-12 border-b border-gray-200 dark:border-gray-800 flex items-center px-2 shrink-0">
                <button
                  onClick={() => {
                    navigate(`/${selectedMailboxId}`);
                    setMobileView("list");
                  }}
                  className="flex items-center gap-1 px-2 py-1 rounded-lg hover:bg-gray-100 dark:hover:bg-white/10 text-gray-600 dark:text-gray-300"
                >
                  <span className="material-symbols-outlined">arrow_back</span>
                  <span>Back</span>
                </button>
              </div>
              <div className="flex-1 overflow-hidden">
                <EmailDetail
                  emailId={selectedEmailId}
                  onToggleStar={handleToggleStar}
                  onReply={handleReply}
                  onReplyAll={handleReplyAll}
                  onForward={handleForward}
                  theme={theme}
                />
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Compose Email Dialog */}
      <ComposeEmail
        open={isComposeOpen}
        onOpenChange={(open) => {
          setIsComposeOpen(open);
          if (!open)
            setComposeInitialData({ to: [], cc: [], subject: "", body: "" });
        }}
        initialTo={composeInitialData.to}
        initialCc={composeInitialData.cc}
        initialSubject={composeInitialData.subject}
        initialBody={composeInitialData.body}
      />
    </div>
  );
}
