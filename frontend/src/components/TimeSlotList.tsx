interface TimeSlot {
  id: number;
  task_name: string;
  start_time: string;
  end_time?: string;
  duration_seconds: number;
}

interface TimeSlotListProps {
  slots: TimeSlot[];
  onEdit?: (slot: TimeSlot) => void;
  onDelete?: (id: number) => void;
}

function TimeSlotList({ slots, onEdit, onDelete }: TimeSlotListProps) {
  const formatTime = (timeString: string): string => {
    const date = new Date(timeString);
    return date.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit' });
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  if (slots.length === 0) {
    return <div className="empty-list">No time slots for this day</div>;
  }

  return (
    <div className="time-slot-list">
      <h3>Time Slots</h3>
      <ul className="slot-list">
        {slots.map((slot) => (
          <li key={slot.id} className="slot-item">
            <div className="slot-info">
              <div className="slot-task">{slot.task_name}</div>
              <div className="slot-time">
                {formatTime(slot.start_time)}
                {slot.end_time && ` - ${formatTime(slot.end_time)}`}
              </div>
              <div className="slot-duration">{formatDuration(slot.duration_seconds)}</div>
            </div>
            {(onEdit || onDelete) && (
              <div className="slot-actions">
                {onEdit && (
                  <button className="btn-edit" onClick={() => onEdit(slot)}>
                    Edit
                  </button>
                )}
                {onDelete && (
                  <button className="btn-delete" onClick={() => onDelete(slot.id)}>
                    Delete
                  </button>
                )}
              </div>
            )}
          </li>
        ))}
      </ul>
    </div>
  );
}

export default TimeSlotList;

