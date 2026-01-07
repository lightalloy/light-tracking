import { useState, useEffect } from 'react';
import { GetTimeSlotsByDate, GetTaskStatistics } from '../../wailsjs/go/app/App';
import TimeSlotList from './TimeSlotList';

interface TimeSlot {
  id: number;
  task_name: string;
  start_time: string;
  end_time?: string;
  duration_seconds: number;
}

interface StatisticsProps {
  onEdit?: (slot: TimeSlot) => void;
  onDelete?: (id: number) => void;
  key?: number;
}

function Statistics({ onEdit, onDelete }: StatisticsProps) {
  const [selectedDate, setSelectedDate] = useState(new Date().toISOString().split('T')[0]);
  const [slots, setSlots] = useState<TimeSlot[]>([]);
  const [taskStats, setTaskStats] = useState<Record<string, number>>({});
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadStatistics();
  }, [selectedDate]);

  const loadStatistics = async () => {
    setLoading(true);
    try {
      const [timeSlots, stats] = await Promise.all([
        GetTimeSlotsByDate(selectedDate),
        GetTaskStatistics(selectedDate),
      ]);
      setSlots(timeSlots || []);
      setTaskStats(stats || {});
    } catch (error) {
      console.error('Failed to load statistics:', error);
    } finally {
      setLoading(false);
    }
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m`;
  };

  return (
    <div className="statistics-container">
      <h2>Statistics</h2>
      
      <div className="date-selector">
        <label htmlFor="date-input">Select date: </label>
        <input
          id="date-input"
          type="date"
          value={selectedDate}
          onChange={(e) => setSelectedDate(e.target.value)}
        />
      </div>

      {loading ? (
        <div>Loading...</div>
      ) : (
        <>
          <div className="task-statistics">
            <h3>Time by Task</h3>
            {Object.keys(taskStats).length === 0 ? (
              <p>No statistics for this day</p>
            ) : (
              <ul className="stats-list">
                {Object.entries(taskStats)
                  .sort((a, b) => b[1] - a[1])
                  .map(([task, seconds]) => (
                    <li key={task} className="stat-item">
                      <span className="stat-task">{task}</span>
                      <span className="stat-duration">{formatDuration(seconds)}</span>
                    </li>
                  ))}
              </ul>
            )}
          </div>

          <TimeSlotList slots={slots} onEdit={onEdit} onDelete={onDelete} />
        </>
      )}
    </div>
  );
}

export default Statistics;

