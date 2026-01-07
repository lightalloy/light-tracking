import { useState } from 'react';
import Timer from './components/Timer';
import Statistics from './components/Statistics';
import EditTimeSlotModal from './components/EditTimeSlotModal';
import { DeleteTimeSlot } from '../wailsjs/go/app/App';
import './App.css';

interface TimeSlot {
  id: number;
  task_name: string;
  start_time: string;
  end_time?: string;
  duration_seconds: number;
}

function App() {
  const [activeTab, setActiveTab] = useState<'timer' | 'statistics'>('timer');
  const [editingSlot, setEditingSlot] = useState<TimeSlot | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);

  const handleEdit = (slot: TimeSlot) => {
    setEditingSlot(slot);
  };

  const handleDelete = async (id: number) => {
    if (confirm('Are you sure you want to delete this time slot?')) {
      try {
        await DeleteTimeSlot(id);
        // Trigger refresh of statistics
        setRefreshKey(prev => prev + 1);
      } catch (error) {
        console.error('Failed to delete time slot:', error);
        alert('Failed to delete time slot');
      }
    }
  };

  const handleSave = () => {
    // Trigger refresh of statistics
    setRefreshKey(prev => prev + 1);
  };

  return (
    <div id="App">
      <div className="header">
        <h1>Light Tracking</h1>
        <nav className="tabs">
          <button
            className={activeTab === 'timer' ? 'tab active' : 'tab'}
            onClick={() => setActiveTab('timer')}
          >
            Timer
          </button>
          <button
            className={activeTab === 'statistics' ? 'tab active' : 'tab'}
            onClick={() => setActiveTab('statistics')}
          >
            Statistics
          </button>
        </nav>
      </div>

      <div className="content">
        {activeTab === 'timer' && <Timer />}
        {activeTab === 'statistics' && (
          <Statistics key={refreshKey} onEdit={handleEdit} onDelete={handleDelete} />
        )}
      </div>

      {editingSlot && (
        <EditTimeSlotModal
          slot={editingSlot}
          onClose={() => setEditingSlot(null)}
          onSave={handleSave}
        />
      )}
    </div>
  );
}

export default App;
