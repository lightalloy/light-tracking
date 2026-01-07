import { useState, useEffect } from 'react';
import { StartTimer, StopTimer, GetActiveTimeSlot, IsTimerRunning, GetElapsedTime } from '../../wailsjs/go/app/App';
import TaskInput from './TaskInput';

function Timer() {
  const [taskName, setTaskName] = useState('');
  const [isRunning, setIsRunning] = useState(false);
  const [elapsedSeconds, setElapsedSeconds] = useState(0);
  const [currentTask, setCurrentTask] = useState<string>('');

  useEffect(() => {
    // Check if timer is already running on mount
    checkTimerStatus();

    // Update elapsed time every second
    const interval = setInterval(() => {
      if (isRunning) {
        updateElapsedTime();
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [isRunning]);

  const checkTimerStatus = async () => {
    try {
      const running = await IsTimerRunning();
      setIsRunning(running);
      
      if (running) {
        const activeSlot = await GetActiveTimeSlot();
        if (activeSlot) {
          setCurrentTask(activeSlot.task_name);
          updateElapsedTime();
        }
      }
    } catch (error) {
      console.error('Failed to check timer status:', error);
    }
  };

  const updateElapsedTime = async () => {
    try {
      const seconds = await GetElapsedTime();
      setElapsedSeconds(seconds);
    } catch (error) {
      console.error('Failed to get elapsed time:', error);
    }
  };

  const handleStart = async () => {
    if (!taskName.trim()) {
      alert('Please enter a task name');
      return;
    }

    try {
      await StartTimer(taskName.trim());
      setCurrentTask(taskName.trim());
      setIsRunning(true);
      setElapsedSeconds(0);
      setTaskName('');
    } catch (error) {
      console.error('Failed to start timer:', error);
      alert('Failed to start timer');
    }
  };

  const handleStop = async () => {
    try {
      await StopTimer();
      setIsRunning(false);
      setElapsedSeconds(0);
      setCurrentTask('');
    } catch (error) {
      console.error('Failed to stop timer:', error);
      alert('Failed to stop timer');
    }
  };

  const formatTime = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (hours > 0) {
      return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    }
    return `${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <div className="timer-container">
      <h2>Time Tracker</h2>
      
      {isRunning && currentTask && (
        <div className="current-task">
          <p>Current task: <strong>{currentTask}</strong></p>
          <div className="elapsed-time">
            {formatTime(elapsedSeconds)}
          </div>
        </div>
      )}

      <TaskInput
        taskName={taskName}
        onTaskNameChange={setTaskName}
        onStart={handleStart}
        onStop={handleStop}
        isRunning={isRunning}
      />
    </div>
  );
}

export default Timer;

