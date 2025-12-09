import { useRef } from "react";

export function useKanbanDrag(onDrop: (emailId: string, targetColumnId: string) => void) {
  const dragEmailId = useRef<string | null>(null);

  function handleDragStart(emailId: string) {
    dragEmailId.current = emailId;
  }

  function handleDragEnd() {
    dragEmailId.current = null;
  }

  function handleDrop(targetColumnId: string) {
    if (dragEmailId.current) {
      onDrop(dragEmailId.current, targetColumnId);
      dragEmailId.current = null;
    }
  }

  return {
    handleDragStart,
    handleDragEnd,
    handleDrop,
  };
}
