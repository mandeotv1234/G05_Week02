import { createBrowserRouter, Navigate } from "react-router-dom";
import { PrivateRoute } from "./PrivateRoute";
import LoginPage from "@/pages/LoginPage";
import SignUpPage from "@/pages/SignUpPage";
import InboxPage from "@/pages/InboxPage";
import KanbanPage from "@/pages/KanbanPage";
import SetPasswordPage from "@/pages/SetPasswordPage";

export const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/signup",
    element: <SignUpPage />,
  },
  {
    path: "/set-password",
    element: (
      <PrivateRoute>
        <SetPasswordPage />
      </PrivateRoute>
    ),
  },
  {
    path: "/kanban",
    element: (
      <PrivateRoute>
        <KanbanPage />
      </PrivateRoute>
    ),
  },
  {
    path: "/",
    element: (
      <PrivateRoute>
        <InboxPage />
      </PrivateRoute>
    ),
  },
  {
    path: "/:mailbox",
    element: (
      <PrivateRoute>
        <InboxPage />
      </PrivateRoute>
    ),
  },
  {
    path: "/:mailbox/:emailId",
    element: (
      <PrivateRoute>
        <InboxPage />
      </PrivateRoute>
    ),
  },
  {
    path: "*",
    element: <Navigate to="/inbox" replace />,
  },
]);
