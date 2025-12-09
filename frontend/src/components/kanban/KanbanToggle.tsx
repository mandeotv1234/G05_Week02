
export type KanbanToggleProps = {
  isKanban: boolean;
  onToggle: () => void;
};

export default function KanbanToggle({ isKanban, onToggle }: KanbanToggleProps) {
  return (
    <button
      className="px-3 py-1 rounded bg-blue-500 text-white font-semibold hover:bg-blue-600 transition"
      onClick={onToggle}
    >
      {isKanban ? "Chuyển về chế độ truyền thống" : "Chuyển sang Kanban"}
    </button>
  );
}
