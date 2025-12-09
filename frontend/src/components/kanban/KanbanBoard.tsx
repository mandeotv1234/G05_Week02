import React from "react";
import { useKanbanDrag } from "./useKanbanDrag";
import type { Email } from "@/types/email";

export type KanbanColumn = {
  id: string;
  title: string;
  emails: Email[];
  offset?: number;
  limit?: number;
  total?: number;
};

export type KanbanBoardProps = {
  columns: KanbanColumn[];
  onEmailDrop: (emailId: string, targetColumnId: string) => void;
  renderCardActions?: (email: Email) => React.ReactNode;
  onPageChange?: (colId: string, dir: 1 | -1) => void;
  onEmailClick?: (emailId: string) => void;
};

export default function KanbanBoard({ columns, onEmailDrop, renderCardActions, onPageChange, onEmailClick }: KanbanBoardProps) {
  const drag = useKanbanDrag(onEmailDrop);

  return (
    <div className="flex gap-4 w-full h-full overflow-x-auto">
      {columns.map((col) => (
        <div
          key={col.id}
          className="flex-1 min-w-[280px] bg-gray-100 dark:bg-gray-900 rounded-lg p-2 shadow"
          onDragOver={(e) => e.preventDefault()}
          onDrop={() => drag.handleDrop(col.id)}
        >
          <div className="font-bold text-lg mb-2 text-center">{col.title}</div>
          {/* Pagination controls */}
          <div className="flex justify-between items-center mb-2">
            <button
              className="px-2 py-1 rounded bg-gray-300 text-xs"
              disabled={col.offset === 0}
              onClick={() => onPageChange?.(col.id, -1)}
            >
              Prev
            </button>
            <span className="text-xs">Page {col.offset !== undefined && col.limit ? Math.floor(col.offset / col.limit) + 1 : 1}</span>
            <button
              className="px-2 py-1 rounded bg-gray-300 text-xs"
              disabled={col.emails.length < (col.limit || 20)}
              onClick={() => onPageChange?.(col.id, 1)}
            >
              Next
            </button>
          </div>
          <div className="flex flex-col gap-2">
            {col.emails.map((email) => (
              <div
                key={email.id}
                className="bg-white dark:bg-gray-800 rounded shadow p-3 cursor-pointer"
                draggable
                onDragStart={() => drag.handleDragStart(email.id)}
                onDragEnd={drag.handleDragEnd}
                onClick={() => onEmailClick?.(email.id)}
              >
                <div className="font-semibold text-sm">{email.from}</div>
                <div className="text-xs text-gray-500">{email.subject}</div>
                <div className="text-xs mt-1 line-clamp-2">{email.preview || email.body?.slice(0, 100)}</div>
                <div className="mt-2 flex gap-2">
                  {renderCardActions?.(email)}
                </div>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
