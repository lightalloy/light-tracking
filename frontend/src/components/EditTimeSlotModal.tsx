import { useState, useEffect } from 'react';
import { UpdateTimeSlot } from '../../wailsjs/go/app/App';

interface TimeSlot {
  id: number;
  task_name: string;
  start_time: string;
  end_time?: string;
  duration_seconds: number;
}

interface EditTimeSlotModalProps {
  slot: TimeSlot | null;
  onClose: () => void;
  onSave: () => void;
}

function EditTimeSlotModal({ slot, onClose, onSave }: EditTimeSlotModalProps) {
  const [taskName, setTaskName] = useState('');
  const [startTime, setStartTime] = useState('');
  const [endTime, setEndTime] = useState('');
  const [hasEndTime, setHasEndTime] = useState(false);

  useEffect(() => {
    if (slot) {
      setTaskName(slot.task_name);
      
      // Format datetime for input (YYYY-MM-DDTHH:mm)
      const start = new Date(slot.start_time);
      const startStr = start.toISOString().slice(0, 16);
      setStartTime(startStr);
      
      if (slot.end_time) {
        const end = new Date(slot.end_time);
        const endStr = end.toISOString().slice(0, 16);
        setEndTime(endStr);
        setHasEndTime(true);
      } else {
        setEndTime('');
        setHasEndTime(false);
      }
    }
  }, [slot]);

  if (!slot) {
    return null;
  }

  const handleSave = async () => {
    if (!taskName.trim()) {
      alert('Task name cannot be empty');
      return;
    }

    try {
      const start = new Date(startTime);
      let end: Date | null = null;
      
      if (hasEndTime && endTime) {
        end = new Date(endTime);
        if (end <= start) {
          alert('End time must be after start time');
          return;
        }
      }

      await UpdateTimeSlot(
        slot.id,
        taskName.trim(),
        start.toISOString(),
        end ? end.toISOString() : ''
      );
      
      onSave();
      onClose();
    } catch (error) {
      console.error('Failed to update time slot:', error);
      alert('Failed to update time slot');
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <h3>Edit Time Slot</h3>
        
        <div className="form-group">
          <label htmlFor="task-name">Task Name:</label>
          <input
            id="task-name"
            type="text"
            value={taskName}
            onChange={(e) => setTaskName(e.target.value)}
            className="form-input"
          />
        </div>

        <div className="form-group">
          <label htmlFor="start-time">Start Time:</label>
          <input
            id="start-time"
            type="datetime-local"
            value={startTime}
            onChange={(e) => setStartTime(e.target.value)}
            className="form-input"
          />
        </div>

        <div className="form-group">
          <label>
            <input
              type="checkbox"
              checked={hasEndTime}
              onChange={(e) => setHasEndTime(e.target.checked)}
            />
            Has End Time
          </label>
        </div>

        {hasEndTime && (
          <div className="form-group">
            <label htmlFor="end-time">End Time:</label>
            <input
              id="end-time"
              type="datetime-local"
              value={endTime}
              onChange={(e) => setEndTime(e.target.value)}
              className="form-input"
            />
          </div>
        )}

        <div className="modal-actions">
          <button className="btn btn-cancel" onClick={onClose}>
            Cancel
          </button>
          <button className="btn btn-save" onClick={handleSave}>
            Save
          </button>
        </div>
      </div>
    </div>
  );
}

export default EditTimeSlotModal;

