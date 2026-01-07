import { KeyboardEvent } from 'react';

interface TaskInputProps {
  taskName: string;
  onTaskNameChange: (name: string) => void;
  onStart: () => void;
  onStop: () => void;
  isRunning: boolean;
}

function TaskInput({ taskName, onTaskNameChange, onStart, onStop, isRunning }: TaskInputProps) {
  const handleKeyPress = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !isRunning) {
      onStart();
    }
  };

  return (
    <div className="task-input-container">
      <input
        type="text"
        className="task-input"
        placeholder="Enter task name..."
        value={taskName}
        onChange={(e) => onTaskNameChange(e.target.value)}
        onKeyPress={handleKeyPress}
        disabled={isRunning}
      />
      <div className="button-group">
        <button
          className="btn btn-start"
          onClick={onStart}
          disabled={isRunning || !taskName.trim()}
        >
          Start
        </button>
        <button
          className="btn btn-stop"
          onClick={onStop}
          disabled={!isRunning}
        >
          Stop
        </button>
      </div>
    </div>
  );
}

export default TaskInput;

